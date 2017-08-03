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

// Tests that the check prefix flag works
func TestCheckPrefix(t *testing.T) {
	testCases := []struct {
		RequirePrefix bool
		input         string
		expectation   bool
	}{
		{true, "!foo", true},
		{false, "!foo", true},
		{true, "foo", false},
		{false, "foo", true},
	}
	for _, testCase := range testCases {
		prefix := "!"
		var message Message
		message.Text = testCase.input

		result := checkPrefix(message, prefix, testCase.RequirePrefix)

		if result != testCase.expectation {
			t.Fatalf("\nGot: %t\nExpected: %t\n--with--\n%v\n",
				result, testCase.expectation, testCase)
		}
	}
}
