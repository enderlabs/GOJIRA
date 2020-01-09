package main

import (
	"context"
	"golang.org/x/oauth2"
	"regexp"
)
import "github.com/google/go-github/github" // with go modules disabled

var ctx context.Context
var client *github.Client

func main() {
	setup("4ae373d0cc57fef0aa53463077599f5744bc1f30")

	//createNewReleaseBranch("RyanHurstTeem", "TestRepo", "release-branch", "b057e74c9b3aeb59ca3f5456acee0232c11a7d99")
	createPullRequest(
		"RyanHurstTeem",
		"TestRepo",
		github.String("release-branch"),
		github.String("master"),
		github.String("test PR"),
		github.String("this is the body of the test PR"),
	)
	//println(createReleaseDocument("enderlabs", "android", 538).String())
}

func setup(accessToken string) {
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)
}

func createReleaseDocument(owner string, repo string, pullRequestNumber int) PullRequestReport {
	var pullRequestReport PullRequestReport
	pullRequestReport.tickets = make(map[string][]*github.Commit)
	pullRequest, _, _ := client.PullRequests.Get(ctx, owner, repo, pullRequestNumber)
	if pullRequest != nil {
		pullRequestReport.pullRequest = pullRequest
	}
	commits, _, _ := client.PullRequests.ListCommits(ctx, owner, repo, pullRequestNumber, nil)
	r, _ := regexp.Compile("(?i)teem-([0-9]+)")
	if commits != nil {
		for _, commit := range commits {
			commitMessage := *commit.Commit.Message
			ticketNumber := r.FindStringSubmatch(commitMessage)
			if ticketNumber != nil {
				if pullRequestReport.tickets[ticketNumber[1]] == nil {
					pullRequestReport.tickets[ticketNumber[1]] = []*github.Commit{}
				}
				pullRequestReport.tickets[ticketNumber[1]] = append(pullRequestReport.tickets[ticketNumber[1]], commit.Commit)
			}
		}
	}
	return pullRequestReport
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

func createPullRequest(owner string, repo string, head *string, base *string, title *string, body *string) *github.PullRequest {
	pull := &github.NewPullRequest{
		Title:               title,
		Body:                body,
		Head:                head,
		Base:                base,
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

func (pullRequestReport PullRequestReport) String() string {
	var pullRequestString = *pullRequestReport.pullRequest.Title
	for key, commits := range pullRequestReport.tickets {
		pullRequestString += "\n"
		pullRequestString += key
		for _, commit := range commits {
			pullRequestString += "\n    " + *commit.Message + " by: " + *commit.Author.Name
		}
	}
	return pullRequestString
}

type PullRequestReport struct {
	pullRequest *github.PullRequest
	tickets     map[string][]*github.Commit
}
