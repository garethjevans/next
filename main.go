package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/garethjevans/next/github"
	"log"
	"os"
	"sort"
)

var (
	sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
	sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
	host        = flag.String("host", "github.com", "The GitHub host to connect to")
)

var ctx = context.Background()

func main() {
	var err error
	flag.Parse()
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}

	var client github.GitHub
	if *host == "github.com" {
		client = github.New(token)
	} else {
		client = github.NewEnterpriseClient(token, fmt.Sprintf("https://%s/api/graphql", *host))
	}

	commitsFromReleases, err := client.FetchCommitsFromReleases(*sourceOwner, *sourceRepo)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("all releases %+v\n", commitsFromReleases)
	var versions []*semver.Version

	for _, v := range commitsFromReleases {
		versions = append(versions, semver.MustParse(v))
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))
	currentVersion := versions[0]

	// fmt.Printf("latest version %+s\n", currentVersion.String())
	var commit string
	for c, v := range commitsFromReleases {
		if currentVersion.String() == v {
			commit = c
		}
	}

	var allPullRequests []github.PullRequest

	if commit != "" {
		latestCommitSHA, err := client.FetchLatestReleaseCommitFromBranch(*sourceOwner, *sourceRepo, "main", commitsFromReleases)
		if err != nil {
			panic(err)
		}

		prs, err := client.FetchPullRequestsAfterCommit(*sourceOwner, *sourceRepo, "main", commit, latestCommitSHA)
		if err != nil {
			panic(err)
		}
		allPullRequests = append(allPullRequests, prs...)
	}

	//fmt.Printf("all prs %+v\n", allPullRequests)
	if containsLabel(allPullRequests, "semver:major") {
		fmt.Println(currentVersion.IncMajor().String())
	} else if containsLabel(allPullRequests, "semver:minor") {
		fmt.Println(currentVersion.IncMinor().String())
	} else {
		fmt.Println(currentVersion.IncPatch().String())
	}

}

func containsLabel(requests []github.PullRequest, label string) bool {
	for _, pr := range requests {
		if pr.HasLabel(label) {
			return true
		}
	}
	return false
}
