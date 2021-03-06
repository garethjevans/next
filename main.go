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
	debug       = flag.Bool("debug", false, "Enable debug logging")
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
		gheHost := fmt.Sprintf("https://%s/api/graphql", *host)
		if *debug {
			fmt.Println("[DEBUG]", "host", gheHost)
		}
		client = github.NewEnterpriseClient(token, gheHost)
	}

	commitsFromReleases, err := client.FetchCommitsFromReleases(*sourceOwner, *sourceRepo)
	if err != nil {
		panic(err)
	}

	if *debug {
		fmt.Printf("[DEBUG] all releases %+v\n", commitsFromReleases)
	}

	var versions []*semver.Version

	for version := range commitsFromReleases {
		versions = append(versions, semver.MustParse(version))
	}

	if len(versions) == 0 {
		// the initial version
		fmt.Println("0.1.0")
		return
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))
	currentVersion := versions[0]

	if *debug {
		fmt.Printf("[DEBUG] latest version %+s\n", currentVersion.String())
	}
	commit := commitsFromReleases[currentVersion.String()]

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
