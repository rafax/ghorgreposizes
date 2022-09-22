package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"

	"github.com/davidji99/bitbucket-go/bitbucket"
	"github.com/google/go-github/v47/github"
	"github.com/inhies/go-bytesize"

	"github.com/montanaflynn/stats"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/oauth2"
)

var (
	pageSizeFlag               = flag.Int("page-size", 100, "Size of page to request from backend API")
	gitHubOrgNameFlag          = flag.String("gh-org-name", "", "GitHub organization name")
	gitHubApiTokenFlag         = flag.String("gh-api-token", "", "GitHub API token")
	gitHubBaseUrlFlag          = flag.String("gh-enterprise-base-url", "", "Base URL of GitHub Enterprise instance")
	bitbucketWorkspaceNameFlag = flag.String("bb-workspace-name", "", "Bitbucket workspace name")
	bitbucketUserNameFlag      = flag.String("bb-user-name", "", "Bitbucket user name")
	bitbucketAppPasswordFlag   = flag.String("bb-app-password", "", "Bitbucket app password")
)

type RepoSize struct {
	name string
	size bytesize.ByteSize
}
type RepoStats struct {
	TotalSize bytesize.ByteSize
	MaxSize   float64
	MeanSize  float64
	P99       float64
	P50       float64
	Largest10 []RepoSize
}

func main() {
	flag.Parse()
	if *gitHubOrgNameFlag == "" && *bitbucketWorkspaceNameFlag == "" {
		log.Fatalln("either gh-org-name or bb-workspace-name flag is required")
	}
	ctx := context.Background()
	if *gitHubOrgNameFlag != "" {
		org := *gitHubOrgNameFlag
		client := buildClient(ctx)
		allRepos := fetchReposForOrg(ctx, org, client)
		fmt.Println()
		fmt.Println("Done fetching, calculating size...")
		rs := calculateStats(allRepos)

		fmt.Printf("Found %d repos for org %s, %s total size\n", len(allRepos), org, rs.TotalSize)
		fmt.Printf("max: %v mean: %v p99: %v p50: %v\n", bytesize.New(rs.MaxSize), bytesize.New(rs.MeanSize), bytesize.New(rs.P99), bytesize.New(rs.P50))
		fmt.Print("Top 10 repos by size:\n")
		for _, r := range rs.Largest10 {
			fmt.Printf("%s: %s\n", r.name, r.size)
		}
		return
	}
	if *bitbucketWorkspaceNameFlag != "" {
		ws := *bitbucketWorkspaceNameFlag
		client, err := bitbucket.New(*bitbucketUserNameFlag, *bitbucketAppPasswordFlag)

		if err != nil {
			log.Fatalf("creating BitBucket client failed: %v", err)
		}
		client.Pagelen = 100
		allRepos := fetchReposForWorkspace(ctx, ws, client)
		fmt.Println(allRepos)
		fmt.Println()
		fmt.Println("Done fetching, calculating size...")
		rs := calculateBitbucketStats(allRepos)

		fmt.Printf("Found %d repos for workspace %s, %s total size\n", len(allRepos.Values), ws, rs.TotalSize)
		fmt.Printf("max: %v mean: %v p99: %v p50: %v\n", bytesize.New(rs.MaxSize), bytesize.New(rs.MeanSize), bytesize.New(rs.P99), bytesize.New(rs.P50))
		fmt.Print("Top 10 repos by size:\n")
		for _, r := range rs.Largest10 {
			fmt.Printf("%s: %s\n", r.name, r.size)
		}
	}
}

func calculateBitbucketStats(allRepos *bitbucket.Repositories) RepoStats {
	sizeB := int64(0)
	repoSizeBytes := []int64{}
	for _, r := range allRepos.Values {
		sizeB += *r.Size
		repoSizeBytes = append(repoSizeBytes, *r.Size)
	}
	size, err := bytesize.Parse(fmt.Sprint(sizeB, "B"))
	if err != nil {
		log.Fatalf("calculating size failed: %v", err)
	}
	data := stats.LoadRawData(repoSizeBytes)
	max, err := data.Max()
	if err != nil {
		log.Fatalf("getting max failed: %v", err)
	}
	p99, err := data.Percentile(99)
	if err != nil {
		log.Fatalf("getting p99 failed: %v", err)
	}
	p50, err := data.Percentile(50)
	if err != nil {
		log.Fatalf("getting p50 failed: %v", err)
	}
	mean, err := data.Mean()
	if err != nil {
		log.Fatalf("getting mean failed: %v", err)
	}
	sort.Slice(allRepos.Values, func(i, j int) bool { return allRepos.Values[i].GetSize() > allRepos.Values[j].GetSize() })
	topN := 10
	if len(allRepos.Values) < topN {
		topN = len(allRepos.Values)
	}
	top10 := []RepoSize{}
	for _, r := range allRepos.Values[:topN] {
		size, err := bytesize.Parse(fmt.Sprint(r.GetSize(), "B"))
		if err != nil {
			log.Fatalf("calculating size failed: %v", err)
		}
		top10 = append(top10, RepoSize{name: r.GetName(), size: size})
	}
	return RepoStats{TotalSize: size, MaxSize: max, MeanSize: mean, P99: p99, P50: p50, Largest10: top10}
}

func fetchReposForWorkspace(ctx context.Context, ws string, client *bitbucket.Client) *bitbucket.Repositories {
	if *bitbucketUserNameFlag == "" {
		log.Fatalln("bb-user-name flag is required")
	}
	if *bitbucketAppPasswordFlag == "" {
		log.Fatalln("bb-app-password flag is required")
	}
	repos, _, err := client.Repositories.List(ws)
	if err != nil {
		log.Fatalf("getting max failed: %v", err)
	}
	return repos
}

func calculateStats(allRepos []*github.Repository) RepoStats {
	sizeKB := 0
	repoSizeBytes := []int{}
	for _, r := range allRepos {
		sizeKB += *r.Size
		repoSizeBytes = append(repoSizeBytes, r.GetSize()*1024)
	}
	size, err := bytesize.Parse(fmt.Sprint(sizeKB, "KB"))
	if err != nil {
		log.Fatalf("calculating size failed: %v", err)
	}
	data := stats.LoadRawData(repoSizeBytes)
	max, err := data.Max()
	if err != nil {
		log.Fatalf("getting max failed: %v", err)
	}
	p99, err := data.Percentile(99)
	if err != nil {
		log.Fatalf("getting p99 failed: %v", err)
	}
	p50, err := data.Percentile(50)
	if err != nil {
		log.Fatalf("getting p50 failed: %v", err)
	}
	mean, err := data.Mean()
	if err != nil {
		log.Fatalf("getting mean failed: %v", err)
	}
	sort.Slice(allRepos, func(i, j int) bool { return allRepos[i].GetSize() > allRepos[j].GetSize() })
	topN := 10
	if len(allRepos) < topN {
		topN = len(allRepos)
	}
	top10 := []RepoSize{}
	for _, r := range allRepos[:topN] {
		size, err := bytesize.Parse(fmt.Sprint(r.GetSize(), "KB"))
		if err != nil {
			log.Fatalf("calculating size failed: %v", err)
		}
		top10 = append(top10, RepoSize{name: r.GetName(), size: size})
	}
	return RepoStats{TotalSize: size, MaxSize: max, MeanSize: mean, P99: p99, P50: p50, Largest10: top10}
}

func fetchReposForOrg(ctx context.Context, org string, client *github.Client) []*github.Repository {
	bar := progressbar.Default(-1, "Fetching repos for org "+org)
	opt := &github.RepositoryListByOrgOptions{Type: "public", ListOptions: github.ListOptions{PerPage: *pageSizeFlag}}
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			log.Fatalf("fetching repos failed: %v", err)
		}
		err = bar.Add(len(repos))
		if err != nil {
			log.Printf("error when updating progress bar [%v], continuing", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos
}

func buildClient(ctx context.Context) *github.Client {
	if *gitHubApiTokenFlag == "" {
		log.Fatalln("api-token flag is required")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *gitHubApiTokenFlag},
	)
	tc := oauth2.NewClient(ctx, ts)
	if *gitHubBaseUrlFlag != "" {
		client, err := github.NewEnterpriseClient(*gitHubBaseUrlFlag, "", tc)
		if err != nil {
			log.Fatalf("creating GHE client failed: %v", err)
		}
		return client
	}
	return github.NewClient(tc)
}
