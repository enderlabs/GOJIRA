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

	"github.com/kelseyhightower/envconfig"
	"github.com/kr/pretty"
	"github.com/nlopes/slack"
)

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
	verificationToken string
}

func makeDialog() *slack.Dialog {
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

	projectSelectElement := &slack.DialogInputSelect{}
	projectSelectElement.Label = "Select Project"
	projectSelectElement.Name = "project"
	projectSelectElement.Type = "select"
	projectSelectElement.Optional = false
	projectSelectElement.Value = "1"
	projectSelectElement.Options = []slack.DialogSelectOption{
		{Label: "iOS", Value: "1"},
		{Label: "Android", Value: "2"},
	}
	projectSelectElement.SelectedOptions = []slack.DialogSelectOption{
		{Label: "iOS", Value: "1"},
	}

	dialog.Elements = []slack.DialogElement{
		projectSelectElement,
		ticketListElement,
	}

	return dialog
}

func handleReleaseSlashCommand(command slack.SlashCommand, client *slack.Client, writer http.ResponseWriter) {
      switch command.Text {
      case "create", "":
	      var dialog = makeDialog()
	      writer.Header().Add("Content-type", "application/json")
	      writer.WriteHeader(http.StatusOK)
	      fmt.Print("\n\n\n")
	      if err := client.OpenDialog(command.TriggerID, *dialog); err != nil {
		      fmt.Errorf("failed to post message: %s", err)
	      }
	      return
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
		      handleReleaseSlashCommand(command, handler.slackClient, writer)
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

	fmt.Printf("%# v", pretty.Formatter(dialogSubmission))
	fmt.Printf("Project: %s\n", dialogSubmission.Submission["project"])
	fmt.Printf("Tickets: %s\n", dialogSubmission.Submission["tickets"])
	stuff := dialogSubmission.Submission["bwent"]
	if stuff == "" {
		fmt.Printf("Whoa it worked\n")
		return
	}
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

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	http.Handle("/interaction", InteractionHandler{
		verificationToken: env.VerificationToken,
		slackClient:       client,
	})

	log.Printf("[INFO] Server listening on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		log.Printf("[ERROR] %s", err)
		return 1
	}

	return 0
}
