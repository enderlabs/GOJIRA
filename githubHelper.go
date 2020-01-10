package main

import (
	"context"
	"golang.org/x/oauth2"
	"regexp"
)
import "github.com/google/go-github/github" // with go modules disabled

var ctx context.Context

func connect(accessToken string) *github.Client {
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}

func createCommitsComparisonReport(client *github.Client, owner string, repo string, base string, head string) CommitsComparisonReport {
	var commitsComparisonReport CommitsComparisonReport
	commitsComparison, _, err := client.Repositories.CompareCommits(ctx, owner, repo, base, head)
	if commitsComparison != nil {
		commitsComparisonReport.commitsComparison = commitsComparison
	}
	if err != nil {
		println("ERROR comparing commits: ", err.Error())
		return commitsComparisonReport
	}

	commitsComparisonReport.tickets = mapTickets(commitsComparison.Commits)
	return commitsComparisonReport
}

func createNewReleaseBranch(client *github.Client, owner string, repo string, branchName string, sha string) *github.Reference {
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

func mapTickets(commits []github.RepositoryCommit) map[string][]*github.Commit {
	ticketMap := make(map[string][]*github.Commit)
	if commits == nil {
		return ticketMap
	}
	r, _ := regexp.Compile("(?i)or-([0-9]+)")
	for _, commit := range commits {
		commitMessage := *commit.Commit.Message
		ticketNumber := r.FindString(commitMessage)
		if ticketNumber != "" {
			if ticketMap[ticketNumber] == nil {
				ticketMap[ticketNumber] = []*github.Commit{}
			}
			ticketMap[ticketNumber] = append(ticketMap[ticketNumber], commit.Commit)
		}
	}
	return ticketMap
}

func createPullRequest(client *github.Client, owner string, repo string, base *string, head *string, title *string, body *string) *github.PullRequest {
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

func (commitsComparisonReport CommitsComparisonReport) String() string {
	var commitsComparisonString = *commitsComparisonReport.commitsComparison.Status
	for key, commits := range commitsComparisonReport.tickets {
		commitsComparisonString += "\n"
		commitsComparisonString += key
		for _, commit := range commits {
			commitsComparisonString += "\n    " + *commit.Message + " by: " + *commit.Author.Name
		}
	}
	return commitsComparisonString
}

func (commitsComparisonReport CommitsComparisonReport) TicketList() []string {
	keys := make([]string, 0, len(commitsComparisonReport.tickets))
	for k := range commitsComparisonReport.tickets {
		keys = append(keys, k)
	}

	return keys
}

type CommitsComparisonReport struct {
	commitsComparison *github.CommitsComparison
	tickets           map[string][]*github.Commit
}
