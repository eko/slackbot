// Slackbot - A Slack robot library written in Go
//
// Author: Vincent Composieux <vincent.composieux@gmail.com>

package slackbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/websocket"
)

var (
	Token           string
	WebsocketStream *websocket.Conn
	BotIdentifier   string
	MessageCounter  uint64
	commands        []Command
	RequirePrefix   bool = true
)

type RtmJsonResponse struct {
	Ok    bool                `json:"ok"`
	Error string              `json:"error"`
	URL   string              `json:"url"`
	Self  RtmJsonResponseSelf `json:"self"`
}

type RtmJsonResponseSelf struct {
	ID string `json:"id"`
}

type Channel struct {
	User     string `json:"user"`
	Token    string `json:"token"`
	ReturnIM bool   `json:"return_im"`
}

type MPInstantMessage struct {
	Users    string `json:"users"`
	Token    string `json:"token"`
	ReturnIM bool   `json:"return_im"`
}

type MPInstantMessageJsonResponse struct {
	Ok    bool              `json:"ok"`
	Error string            `json:"error"`
	Group GroupJsonResponse `json:"group"`
}

type GroupJsonResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Message struct {
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	AsUser  bool   `json:"as_user"`
	Text    string `json:"text"`
	Token   string `json:"token"`
}

type ChannelJsonResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ChannelsJsonResponse struct {
	Ok       bool                  `json:"ok"`
	Error    string                `json:"error"`
	Channels []ChannelJsonResponse `json:"channels"`
}

type IMJsonResponse struct {
	Ok      bool                `json:"ok"`
	Error   string              `json:"error"`
	Channel ChannelJsonResponse `json:"channel"`
}

type UserJsonResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

type UsersJsonResponse struct {
	Ok      bool               `json:"ok"`
	Error   string             `json:"error"`
	Members []UserJsonResponse `json:"members"`
}

type handler func(pattern Command, message Message)

type Command struct {
	Pattern     *regexp.Regexp
	Name        string
	Description string
	Handler     handler
}

func check_error(err error) {
	if err != nil {
		panic(err)
	}
}

// Returns a string containing a pretty list of commands using the Name and Description fields in the Command struct
func generateHelpOutput() string {
	helpString := ""
	for _, command := range commands {
		helpString = fmt.Sprintf("`%s`: %s\n", command.Name, command.Description)
	}
	return helpString
}

// Check if prefix requirement is enabled, and check for the prefix if so.
func checkPrefix(message Message, prefix string, requirement bool) bool {
	if requirement == false {
		return true
	} else if strings.HasPrefix(message.Text, prefix) {
		return true
	} else {
		return false
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

	var jsonResponse RtmJsonResponse
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

		if message.Type == "message" && checkPrefix(message, prefix, RequirePrefix) == true {
			go func(commands []Command, message Message) {
				message.Text = strings.Replace(message.Text, prefix, "", -1)
				for _, command := range commands {
					if m, _ := regexp.MatchString(command.Pattern.String(), message.Text); m {
						fmt.Printf("-> Command: %s\n", message.Text)
						command.Handler(command, message)
					} else if m, _ := regexp.MatchString("^help", message.Text); m {
						fmt.Printf("-> Command: %s\n", message.Text)
						message.Text = generateHelpOutput()
						Respond(message)
					}
				}
			}(commands, message)
		}
	}
}

// AddCommand function adds a new command into the commands list.
func AddCommand(pattern string, name string, description string, handler handler) {
	commands = append(commands, Command{Pattern: regexp.MustCompile(pattern), Name: name, Description: description, Handler: handler})
}

// Respond function sends a message back into the websocket stream.
func Respond(message Message) error {
	return websocket.JSON.Send(WebsocketStream, message)
}

// Opens an Instant Messaging window with a user
func OpenIM(channel Channel) (IMJsonResponse, error) {
	var jsonResponse IMJsonResponse

	data := url.Values{}
	data.Set("token", Token)
	data.Add("user", channel.User)

	url := "https://slack.com/api/im.open"

	response, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	check_error(err)

	if response.StatusCode != 200 {
		return jsonResponse, fmt.Errorf("Unable to open an instant message. Code status: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	check_error(err)

	err = json.Unmarshal(body, &jsonResponse)
	check_error(err)

	fmt.Printf("%s", jsonResponse.Error)

	return jsonResponse, nil
}

// Opens a multi-pluriparty instant message window
func OpenMPIM(instantMessage MPInstantMessage) (MPInstantMessageJsonResponse, error) {
	var jsonResponse MPInstantMessageJsonResponse

	data := url.Values{}
	data.Set("token", Token)
	data.Add("users", instantMessage.Users)

	url := "https://slack.com/api/mpim.open"

	response, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	check_error(err)

	if response.StatusCode != 200 {
		return jsonResponse, fmt.Errorf("Unable to open a grouped conversation. Code status: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	check_error(err)

	err = json.Unmarshal(body, &jsonResponse)
	check_error(err)

	fmt.Printf("%s", jsonResponse.Error)

	return jsonResponse, nil
}

// Posts a message using the Slack API (not the Stream one).
func PostMessage(message Message) {
	data := url.Values{}
	data.Set("token", Token)
	data.Add("channel", message.Channel)
	data.Add("text", message.Text)
	data.Add("as_user", strconv.FormatBool(message.AsUser))

	url := "https://slack.com/api/chat.postMessage"

	_, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	check_error(err)
}

// List channels
func ListChannels() (ChannelsJsonResponse, error) {
	var jsonResponse ChannelsJsonResponse

	url := fmt.Sprintf("https://slack.com/api/channels.list?token=%s", Token)

	response, err := http.Get(url)
	check_error(err)

	if response.StatusCode != 200 {
		return jsonResponse, fmt.Errorf("Unable to retrieve channels list. Code status: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	check_error(err)

	err = json.Unmarshal(body, &jsonResponse)
	check_error(err)

	return jsonResponse, nil
}

// List users
func ListUsers() (UsersJsonResponse, error) {
	var jsonResponse UsersJsonResponse

	url := fmt.Sprintf("https://slack.com/api/users.list?token=%s", Token)

	response, err := http.Get(url)
	check_error(err)

	if response.StatusCode != 200 {
		return jsonResponse, fmt.Errorf("Unable to retrieve users list. Code status: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	check_error(err)

	err = json.Unmarshal(body, &jsonResponse)
	check_error(err)

	return jsonResponse, nil
}
