package slackbot

import "github.com/slack-go/slack"

type BlockActionHandler func(callback *slack.InteractionCallback, action *slack.BlockAction) error
type ViewSubmissionHandler func(callback *slack.InteractionCallback) (*slack.ViewSubmissionResponse, error)
