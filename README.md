# ghorgreposizes
Get the total size of public repositories owned by a Github org.

For github.com, use:
```
go run main.go --org-name=github --api-token=ghp_YOURTOKEN
```
where `api-token` is a token [accepted by go-github API client](https://github.com/google/go-github#authentication).

For GitHub enterprise, use:
```
go run main.go --org-name=github --api-token=ghp_YOURTOKEN  --enterprise-base-url=https://ghe.example.org
```
where `api-token` is a token [accepted by go-github API client](https://github.com/google/go-github#authentication) and `enterprise-base-url` is the base URL of your GitHub Enterprise instance.


Sample output:
```
⠹ Fetching repos for org github (427/-, 45 it/s) [9s]
Done fetching, calculating size...
Found 427 repos for org github, 14.28GB total size
max: 2.02GB mean: 34.25MB p99: 830.76MB p50: 532.00KB
```