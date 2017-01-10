# Slackbot

This is a Slack Robot written in Go.

[![GoDoc](https://godoc.org/github.com/eko/slackbot?status.png)](https://godoc.org/github.com/eko/slackbot)
[![Build Status](https://travis-ci.org/eko/slackbot.png?branch=master)](https://travis-ci.org/eko/slackbot)


## Robot creation

1. Go on the following uri to declare your new bot: https://<team>.slack.com/services/new/bot

2. Retrieve the given token.

## Installation

```bash
$ go get -u github.com/eko/slackbot
```

## Run the robot

```bash
$ go run app.go
Bot is ready, hit ^C to exit.
-> Command: hello dude
```

## A robot example application

This example application answers to the following command:

* @yourbotname hello <name>: Renders "hello <name>!",

```go
package main

import (
    "github.com/eko/slackbot"
    "fmt"
)

func main() {
	slackbot.Token = "<your-bot-token>"
	slackbot.Init()

    slackbot.AddCommand("^hello (.*)", func(command slackbot.Command, message slackbot.Message) {
		name := command.Pattern.FindStringSubmatch(message.Text)[1]
		message.Text = string(fmt.Sprintf("hello, %s!", name))

		slackbot.Send(message)
	})

    slackbot.Stream()
}
```
