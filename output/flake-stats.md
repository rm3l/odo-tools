# odo test statistics
Last update: 2022-02-20 02:04:18 (UTC)

Generated with https://github.com/jgwest/odo-tools/ and https://github.com/kadel/odo-tools
## FLAKY TESTS: Failed test scenarios in past 14 days
| Failure Score<sup>*</sup> | Failures | Test Name | Last Seen | PR List and Logs 
|---|---|---|---|---|
| 40 | 2 | [Fail] odo devfile init command tests when devfile contains parent URI [It] should not replace the original devfile  |  | 2: [#5467](https://github.com/openshift/odo/pull/5467)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/openshift_odo/5467/pull-ci-redhat-developer-odo-main-v4.9-integration-e2e/1494354107727613952/build-log.txt)</sup> [#5434](https://github.com/openshift/odo/pull/5434)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/openshift_odo/5434/pull-ci-redhat-developer-odo-main-v4.9-integration-e2e/1491017991117082624/build-log.txt)</sup> 
| 40 | 2 | [Fail] odo devfile init command tests when running odo init with valid flags [BeforeEach] should download a devfile.yaml file  |  | 2: [#5467](https://github.com/openshift/odo/pull/5467)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/openshift_odo/5467/pull-ci-redhat-developer-odo-main-v4.9-integration-e2e/1494354107727613952/build-log.txt)</sup> [#5434](https://github.com/openshift/odo/pull/5434)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/openshift_odo/5434/pull-ci-redhat-developer-odo-main-v4.9-integration-e2e/1491017991117082624/build-log.txt)</sup> 
| 40 | 2 | [Fail] odo devfile push command tests when creating a nodejs component when doing odo push when the pod is deleted and replaced [It] should execute run command on odo push  |  | 2: [#5460](https://github.com/openshift/odo/pull/5460)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/openshift_odo/5460/pull-ci-redhat-developer-odo-main-v4.9-integration-e2e/1493563633987227648/build-log.txt)</sup> [#5428](https://github.com/openshift/odo/pull/5428)<sup>[1](https://storage.googleapis.com/origin-ci-test/pr-logs/pull/openshift_odo/5428/pull-ci-redhat-developer-odo-main-v4.9-integration-e2e/1491473934741147648/build-log.txt)</sup> 


# odo test statistics for periodic jobs
Last update: 2022-02-20 02:04:19 (UTC)

| Failure Score<sup>*</sup> | Failures | Test Name | Last Seen | Cluster version and Logs 
|---|---|---|---|---|
| 200 | 5 | [Fail] odo devfile push command tests when Create and push java-springboot component [It] should execute default build and run commands correctly  |  | 4: [v4.10]<sup>[1](https://storage.googleapis.com/origin-ci-test/logs/periodic-ci-redhat-developer-odo-main-v4.10-integration-e2e-periodic/1490747291840024576/build-log.txt)</sup> [v4.8]<sup>[1](https://storage.googleapis.com/origin-ci-test/logs/periodic-ci-redhat-developer-odo-main-v4.8-integration-e2e-periodic/1492740390317461504/build-log.txt), [2](https://storage.googleapis.com/origin-ci-test/logs/periodic-ci-redhat-developer-odo-main-v4.8-integration-e2e-periodic/1490203784008175616/build-log.txt)</sup> [v4.7]<sup>[1](https://storage.googleapis.com/origin-ci-test/logs/periodic-ci-redhat-developer-odo-main-v4.7-integration-e2e-periodic/1490928490445279232/build-log.txt)</sup> [v4.9]<sup>[1](https://storage.googleapis.com/origin-ci-test/logs/periodic-ci-redhat-developer-odo-main-v4.9-integration-e2e-periodic/1493918338915504128/build-log.txt)</sup> 



<sup>*</sup> - Failure score is an arbitrary severity estimate, and is approximately `(# of PRs the test failure was seen in * # of test failures) / (days since failure)`. See code for full algorithm -- PRs welcome for algorithm improvements.
