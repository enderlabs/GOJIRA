package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"io/ioutil"
	"os"
	"encoding/json"
	"bytes"
	"regexp"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/kr/pretty"
	"github.com/nlopes/slack"
	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
)

var currentTicket *jira.Issue

// https://api.slack.com/slack-apps
// https://api.slack.com/internal-integrations
type envConfig struct {
	// Port is server port to be listened.
	Port string `envconfig:"PORT" default:"3000"`

	// BotToken is bot user token to access to slack API.
	BotToken string `envconfig:"BOT_TOKEN" required:"true"`

	// VerificationToken is used to validate interactive messages from slack.
	VerificationToken string `envconfig:"VERIFICATION_TOKEN" required:"true"`

	// BotID is bot user ID.
	BotID string `envconfig:"BOT_ID" required:"true"`

	// ChannelID is slack channel ID where bot is working.
	// Bot responses to the mention in this channel.
	ChannelID string `envconfig:"CHANNEL_ID" required:"true"`

	JiraUsername string `envconfig:"JIRA_USERNAME" required:"true"`
	JiraPassword string `envconfig:"JIRA_PASSWORD" required:"true"`

	GithubToken string `envconfig:"GITHUB_TOKEN" required:"true"`
}

type SlackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

// LstenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()

	// Start listening slack events
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	// message handler
	return nil
}

// InteractionHandler handles interactive message response.
type InteractionHandler struct {
	slackClient       *slack.Client
	jiraClient        *jira.Client
	gitClient         *github.Client
	verificationToken string
}

func makeDialog(ticketList []string) *slack.Dialog {
	dialog := &slack.Dialog{}
	dialog.CallbackID = "release-create"
	dialog.Title = "Create a Release"
	dialog.SubmitLabel = "Create"
	dialog.NotifyOnCancel = false

	ticketListElement := &slack.TextInputElement{}
	ticketListElement.Label = "Ticket List"
	ticketListElement.Name = "tickets"
	ticketListElement.Type = "textarea"
	ticketListElement.Placeholder = "TEEM-1234\nTEEM-7890"
	ticketListElement.Optional = false
	ticketListElement.Hint = "List of Teem Tickets"
	ticketListElement.Value = strings.Join(ticketList, "\n")

	dialog.Elements = []slack.DialogElement{
		ticketListElement,
	}

	return dialog
}

func createRelease(slackClient *slack.Client, jiraClient *jira.Client, linkedTickets []string) (string, error) {
	ticket, err := createIssue(client)
	var msg string

	if err != nil {
		msg = fmt.Sprintf("Unable to create release: %s", err)
	} else {
		msg = fmt.Sprintf("Release created succesfully! Ticket link: https://teem-gojira.atlassian.net/browse/%s", ticket.Key)
	}

	fmt.Printf("Release creation result: %s\n", msg)

	_, _, err = slackClient.PostMessage(
		"CSBADECGG",
		slack.MsgOptionText(msg, false))

	currentTicket = ticket

	go addTicketsToRelease(jiraClient, linkedTickets, ticket.Key)

	fmt.Printf("RELEASE CREATED: %s\n", ticket.Key)
	fmt.Printf("Error was: %s\n", err)

	return ticket.Key, err
}

func handleReleaseSlashCommand(
	command slack.SlashCommand,
	client *slack.Client,
	jiraClient *jira.Client,
	gitClient *github.Client,
	writer http.ResponseWriter) {

      switch command.Text {
      case "create", "":
	      report := createCommitsComparisonReport(gitClient, "RyanHurstTeem", "TestRepo", "master", "develop")

	      var dialog = makeDialog(report.TicketList())
	      writer.Header().Add("Content-type", "application/json")
	      writer.WriteHeader(http.StatusOK)

	      if err := client.OpenDialog(command.TriggerID, *dialog); err != nil {
		      fmt.Errorf("failed to post message: %s", err)
	      }

	      return

      case "status":
	      if currentTicket == nil {
		      fmt.Printf("No")
		      return
	      }

	      report := statusForIssue(jiraClient, currentTicket.Key)

	      client.PostMessage(
		      "CSBADECGG",
		      slack.MsgOptionText(fmt.Sprintf("%s", report), false))
      default:
	      client.PostEphemeral(
		      command.ChannelID,
		      command.UserID,
		      slack.MsgOptionText(fmt.Sprintf("Invalid command: \"%s\".", command.Text), false))
	      return
      }
}


func (handler InteractionHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", request.Method)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("[Error] Failed to read request body: %s", err)
		return
	}

	request.Body.Close()
	request.Body = ioutil.NopCloser(bytes.NewReader(buf))

	command, err := slack.SlashCommandParse(request)

	fmt.Printf("%# v\n", pretty.Formatter(command))
	fmt.Printf("%# v\n", pretty.Formatter(request.PostForm))

	if err == nil && command.Command != "" {
	      switch command.Command {
	      case "/release":
		      handleReleaseSlashCommand(command, handler.slackClient, handler.jiraClient, handler.gitClient, writer)
		      return
	      default:
	      }
	}

	fmt.Printf("DATA: \"%s\"\n", string(buf))
	fmt.Print("GOT HERE!!!\n")

	jsonBody, err := url.QueryUnescape(string(buf)[8:])

	if err != nil {
		log.Printf("[ERROR] failed to unescape request body: %s", err)
		return
	}

	var dialogSubmission slack.DialogSubmissionCallback
	if err := json.Unmarshal([]byte(jsonBody), &dialogSubmission); err != nil {
		log.Printf("[ERROR] failed to decode dialog submission: %s", jsonBody)
		return
	}

	fmt.Print("\n\n\n")
	ticketList := regexp.MustCompile("[\\s,]").Split(dialogSubmission.Submission["tickets"], -1)

	fmt.Printf("%# v\n", pretty.Formatter(ticketList))

	createRelease(handler.slackClient, handler.jiraClient, ticketList)

}

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return 1
	}


	// Listening slack event and response
	log.Printf("[INFO] Start slack event listening")
	client := slack.New(env.BotToken)
	slackListener := &SlackListener{
		client:    client,
		botID:     env.BotID,
		channelID: env.ChannelID,
	}
	go slackListener.ListenAndResponse()

	tp := jira.BasicAuthTransport{
		Username: env.JiraUsername,
		Password: env.JiraPassword,
	}

	jiraClient, err := jira.NewClient(tp.Client(), "https://teem-gojira.atlassian.net")
	if err != nil {
		panic(err)
	}

	gitClient := connect(env.GithubToken)

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	http.Handle("/interaction", InteractionHandler{
		verificationToken: env.VerificationToken,
		slackClient:       client,
		jiraClient:        jiraClient,
		gitClient:         gitClient,
	})

	log.Printf("[INFO] Server listening on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		log.Printf("[ERROR] %s", err)
		return 1
	}

	return 0
}
