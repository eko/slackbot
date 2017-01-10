// Slackbot - A Slack robot library written in Go
//
// Author: Vincent Composieux <vincent.composieux@gmail.com>

package slackbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/websocket"
)

var (
	Token           string
	WebsocketStream *websocket.Conn
	BotIdentifier   string
	MessageCounter  uint64
	commands        []Command
)

type JsonResponse struct {
	Ok   bool             `json:"ok"`
	URL  string           `json:"url"`
	Self JsonResponseSelf `json:"self"`
}

type JsonResponseSelf struct {
	ID string `json:"id"`
}

type Message struct {
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type handler func(pattern Command, message Message)

type Command struct {
	Pattern *regexp.Regexp
	Handler handler
}

func check_error(err error) {
	if err != nil {
		panic(err)
	}
}

// Init function initializes the Slack websocket.
func Init() {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", Token)

	response, err := http.Get(url)
	check_error(err)

	if response.StatusCode != 200 {
		err = fmt.Errorf("Unable to connect to streaming API. Code status: %d", response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	check_error(err)

	var jsonResponse JsonResponse
	err = json.Unmarshal(body, &jsonResponse)
	check_error(err)

	WebsocketStream, err = websocket.Dial(jsonResponse.URL, "", "https://api.slack.com/")
	check_error(err)

	BotIdentifier = jsonResponse.Self.ID
}

// Stream function listens for the websocket for new messages.
func Stream() {
	fmt.Println("Bot is ready, hit ^C to exit.")

	for {
		var message Message
		err := websocket.JSON.Receive(WebsocketStream, &message)
		check_error(err)

		prefix := "<@" + BotIdentifier + "> "

		if message.Type == "message" && strings.HasPrefix(message.Text, prefix) {
			go func(commands []Command, message Message) {
				message.Text = strings.Replace(message.Text, prefix, "", -1)

				for _, command := range commands {
					if m, _ := regexp.MatchString(command.Pattern.String(), message.Text); m {
						fmt.Printf("-> Command: %s\n", message.Text)
						command.Handler(command, message)
						break
					}
				}
			}(commands, message)
		}
	}
}

// AddCommand function adds a new command into the commands list.
func AddCommand(pattern string, handler handler) {
	commands = append(commands, Command{Pattern: regexp.MustCompile(pattern), Handler: handler})
}

// Send function sends a new message into the websocket stream.
func Send(message Message) error {
	return websocket.JSON.Send(WebsocketStream, message)
}
