package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	reFirstMention, _ = regexp.Compile(`^(?P<user><@\w+>)(?P<message>.*)$`)
)

type ConversationFn func(context.Context, *Conversation) *ConversationResponse

type ConversationResponse struct {
	ParentID string
	Text     string
	Channel  string
}

type Conversation struct {
	ID           string
	Message      string
	Channel      string
	Params       map[string]string
	FirstMention string
	Text         string
	ResponseTmpl string
	Action       ConversationFn
}

func NewConversation(message string, channel string) *Conversation {
	c := &Conversation{
		ID:      uuid.NewV4().String(),
		Message: message,
		Channel: channel,
	}

	c.loadMentionAndText()

	return c
}

func (c *Conversation) Start(r chan *ConversationResponse) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	response := c.Action(ctx, c)
	r <- response
}

func (c *Conversation) String() string {
	return fmt.Sprintf("original: %s, text: %s, mention: %s, channel: %s",
		c.Message, c.Text, c.FirstMention, c.Channel)
}

func (c *Conversation) loadMentionAndText() {
	match := reFirstMention.FindStringSubmatch(c.Message)

	if len(match) >= 3 {
		c.FirstMention = match[1]
		c.Text = match[2]
	}
}

func (c *Conversation) IsBotUserMentioned(user string) bool {
	return strings.Contains(c.FirstMention, user)
}
