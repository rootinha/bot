package main

import (
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Slack struct {
	client *slack.Client
	rtm    *slack.RTM
	userId string
	user   string
}

func NewSlack(token, user, userId string) *Slack {
	s := &Slack{
		client: slack.New(token),
		user:   user,
		userId: userId,
	}

	s.rtm = s.client.NewRTM()

	logrus.Info("creating a new slack client...")

	return s
}

func (s *Slack) StartRTM(r *Rootinha) error {

	go s.rtm.ManageConnection()

	resp := make(chan *ConversationResponse)
	go s.sendMessage(resp)

	for {
		select {
		case msg := <-s.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:

				logrus.WithField("conversation", ev).Debug("an event was received")

				c, err := r.CreateConversation(ev.Text, ev.Channel)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"text":    ev.Text,
						"channel": ev.Channel,
					}).Warn(err)

					_, err := s.rtm.PostEphemeral(ev.Channel, ev.User,
						slack.MsgOptionText(err.Error(), false),
					)

					if err != nil {
						logrus.WithFields(logrus.Fields{
							"err":  err,
							"user": ev.User,
						}).Error("could not send an error message to user")
					}

					continue
				}

				if c.IsBotUserMentioned(s.userId) {
					logrus.WithFields(logrus.Fields{
						"uuid":    c.ID,
						"message": c.Message,
						"channel": c.Channel,
					}).Info("the message event has been accepted... starting conversation")

					s.rtm.SendMessage(s.rtm.NewTypingMessage(ev.Channel))

					go c.Start(resp)
				}

			case *slack.RTMError:
				//TODO check if the rtm is able to not abort the process, just logging
				logrus.Error(ev.Error())
			case *slack.InvalidAuthEvent:
				return errors.New("invalid credentials")
			default:
				// Take no action
			}
		}
	}
}

func (s *Slack) sendMessage(resp chan *ConversationResponse) {
	for r := range resp {

		logrus.WithFields(logrus.Fields{
			"uuid":    r.ParentID,
			"message": r.Text,
			"channel": r.Channel,
		}).Info("sending a reply to the user")

		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(r.Text, r.Channel))
	}
}
