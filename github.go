package main

import (
	"context"
	"strings"

	gh "github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type GH struct {
	client *gh.Client
	url    string
}

func NewGitHubPlugin(url, token string) (*GH, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	c, err := gh.NewEnterpriseClient(url, url, tc)
	if err != nil {
		return nil, errors.Wrap(err, "attempt to connect to github failed")
	}

	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Info("creating a new github client...")

	return &GH{
		client: c,
		url:    url,
	}, nil
}

func (g *GH) ListActions() map[string]ConversationFn {
	return map[string]ConversationFn{
		"list-prs": g.ListPullRequests,
	}
}

func (g *GH) ListPullRequests(ctx context.Context, c *Conversation) *ConversationResponse {

	logrus.WithFields(logrus.Fields{
		"owner": c.Params["org"],
		"repo":  c.Params["repository"],
		"state": c.Params["state"],
	}).Debug("listing pull requests")

	prs, _, err := g.client.PullRequests.List(ctx, c.Params["org"], c.Params["repository"], &gh.PullRequestListOptions{State: c.Params["state"]})
	if err != nil {
		logrus.WithField("conversation", c).Error(err)

		return &ConversationResponse{
			Channel:  c.Channel,
			Text:     err.Error(),
			ParentID: c.ID,
		}
	}

	writer, err := NewTemplateWriter(c.ResponseTmpl)
	if err != nil {
		return &ConversationResponse{
			Channel:  c.Channel,
			Text:     err.Error(),
			ParentID: c.ID,
		}
	}

	var resp []string
	for _, pr := range prs {
		resp = append(resp, writer.Write(pr))
	}

	return &ConversationResponse{
		Channel:  c.Channel,
		Text:     strings.Join(resp, "\n"),
		ParentID: c.ID,
	}
}
