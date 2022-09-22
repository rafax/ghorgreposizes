# ghorgreposizes
Get the total size of repositories owned by a Github org or Bitbucket Cloud workspace.

For github.com, use:
```
go run main.go --gh-org-name=github --gh-api-token=ghp_YOURTOKEN
```
where `api-token` is a token [accepted by go-github API client](https://github.com/google/go-github#authentication).

For GitHub Enterprise, use:
```
go run main.go --gh-org-name=github --gh-api-token=ghp_YOURTOKEN  --gh-enterprise-base-url=https://ghe.example.org
```
where `api-token` is a token [accepted by go-github API client](https://github.com/google/go-github#authentication) and `enterprise-base-url` is the base URL of your GitHub Enterprise instance.

For Bitbucket Cloud, use:
```
go run main.go --bb-workspace-name=YOUR_WORKSPACE --bb-app-password=APP_PASSWORD --bb-user-name=USERNAME
```
where `YOUR_WORKSPACE` is the name of your Bitbucket Cloud workspace, `APP_PASSWORD` is your Atlassian [App Password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) and USERNAME is your Atlassian user name.

Sample output:
```
⠹ Fetching repos for org github (427/-, 45 it/s) [9s]
Done fetching, calculating size...
Found 427 repos for org github, 14.28GB total size
max: 2.02GB mean: 34.25MB p99: 830.76MB p50: 532.00KB
```