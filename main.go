package slackbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Handler interface {
	OnCallbackMessage(req *http.Request, event *slackevents.MessageEvent)
	OnBlockActions(req *http.Request, cb *slack.InteractionCallback)
}

type baseHandler struct {
	handler Handler
}

func RegisterHandler(handler Handler) {
	bh := baseHandler{handler: handler}
	functions.HTTP("main", bh.handleMain)
}

func (h baseHandler) handleMain(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePostRequest(w, r)
	}
}

func (h baseHandler) handlePostRequest(rw http.ResponseWriter, req *http.Request) error {
	payload, err := h.getPayload(req)
	if err != nil {
		return err
	}

	// TODO:
	// if err := r.verifyHeader(req.Header); err != nil {
	// 	return err
	// }

	event, err := slackevents.ParseEvent(payload, slackevents.OptionNoVerifyToken())
	if err != nil {
		return err
	}

	switch event.Type {
	case slackevents.URLVerification:
		return h.verifyURL(rw, event.Data.(slackevents.EventsAPIURLVerificationEvent))

	case slackevents.CallbackEvent:
		if err := h.handleCallback(req, &event); err != nil {
			return err
		}
		return nil

	case string(slack.InteractionTypeBlockActions):
		intCb := slack.InteractionCallback{}
		if err := json.Unmarshal(payload, &intCb); err != nil {
			return err
		}

		h.handler.OnBlockActions(req, &intCb)
		return nil

	default:
		return fmt.Errorf("unknown type: %v", event.Type)
	}
}

func (h baseHandler) handleCallback(req *http.Request, event *slackevents.EventsAPIEvent) error {
	switch innerEvent := event.InnerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		h.handler.OnCallbackMessage(req, innerEvent)
		return nil

	default:
		return fmt.Errorf("unknown type: %v/%v", event.Type, event.InnerEvent.Type)
	}
}

func (h baseHandler) getPayload(req *http.Request) ([]byte, error) {
	switch req.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		return []byte(req.Form.Get("payload")), nil
	case "application/json":
		return ioutil.ReadAll(req.Body)
	default:
		return nil, fmt.Errorf("unsupported content-type: %v", req.Header.Get("Content-Type"))
	}
}

func (h baseHandler) verifyURL(rw http.ResponseWriter, uvEvent slackevents.EventsAPIURLVerificationEvent) error {
	rw.Header().Set("Content-Type", "text/plain")
	_, err := rw.Write([]byte(uvEvent.Challenge))
	return err
}
