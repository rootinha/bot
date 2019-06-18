package main

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"text/template"

	"github.com/nlopes/slack"
	"github.com/rootinha/bot/plugins/github"
	"github.com/sirupsen/logrus"
)

type BotConfig struct {
	Slack    *Slack
	GitHub   *GitHub `yaml:"github"`
	Entities []*Entity
	Intents  []*Intent
}

type Slack struct {
	Token string
	User  string
}

type GitHub struct {
	URL    string `yaml:"url"`
	APIURL string `yaml:"apiurl"`
	Token  string `yaml:"token"`
}

type Entity struct {
	Name   string
	Format string
	Values []string
}

type Intent struct {
	Expression []string
	Regex      []*regexp.Regexp
	Plugin     *Plugin
	Response   *Response
}

type Response struct {
	Template string
}

func (c *BotConfig) Compile() {
	for _, it := range c.Intents {
		it.Regex = make([]*regexp.Regexp, len(it.Expression))
		for i, exp := range it.Expression {
			it.Regex[i] = regexp.MustCompile(exp)
		}
	}
}

type Rootinha struct {
	Config *BotConfig
}

var (
	reMention, _ = regexp.Compile(`^(?P<user>\<@\w+\>)(?P<message>.*)$`)
)

type Plugin struct{}

func New() *Rootinha {
	return &Rootinha{
		Config: &BotConfig{},
	}
}

func (r *Rootinha) Start() error {

	api := slack.New(r.Config.Slack.Token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	cli, err := github.New(r.Config.GitHub.APIURL, r.Config.GitHub.Token)
	if err != nil {
		return err
	}

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:

				ok, mention, text := findByMention(ev.Text, r.Config.Slack.User)

				logrus.Info("found mention: ", ok)
				logrus.Info("user: ", mention)
				logrus.Info("message: ", text)
				logrus.Info("message: ", ev.Text)

				if ok {

					rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))

					for _, it := range r.Config.Intents {
						tmp, err := template.New("tmpl").Parse(it.Response.Template)
						if err != nil {
							logrus.Error(err)
						}

						for _, re := range it.Regex {
							res := extractEntities(re, strings.TrimSpace(text))
							if len(res) == 0 {
								continue
							}

							prs, err := cli.ListPullRequests(context.Background(), "clarobr", res["repository"], res["state"])
							if err != nil {
								logrus.Error(err)
							}

							for _, pr := range prs {
								var tpl bytes.Buffer
								err = tmp.Execute(&tpl, pr)
								if err != nil {
									logrus.Error(err)
								}
								rtm.SendMessage(rtm.NewOutgoingMessage(tpl.String(), ev.Channel))
							}

						}
					}
				}

			case *slack.RTMError:
				logrus.Error(ev.Error())
			case *slack.InvalidAuthEvent:
				logrus.Error("invalid credentials")
			default:
				// Take no action
			}
		}
	}

	return nil
}

func extractEntities(re *regexp.Regexp, text string) map[string]string {
	match := re.FindStringSubmatch(text)

	if len(match) == 0 {
		return make(map[string]string)
	}

	result := make(map[string]string)

	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			if match[i] != "" {
				result[name] = match[i]
			}
		}
	}

	return result
}

func findByMention(msg, user string) (bool, string, string) {
	match := reMention.FindStringSubmatch(msg)

	if len(match) >= 3 {
		return true, match[1], strings.TrimSpace(match[2])
	}

	return false, "", ""
}
