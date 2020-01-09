package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"strconv"
)

func findIssue(id string) *jira.Issue {
     tp := jira.BasicAuthTransport{
	  Username: "nathancannonperry@gmail.com",
	  Password: "oueM1EGE3Twk0gGRteGI7CAA",
     }


     client, err := jira.NewClient(tp.Client(), "https://teem-gojira.atlassian.net")
     if err != nil {
	  panic(err)
     }
     issue, _, err := client.Issue.Get(id, nil)
     if err != nil {
	  panic(err)
     }
     return issue
}

type TicketReport struct {
     completedCount int
     totalCount	    int
     tickets	    map[string]string
}

func (ticketReport TicketReport) String() string {
     statuses := ""
     for key, status := range ticketReport.tickets {
	  statuses = statuses + key + ": " + status + "\n"
     }
     statuses = statuses + "Completed: " + strconv.Itoa(ticketReport.completedCount) + "/" + strconv.Itoa(ticketReport.totalCount) + "\n"
     return statuses
}

func statusForIssue(id string) TicketReport {
     issue := findIssue(id)
     linked := issue.Fields.IssueLinks
     finishedCount := 0
     var report TicketReport
     report.tickets = make(map[string]string)
     for _, issue := range linked {
	  var linkedIssue *jira.Issue
	  if issue.OutwardIssue != nil {
	       linkedIssue = issue.OutwardIssue
	  } else if issue.InwardIssue != nil {
	       linkedIssue = issue.InwardIssue
	  } else {
	       continue
	  }

	  key := linkedIssue.Key
	  status := linkedIssue.Fields.Status.Name
	  report.tickets[key] = status
	  if status == "Done" {
	       finishedCount += 1
	  }
     }
     report.completedCount = finishedCount
     report.totalCount = len(linked)
     return report
}

func createIssue() *jira.Issue {
     tp := jira.BasicAuthTransport{
	  Username: "nathancannonperry@gmail.com",
	  Password: "oueM1EGE3Twk0gGRteGI7CAA",
     }


     client, err := jira.NewClient(tp.Client(), "https://teem-gojira.atlassian.net")
     if err != nil {
	  panic(err)
     }
     i := jira.Issue{
	     Fields: &jira.IssueFields{
		     Assignee: &jira.User{
			     AccountID: "557058:0867a421-a9ee-4659-801a-bc0ee4a4487e",
		     },
		     Type: jira.IssueType{
			     ID: "10006",
		     },
		     Project: jira.Project{
			     ID: "10002",
		     },
		     Summary: "lets do this",
	     },
     }
     fmt.Printf("trying to make: %s\n", i.Fields.Summary)

     newIssue, newBody, newErr := client.Issue.Create(&i)
     if newErr != nil {
	  fmt.Printf("body: %s\n", newBody)
	  fmt.Printf("issue: %s\n", newIssue)
	  panic(newErr)
     }

     fmt.Printf("issue created!\n")
     fmt.Printf("%s\n", newIssue)
     return newIssue
}

func firstBlockedBySecond(firstKey string, secondKey string) {
          tp := jira.BasicAuthTransport{
	  Username: "nathancannonperry@gmail.com",
	  Password: "oueM1EGE3Twk0gGRteGI7CAA",
     }


     client, err := jira.NewClient(tp.Client(), "https://teem-gojira.atlassian.net")
     if err != nil {
	  panic(err)
     }
     link := &jira.IssueLink{
	  Type: jira.IssueLinkType{
	       Name: "Blocks",
	  },
	  OutwardIssue: &jira.Issue{
	       Key: firstKey,
	  },
	  InwardIssue: &jira.Issue{
	       Key: secondKey,
	  },
     }
     response, error := client.Issue.AddLink(link)
     fmt.Printf("link response: %s\n", response)
     fmt.Printf("link error: %s\n", error)
}

func addTicketsToRelease(tickets []string, release string) {
     for _, ticket := range tickets {
	  firstBlockedBySecond(release, ticket)
     }
}

//func main() {
//     firstBlockedBySecond("OR-9", "OR-10")
//     createIssue()
//     report := statusForIssue("OR-10")
//     fmt.Printf("%s", report)
//     tickets := []string{"OR-9", "OR-15", "OR-16", "OR-17"}
//     addTicketsToRelease(tickets, "OR-10")
//}

