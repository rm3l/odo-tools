package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func StripAnsi(str string) string {
	return re.ReplaceAllString(str, "")
}

type Match struct {
	FileType  string   `json:"filename"`
	Context   []string `json:"context,omitempty"`
	MoreLines int      `json:"moreLines,omitempty"`
}

type Result map[string]map[string][]Match

func main() {

	// some simple stats
	TestFailCount := map[string]int{}

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

	// itarate over all results
	for _, search := range result {
		//fmt.Printf("%s\n", file)
		for _, matches := range search {
			// fmt.Printf("  %v\n", regexp)
			for _, match := range matches {
				for _, line := range match.Context {
					//fmt.Printf("    %v\n", line)
					cleanLine := strings.TrimSpace(line)
					cleanLine = StripAnsi(cleanLine)
					TestFailCount[cleanLine]++
				}
			}
		}

	}

	type TestFails struct {
		TestName string
		Fails    int
	}
	// convert tests to slice so we can easily sort it
	fails := []TestFails{}
	for test, count := range TestFailCount {
		fails = append(fails, TestFails{TestName: test, Fails: count})
	}

	sort.Slice(fails, func(i, j int) bool { return fails[i].Fails > fails[j].Fails })

	fmt.Println("# odo test statistics")
	fmt.Printf("Last update: %s (UTC)\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Println("## FLAKY TESTS: Failed test scenarios in past 14 days")
	fmt.Println("| Failures | Test Name |")
	fmt.Println("|---|---|")
	for _, f := range fails {
		fmt.Printf("| %d | %s |\n", f.Fails, f.TestName)
	}

}
