// Slackbot - A Slack robot library written in Go
//
// Author: Vincent Composieux <vincent.composieux@gmail.com>

package slackbot

import (
	"testing"
)

// Tests pre-defined and variables values
func TestVariablesValues(t *testing.T) {
	Token = "test-token"

	if Token != "test-token" {
		t.Error("Should have the Token defined")
	}
}

// Tests the command addition
func TestAddCommand(t *testing.T) {
	if len(commands) != 0 {
		t.Error("Should have 0 commands by default")
	}

	AddCommand("test-a", "a", "test a", func(command Command, message Message) {})
	AddCommand("test-b", "b", "test b", func(command Command, message Message) {})

	if len(commands) != 2 {
		t.Error("Should have 2 commands added")
	}
}
