package slackbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"golang.org/x/xerrors"
)

type Handler interface {
	OnCallbackMessage(req *http.Request, event *slackevents.MessageEvent)
	OnBlockActions(req *http.Request, cb *slack.InteractionCallback)
}

type BaseHandler struct {
	handler Handler
}

func New(handler Handler) *BaseHandler {
	return &BaseHandler{handler: handler}
}

func (h *BaseHandler) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if err := h.handlePostRequest(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%#v", err)
		}
	}
}

func (h *BaseHandler) handlePostRequest(rw http.ResponseWriter, req *http.Request) error {
	payload, err := h.getPayload(req)
	if err != nil {
		return xerrors.Errorf("getPayload(%#v): %#v", req, err)
	}

	// TODO:
	// if err := r.verifyHeader(req.Header); err != nil {
	// 	return err
	// }

	event, err := slackevents.ParseEvent(payload, slackevents.OptionNoVerifyToken())
	if err != nil {
		return xerrors.Errorf("slackevents.ParseEvent(%#v): %#v", payload, err)
	}

	switch event.Type {
	case slackevents.URLVerification:
		if uve, ok := event.Data.(slackevents.EventsAPIURLVerificationEvent); ok {
			return h.verifyURL(rw, uve)
		}
		return fmt.Errorf("event.Data is not EventsAPIURLVerificationEvent: %#v", event.Data)

	case slackevents.CallbackEvent:
		if err := h.handleCallback(req, &event); err != nil {
			return xerrors.Errorf("handleCallback: %#v", err)
		}
		return nil

	case string(slack.InteractionTypeBlockActions):
		intCb := slack.InteractionCallback{}
		if err := json.Unmarshal(payload, &intCb); err != nil {
			return xerrors.Errorf("json.Unmarshal(%#v): %#v", payload, err)
		}

		h.handler.OnBlockActions(req, &intCb)
		return nil

	default:
		return xerrors.Errorf("unknown type: %#v", event.Type)
	}
}

func (h *BaseHandler) handleCallback(req *http.Request, event *slackevents.EventsAPIEvent) error {
	switch innerEvent := event.InnerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		h.handler.OnCallbackMessage(req, innerEvent)
		return nil

	default:
		return xerrors.Errorf("unknown type: %#v/%#v", event.Type, event.InnerEvent.Type)
	}
}

func (h *BaseHandler) getPayload(req *http.Request) ([]byte, error) {
	switch req.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		return []byte(req.Form.Get("payload")), nil
	case "application/json":
		return ioutil.ReadAll(req.Body)
	default:
		return nil, xerrors.Errorf("unsupported content-type: %#v", req.Header.Get("Content-Type"))
	}
}

func (h *BaseHandler) verifyURL(rw http.ResponseWriter, uvEvent slackevents.EventsAPIURLVerificationEvent) error {
	rw.Header().Set("Content-Type", "text/plain")
	_, err := rw.Write([]byte(uvEvent.Challenge))
	return err
}
