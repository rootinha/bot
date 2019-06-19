package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Rootinha struct {
	config  *BotConfig
	plugins map[string]Plugin
}

func New(c *BotConfig) (*Rootinha, error) {

	c.compileIntents()

	//TODO check if it is better to use a register or init package
	//Maybe the plugin should be responsible for creating your dependencies
	//and defining your name
	//plugin init func
	gh, err := NewGitHubPlugin(c.GitHub.APIURL, c.GitHub.Token)
	if err != nil {
		return nil, err
	}

	return &Rootinha{
		config: c,
		plugins: map[string]Plugin{
			"github": gh,
		},
	}, nil
}

func (r *Rootinha) Start() error {

	logrus.Info("starting the bot...")

	slack := NewSlack(r.config.Slack.Token, r.config.Slack.User, r.config.Slack.UserId)
	return slack.StartRTM(r)
}

func (r *Rootinha) CreateConversation(message, channel string) (*Conversation, error) {
	for _, it := range r.config.Intents {
		for _, re := range it.Regex {

			c := NewConversation(message, channel)
			params, ok := r.extractParams(re, strings.TrimSpace(c.Text))
			if !ok {
				continue
			}

			err := r.validateParams(params)
			if err != nil {
				return nil, err
			}

			err = mergo.Merge(&params, it.Plugin.Params)
			if err != nil {
				logrus.WithError(err).Error("could not merge the params")
			}

			c.Params = params
			c.ResponseTmpl = it.Response.Template

			p, ok := r.plugins[it.Plugin.Name]
			if !ok {
				return nil, errors.New(fmt.Sprintf("plugin not found [%s]", it.Plugin.Name))
			}

			c.Action, ok = p.ListActions()[it.Plugin.Action]
			if !ok {
				return nil, errors.New(fmt.Sprintf("action not found [%s]", it.Plugin.Action))
			}
			return c, nil
		}
	}

	return nil, errors.New("no intent has been found")
}

func (r *Rootinha) extractParams(re *regexp.Regexp, text string) (map[string]string, bool) {
	match := re.FindStringSubmatch(text)

	if len(match) == 0 {
		return nil, false
	}

	result := make(map[string]string)

	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			if match[i] != "" {
				result[name] = match[i]
			}
		}
	}

	return result, true
}

func (r *Rootinha) validateParams(params map[string]string) error {
	var notfound []string

	for _, e := range r.config.Entities {
		val, ok := params[e.Name]
		if ok {
			var found bool
			for _, ee := range e.Values {
				if ee == val {
					found = true
				}
			}
			if !found {
				notfound = append(notfound, e.Name)
			}
		}
	}

	if len(notfound) > 0 {
		return errors.New(fmt.Sprintf("invalid parameter %s", strings.Join(notfound, ",")))
	}

	return nil
}
