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

var _ HandlerRegistry = &Router{}

type Router struct {
	*handlerRegistry
	AppHomeOpened func(req *http.Request, event *slackevents.AppHomeOpenedEvent) error
	Message       func(req *http.Request, event *slackevents.MessageEvent) error
	Error         func(w http.ResponseWriter, r *http.Request, err error)
}

func New() *Router {
	return &Router{handlerRegistry: newHandlerRegistry()}
}

func (h *Router) Route(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		defer func() {
			if err := recover(); err != nil {
				if h.Error != nil {
					h.Error(w, r, fmt.Errorf("panic: %#v", err))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%#v", err)
				}
			}
		}()

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

func (r *Router) handlePostRequest(rw http.ResponseWriter, req *http.Request) error {
	payload, err := r.getPayload(req)
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
			return r.verifyURL(rw, uve)
		}
		return fmt.Errorf("event.Data is not *EventsAPIURLVerificationEvent: %#v", event.Data)

	case slackevents.CallbackEvent:
		if err := r.handleCallback(req, &event); err != nil {
			return xerrors.Errorf("handleCallback: %#v", err)
		}
		return nil

	case string(slack.InteractionTypeBlockActions):
		callback := slack.InteractionCallback{}
		if err := json.Unmarshal(payload, &callback); err != nil {
			return xerrors.Errorf("json.Unmarshal(%#v): %#v", payload, err)
		}

		for _, action := range callback.ActionCallback.BlockActions {
			if handler, ok := r.handlerRegistry.blockAction[action.ActionID]; ok {
				if err := handler(&callback, action); err != nil {
					return xerrors.Errorf("blockActions: %#v", err)
				}
			} else {
				return fmt.Errorf("unknown actionID: %v", action.ActionID)
			}
		}
		return nil

	case string(slack.InteractionTypeViewSubmission):
		callback := slack.InteractionCallback{}
		if err := json.Unmarshal(payload, &callback); err != nil {
			return xerrors.Errorf("json.Unmarshal(%#v): %#v", payload, err)
		}
		if handler, ok := r.handlerRegistry.viewSubmission[callback.View.CallbackID]; ok {
			res, err := handler(&callback)
			if err != nil {
				return err
			}
			if res != nil {
				rw.Header().Add("Content-Type", "application/json")
				rw.WriteHeader(http.StatusOK)
				return json.NewEncoder(rw).Encode(res)
			}
			return nil
		} else {
			return fmt.Errorf("unknown callbackID: %v", callback.View.CallbackID)
		}

	default:
		return xerrors.Errorf("unknown type: %#v", event.Type)
	}
}

func (h *Router) handleCallback(req *http.Request, event *slackevents.EventsAPIEvent) error {
	switch innerEvent := event.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		if h.AppHomeOpened != nil {
			return h.AppHomeOpened(req, innerEvent)
		}
		return nil
	case *slackevents.MessageEvent:
		if h.Message != nil {
			return h.Message(req, innerEvent)
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
