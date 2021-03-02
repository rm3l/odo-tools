package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

// StripAnsi ...
func StripAnsi(str string) string {
	return re.ReplaceAllString(str, "")
}

// Match ...
type Match struct {
	FileType  string   `json:"filename"`
	Context   []string `json:"context,omitempty"`
	MoreLines int      `json:"moreLines,omitempty"`
}

// TestFailEntry ...
type TestFailEntry struct {
	PRList   []int
	TestFail int
	LastSeen *time.Time
	LogURLs  map[int] /* pr number -> log urls */ []string
}
type periodicJobData struct {
	failure        string
	clusterVersion string
	url            string
	flag           bool
}

// Result ...
type Result map[string]map[string][]Match

func main() {
	ex, errorListNonDuplicate := periodicjobstats()
	c := 0
	var failurelist string
	blobStorage, err := NewBlobStorage("./.cache")
	if err != nil {
		fmt.Println(err)
		return
	}

	testFailMap := map[string]TestFailEntry{}

	// store search results
	var result Result

	// jsonFile, err := os.Open("search.json")
	// if err != nil {
	// 	panic(err)
	// }
	// defer jsonFile.Close()

	req, err := http.NewRequest("GET", "https://search.svc.ci.openshift.org/search", nil)
	if err != nil {
		panic(err)
	}
	// https://search.svc.ci.openshift.org/search?search=%5C%5BFail%5C%5D&maxAge=336h&context=0&type=build-log&name=pull-ci-openshift-odo-master-&maxMatches=5&maxBytes=20971520
	q := req.URL.Query()
	q.Add("search", "\\[Fail\\]")
	q.Add("maxAge", "336h")
	q.Add("context", "0")
	q.Add("type", "build-log")
	q.Add("name", "pull-ci-openshift-odo-master-")
	q.Add("maxMatches", "5")
	q.Add("maxBytes", "20971520")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		panic(err)
	}

	// fmt.Println("map:", string(byteValue))

	// iterate over all results
	for k, search := range result {

		expectedBuildLogURL, err := parseURL(k)
		if err != nil {
			expectedBuildLogURL = ""
		}

		runTime, err := getTestJobRunTime(k, *blobStorage)
		if err != nil {
			fmt.Printf("Error occurred on test log download: %v ", err)
			return
		}

		odoIndex := strings.Index(k, "openshift_odo")

		prNumber := int64(-1)

		if odoIndex > -1 {
			str := k[odoIndex:]
			strArr := strings.Split(str, "/")

			prNumber, err = strconv.ParseInt(strArr[1], 10, 64)
		}

		//fmt.Printf("%s\n", file)
		// fmt.Println("map:", search)
		for _, matches := range search {
			// fmt.Printf("  %v\n", regexp)
			for _, match := range matches {
				lines := []string{}
				for _, line := range match.Context {
					// fmt.Printf("    %v\n", line)
					cleanLine := strings.TrimSpace(line)
					cleanLine = StripAnsi(cleanLine)
					// de-duplication
					// count each line only once
					dup := false
					for _, l := range lines {
						if l == cleanLine {
							dup = true
						}
					}
					if !dup {

						entry, exists := testFailMap[cleanLine]
						if !exists {
							entry = TestFailEntry{LogURLs: map[int][]string{}}
						}

						entry.TestFail++

						lines = append(lines, cleanLine)

						if runTime != nil {

							val := entry.LastSeen
							if val == nil {
								val = runTime
							}

							if runTime.After(*val) {
								val = runTime
							}
							entry.LastSeen = val

						}

						if prNumber >= 0 {

							matchFound := false
							for _, existingEntry := range entry.PRList {
								if int64(existingEntry) == prNumber {
									matchFound = true
								}
							}

							if !matchFound {
								entry.PRList = append(entry.PRList, int(prNumber))
							}

							// Add build log URL for the PR
							logURLList := entry.LogURLs[int(prNumber)]
							logURLList = append(logURLList, expectedBuildLogURL)
							entry.LogURLs[int(prNumber)] = logURLList

						}

						testFailMap[cleanLine] = entry

					}

				}
			}
		}

	}

	type TestFails struct {
		Score    int
		TestName string
		Fails    int
		LastSeen string
		PRList   []int
		Entry    TestFailEntry
	}

	// convert tests to slice so we can easily sort it
	fails := []TestFails{}
	for test, entry := range testFailMap {

		prList := entry.PRList

		sort.Sort(sort.Reverse(sort.IntSlice(prList)))

		lastSeenVal := ""

		// Score calculation
		score := 0
		{
			daysSinceLastSeen := 1

			lastSeenTime := entry.LastSeen
			if lastSeenTime != nil {

				days := time.Now().Sub(*lastSeenTime).Hours() / 24

				lastSeenVal = fmt.Sprintf("%d days ago", int(days))

				daysSinceLastSeen = int(days)
			}

			if daysSinceLastSeen == 0 {
				daysSinceLastSeen = 1
			}

			prListSize := len(prList)

			if prListSize > 6 {
				// >6 PRs does not imply any further strength than 6 PRs, for score calculation purposes.
				prListSize = 6
			}

			score = (10 * prListSize * entry.TestFail) / (daysSinceLastSeen)

			// fmt.Printf("%s %d %d\n", test, score, count)

			// Minimum score if there is at least one PR, and at least one fail, is 1
			if score == 0 && len(prList) > 0 && entry.TestFail > 0 {
				score = 1
			}
		}

		fails = append(fails, TestFails{TestName: test, Fails: entry.TestFail, PRList: prList, Score: score, LastSeen: lastSeenVal, Entry: entry})
	}

	sort.Slice(fails, func(i, j int) bool {
		one := fails[i].Score
		two := fails[j].Score

		// Primary sort: descending by score
		if one != two {
			return one > two
		}

		// Secondary sort: descending by fails
		one = fails[i].Fails
		two = fails[j].Fails
		if one != two {
			return one > two
		}

		// Tertiary sort: descring by pr list size
		one = len(fails[i].PRList)
		two = len(fails[j].PRList)
		if one != two {
			return one > two
		}

		// Finally, sort ascending by name
		return fails[j].TestName > fails[i].TestName
	})

	fmt.Println("# odo test statistics")
	fmt.Printf("Last update: %s (UTC)\n\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Println("Generated with https://github.com/jgwest/odo-tools/ and https://github.com/kadel/odo-tools")
	fmt.Println("## FLAKY TESTS: Failed test scenarios in past 14 days")
	fmt.Println("| Failure Score<sup>*</sup> | Failures on PR | Failures on PrediocJob | Test Name | Last Seen | PR List and Logs | PeriodicJob failure List")
	fmt.Println("|---|---|---|---|---|---|---|")
	for _, f := range fails {

		// Skip failures that appear to be contained to a single PR
		if len(f.PRList) <= 1 {
			continue
		}

		prListString := fmt.Sprintf("%d: ", len(f.PRList))
		for _, prNumber := range f.PRList {

			logURLs := f.Entry.LogURLs[prNumber]

			prListString += fmt.Sprintf("[#%d](%s/%d)", prNumber, "https://github.com/openshift/odo/pull", prNumber)

			if len(logURLs) > 0 {

				prListString += "<sup>"

				for index, logURL := range logURLs {
					prListString += "[" + strconv.FormatInt(int64(index+1), 10) + "](" + logURL + ")"

					if index+1 != len(logURLs) {
						prListString += ", "
					}
				}

				prListString += "</sup>"
			}

			prListString += " "
		}
		for _, k := range ex {
			if k.failure == f.TestName {
				failurelist = fmt.Sprintf("%s [%s](%s),", failurelist, k.clusterVersion, k.url)
				k.flag = true
				c = c + 1
			}

		}

		fmt.Printf("| %d | %d | %d | %s | %s | %s| %s\n", f.Score, f.Fails, c, f.TestName, f.LastSeen, prListString, failurelist)
		failurelist = ""
	}
	for _, k := range errorListNonDuplicate {
		c = 0
		for _, t := range ex {
			if k == t.failure && t.flag == false {
				t.flag = true
				c = c + 1
				failurelist = fmt.Sprintf("%s [%s](%s),", failurelist, t.clusterVersion, t.url)
			}
		}
		fmt.Printf("| %s | %s | %d | %s | %s | %s| %s\n", "-", "-", c, k, "-", "-", failurelist)
		failurelist = ""
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("<sup>*</sup> - Failure score is an arbitrary severity estimate, and is approximately `(# of PRs the test failure was seen in * # of test failures) / (days since failure)`. See code for full algorithm -- PRs welcome for algorithm improvements.")

}

func getTestJobRunTime(url string, blobStorage BlobStorage) (*time.Time, error) {
	urlContents, err := downloadTestLog(url, blobStorage)
	if err != nil {
		return nil, err
	}

	contentsByLine := strings.Split(strings.Replace(urlContents, "\r\n", "\n", -1), "\n")

	if len(contentsByLine) == 0 {
		return nil, nil
	}

	// Parse the first line in the file to determine when the test started (and failed.)
	{
		topLine := contentsByLine[0]

		tok1 := strings.Split(topLine, " ")
		if len(tok1) == 0 {
			return nil, nil
		}

		// There's definitely a better way to parse this :P
		tok2 := strings.Split(tok1[0], "/")
		if len(tok2) < 3 {
			return nil, nil
		}

		year, err := strconv.ParseInt(tok2[0], 10, 32)
		if err != nil {
			return nil, err
		}

		month, err := strconv.ParseInt(tok2[1], 10, 32)
		if err != nil {
			return nil, err
		}

		day, err := strconv.ParseInt(tok2[2], 10, 32)
		if err != nil {
			return nil, err
		}

		result := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Now().Location())

		return &result, nil

	}

}

func parseURL(url string) (string, error) {
	index := strings.LastIndex(url, "/")
	if index == -1 {
		return "", fmt.Errorf("Parsing error")
	}

	index = strings.LastIndex(url[0:index-1], "/")
	if index == -1 {
		return "", fmt.Errorf("Parsing error")
	}
	index = strings.LastIndex(url[0:index-1], "/")

	return "https://storage.googleapis.com/origin-ci-test/pr-logs/pull/openshift_odo" + url[index:] + "/build-log.txt", nil

}

func downloadTestLog(url string, blobStorage BlobStorage) (string, error) {

	value, err := blobStorage.retrieve(url)
	if err != nil {
		return "", err
	}

	if value != "" {
		return value, nil
	}

	// convert
	// https://prow.svc.ci.openshift.org/view/gcs/origin-ci-test/pr-logs/pull/batch/pull-ci-openshift-odo-master-v4.2-integration-e2e-benchmark/2047
	// to
	// https://storage.googleapis.com/origin-ci-test/pr-logs/pull/batch/pull-ci-openshift-odo-master-v4.2-integration-e2e-benchmark/2047/build-log.txt

	newURL, err := parseURL(url)

	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resp.Body.Close()

	err = blobStorage.store(url, string(byteValue))
	if err != nil {
		return "", err
	}

	return string(byteValue), nil

}

// BlobStorage ...
type BlobStorage struct {
	path string
}

// NewBlobStorage ...
func NewBlobStorage(pathParam string) (*BlobStorage, error) {
	blobStorage := BlobStorage{
		path: pathParam,
	}

	if _, err := os.Stat(pathParam); os.IsNotExist(err) {
		err = os.Mkdir(pathParam, 0755)
		if err != nil {
			return nil, err
		}
	}

	files, err := ioutil.ReadDir(pathParam)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		info, err := os.Stat(pathParam + "/" + f.Name())

		if err != nil {
			return nil, err
		}

		modTime := info.ModTime()

		diff := time.Now().Sub(modTime)

		// Delete cache entries older than 3 weeks
		if diff.Hours() > 24*7*3 {
			err := os.Remove(pathParam + "/" + f.Name())

			if err != nil {
				return nil, err
			}
		}

	}

	return &blobStorage, nil
}

func (s BlobStorage) store(key string, value string) error {
	base64Key := base64.StdEncoding.EncodeToString([]byte(key))

	expectedPath := s.path + "/" + base64Key

	err := ioutil.WriteFile(expectedPath, []byte(value), 0755)

	return err

}

func (s BlobStorage) retrieve(key string) (string, error) {
	base64Key := base64.StdEncoding.EncodeToString([]byte(key))

	expectedPath := s.path + "/" + base64Key

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		return "", nil
	}

	contents, err := ioutil.ReadFile(expectedPath)
	if err != nil {
		return "", err
	}

	return string(contents), nil

}

func periodicjobstats() ([]periodicJobData, []string) {

	var ex []periodicJobData
	var AllFailures []string
	buildid := []string{}
	clusters := []string{"4.2", "4.3", "4.4", "4.5", "4.6", "4.7"}
	for _, testclusterVersion := range clusters {

		historyurl := fmt.Sprintf("https://prow.ci.openshift.org/job-history/gs/origin-ci-test/logs/periodic-ci-openshift-odo-master-v%s-integration-e2e-periodic?buildId=", testclusterVersion)
		// Request the HTML page.
		res, err := http.Get(historyurl)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}

		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Find the review items
		doc.Find(".table-container tbody .run-failure").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the band and title
			buildid = append(buildid, s.Find("a").Text())
			//bt := strings.Split(s.Find("td").Text(), " ")
			//fmt.Println(dt.Format(bt[28]))
		})
		//for _, bid := range buildid {
		for i := 1; i < len(buildid); i++ {
			// https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/periodic-ci-openshift-odo-master-v4.4-integration-e2e-periodic/1331930925767856128/artifacts/integration-e2e-periodic/integration-e2e-periodic-steps/container-logs/test.log
			logurl := fmt.Sprintf("https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/periodic-ci-openshift-odo-master-v%s-integration-e2e-periodic/%s/artifacts/integration-e2e-periodic/integration-e2e-periodic-steps/container-logs/test.log", testclusterVersion, buildid[i])

			//check the log file and retrive failure reason and store it in the struct
			res, err := http.Get(logurl)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {

				panic(err)
			}
			cleanLine := strings.TrimSpace(string(body))
			cleanLine = StripAnsi(cleanLine)
			for _, failCase := range strings.Split(string(cleanLine), "\n") {
				if strings.Contains(failCase, "[Fail]") {
					//fmt.Println(testclusterVersion, failCase, logurl)
					n := periodicJobData{failure: failCase, clusterVersion: testclusterVersion, url: logurl, flag: false}
					ex = append(ex, n)
					AllFailures = append(AllFailures, failCase)
				}
			}
		}
	}
	occured := map[string]bool{}
	result := []string{}
	for e := range AllFailures {

		// check if already the mapped
		// variable is set to true or not
		if occured[AllFailures[e]] != true {
			occured[AllFailures[e]] = true

			// Append to result slice.
			result = append(result, AllFailures[e])
		}
	}
	return ex, result
}
