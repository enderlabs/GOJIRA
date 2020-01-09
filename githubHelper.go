package main

import (
	"context"
	"golang.org/x/oauth2"
	"regexp"
)
import "github.com/google/go-github/github" // with go modules disabled

var ctx context.Context
var client *github.Client

func setup(accessToken string) {
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)
}

//func createPullRequestReport(owner string, repo string, pullRequestNumber int) PullRequestReport {
//	var pullRequestReport PullRequestReport
//	pullRequest, _, _ := client.PullRequests.Get(ctx, owner, repo, pullRequestNumber)
//	if pullRequest != nil {
//		pullRequestReport.pullRequest = pullRequest
//	}
//	commits, _, err := client.PullRequests.ListCommits(ctx, owner, repo, pullRequestNumber, nil)
//	if err != nil {
//		println("ERROR comparing commits: ", err.Error())
//		return pullRequestReport
//	}
//	pullRequestReport.tickets = createSomeShit(commits)
//	return pullRequestReport
//}

func createCommitsComparisonReport(owner string, repo string, base string, head string) CommitsComparisonReport {
	var commitsComparisonReport CommitsComparisonReport
	commitsComparison, _, err := client.Repositories.CompareCommits(ctx, owner, repo, base, head)
	if commitsComparison != nil {
		commitsComparisonReport.commitsComparison = commitsComparison
	}
	if err != nil {
		println("ERROR comparing commits: ", err.Error())
		return commitsComparisonReport
	}

	commitsComparisonReport.tickets = createSomeShit(commitsComparison.Commits)
	return commitsComparisonReport
}

func createNewReleaseBranch(owner string, repo string, branchName string, sha string) *github.Reference{
	gitObject := &github.GitObject{
		Type: nil,
		URL:  nil,
		SHA:  github.String(sha),
	}
	var refString = "refs/heads/" + branchName
	ref := &github.Reference{
		Ref:    &refString,
		URL:    nil,
		Object: gitObject,
		NodeID: nil,
	}
	returnedReference, _, err := client.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		println("ERROR creating ref: ", err.Error())
	}
	return returnedReference
}

func createSomeShit(commits []github.RepositoryCommit) map[string][]*github.Commit {
	ticketMap := make(map[string][]*github.Commit)
	if commits == nil {
		return ticketMap
	}
	r, _ := regexp.Compile("(?i)teem-([0-9]+)")
	for _, commit := range commits {
		commitMessage := *commit.Commit.Message
		ticketNumber := r.FindStringSubmatch(commitMessage)
		if ticketNumber != nil {
			if ticketMap[ticketNumber[1]] == nil {
				ticketMap[ticketNumber[1]] = []*github.Commit{}
			}
			ticketMap[ticketNumber[1]] = append(ticketMap[ticketNumber[1]], commit.Commit)
		}
	}
	return ticketMap
}

func createPullRequest(owner string, repo string, base *string, head *string, title *string, body *string) *github.PullRequest {
	pull := &github.NewPullRequest{
		Title:               title,
		Body:                body,
		Base:                base,
		Head:                head,
		Issue:               nil,
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	}
	pullRequest, _, err := client.PullRequests.Create(ctx, owner, repo, pull)
	if err != nil {
		println("ERROR creating PR: ", err.Error())
	}
	return pullRequest
}

//func (pullRequestReport PullRequestReport) String() string {
//	var pullRequestString = *pullRequestReport.pullRequest.Title
//	for key, commits := range pullRequestReport.tickets {
//		pullRequestString += "\n"
//		pullRequestString += key
//		for _, commit := range commits {
//			pullRequestString += "\n    " + *commit.Message + " by: " + *commit.Author.Name
//		}
//	}
//	return pullRequestString
//}
//
//type PullRequestReport struct {
//	pullRequest *github.PullRequest
//	tickets     map[string][]*github.Commit
//}

func (commitsComparisonReport CommitsComparisonReport) String() string {
	var pullRequestString = *commitsComparisonReport.commitsComparison.Status
	for key, commits := range commitsComparisonReport.tickets {
		pullRequestString += "\n"
		pullRequestString += key
		for _, commit := range commits {
			pullRequestString += "\n    " + *commit.Message + " by: " + *commit.Author.Name
		}
	}
	return pullRequestString
}

type CommitsComparisonReport struct {
	commitsComparison *github.CommitsComparison
	tickets     map[string][]*github.Commit
}
