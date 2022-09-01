package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"

	"github.com/google/go-github/v47/github"
	"github.com/inhies/go-bytesize"
	"github.com/montanaflynn/stats"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/oauth2"
)

var (
	pageSizeFlag = flag.Int("page-size", 100, "Size of page to request from GitHub API")
	orgNameFlag  = flag.String("org-name", "", "Size of page to request from GitHub API")
	apiTokenFlag = flag.String("api-token", "", "Size of page to request from GitHub API")
	baseUrlFlag  = flag.String("enterprise-base-url", "", "Base URL of GitHub Enterprise instance")
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
	if *orgNameFlag == "" {
		log.Fatalln("org-name flag is required")
	}
	org := *orgNameFlag
	ctx := context.Background()
	client := buildClient(ctx)
	allRepos := fetchReposForOrg(ctx, org, client)
	fmt.Println()
	fmt.Println("Done fetching, calculating size...")
	rs := calculateStats(allRepos)

	fmt.Printf("Found %d repos for org %s, %s total size\n", len(allRepos), *orgNameFlag, rs.TotalSize)
	fmt.Printf("max: %v mean: %v p99: %v p50: %v\n", bytesize.New(rs.MaxSize), bytesize.New(rs.MeanSize), bytesize.New(rs.P99), bytesize.New(rs.P50))
	fmt.Print("Top 10 repos by size:\n")
	for _, r := range rs.Largest10 {
		fmt.Printf("%s: %s\n", r.name, r.size)
	}
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
	top10 := []RepoSize{}
	for _, r := range allRepos[:10] {
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
	if *apiTokenFlag == "" {
		log.Fatalln("api-token flag is required")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *apiTokenFlag},
	)
	tc := oauth2.NewClient(ctx, ts)
	if *baseUrlFlag != "" {
		client, err := github.NewEnterpriseClient(*baseUrlFlag, "", tc)
		if err != nil {
			log.Fatalf("creating GHE client failed: %v", err)
		}
		return client
	}
	return github.NewClient(tc)
}
