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

type Router struct {
	AppHomeOpened func(req *http.Request, event *slackevents.AppHomeOpenedEvent)
	Message       func(req *http.Request, event *slackevents.MessageEvent)
	BlockActions  func(req *http.Request, cb *slack.InteractionCallback)
	Error         func(w http.ResponseWriter, r *http.Request, err error)
}

func (h *Router) Route(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if err := h.handlePostRequest(w, r); err != nil {
			if h.Error != nil {
				h.Error(w, r, err)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%#v", err)
			}
		}
	}
}

func (h *Router) handlePostRequest(rw http.ResponseWriter, req *http.Request) error {
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
		if uve, ok := event.Data.(*slackevents.EventsAPIURLVerificationEvent); ok {
			return h.verifyURL(rw, uve)
		}
		return fmt.Errorf("event.Data is not *EventsAPIURLVerificationEvent: %#v", event.Data)

	case slackevents.CallbackEvent:
		if err := h.handleCallback(req, &event); err != nil {
			return xerrors.Errorf("handleCallback: %#v", err)
		}
		return nil

	case string(slack.InteractionTypeBlockActions):
		if h.BlockActions != nil {
			intCb := slack.InteractionCallback{}
			if err := json.Unmarshal(payload, &intCb); err != nil {
				return xerrors.Errorf("json.Unmarshal(%#v): %#v", payload, err)
			}
			h.BlockActions(req, &intCb)
		}
		return nil

	default:
		return xerrors.Errorf("unknown type: %#v", event.Type)
	}
}

func (h *Router) handleCallback(req *http.Request, event *slackevents.EventsAPIEvent) error {
	switch innerEvent := event.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		if h.AppHomeOpened != nil {
			h.AppHomeOpened(req, innerEvent)
		}
		return nil
	case *slackevents.MessageEvent:
		if h.Message != nil {
			h.Message(req, innerEvent)
		}
		return nil

	default:
		return xerrors.Errorf("unknown type: %#v/%#v", event.Type, event.InnerEvent.Type)
	}
}

func (h *Router) getPayload(req *http.Request) ([]byte, error) {
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

func (h *Router) verifyURL(rw http.ResponseWriter, uvEvent *slackevents.EventsAPIURLVerificationEvent) error {
	rw.Header().Set("Content-Type", "text/plain")
	_, err := rw.Write([]byte(uvEvent.Challenge))
	return err
}
