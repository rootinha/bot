package main

import "regexp"

type BotConfig struct {
	Slack    *SlackConfig
	GitHub   *GitHub `yaml:"github"`
	Entities []*Entity
	Intents  []*Intent
}

func (c *BotConfig) compileIntents() {
	for _, it := range c.Intents {
		it.Regex = make([]*regexp.Regexp, len(it.Expression))
		for i, exp := range it.Expression {
			it.Regex[i] = regexp.MustCompile(exp)
		}
	}
}

type SlackConfig struct {
	Token  string
	User   string
	UserId string
}

type GitHub struct {
	URL    string
	APIURL string `yaml:"apiurl"`
	Token  string
}

type Entity struct {
	Name   string
	Values []string
}

type Intent struct {
	Expression []string
	Regex      []*regexp.Regexp
	Plugin     *PluginConfig
	Response   *Response
}

type PluginConfig struct {
	Name   string
	Action string
	Params map[string]string
}

type Response struct {
	Template string
}
