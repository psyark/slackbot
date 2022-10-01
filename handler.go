package slackbot

import (
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type AppHomeOpenedHandler func(args *AppHomeOpenedHandlerArgs) error
type AppHomeOpenedHandlerArgs struct {
	Request            *http.Request
	AppHomeOpenedEvent *slackevents.AppHomeOpenedEvent
}

type MessageHandler func(args *MessageHandlerArgs) error
type MessageHandlerArgs struct {
	Request      *http.Request
	MessageEvent *slackevents.MessageEvent
}

type ErrorHandler func(args *ErrorHandlerArgs)
type ErrorHandlerArgs struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Err            error
}

type BlockActionHandler func(args *BlockActionHandlerArgs) error
type BlockActionHandlerArgs struct {
	Request             *http.Request
	InteractionCallback *slack.InteractionCallback
	BlockAction         *slack.BlockAction
}

type ViewSubmissionHandler func(args *ViewSubmissionHandlerArgs) (*slack.ViewSubmissionResponse, error)
type ViewSubmissionHandlerArgs struct {
	Request             *http.Request
	InteractionCallback *slack.InteractionCallback
}
