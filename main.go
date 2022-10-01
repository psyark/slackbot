package slackbot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func GetHandler(opt *GetHandlerOption) func(http.ResponseWriter, *http.Request) {
	return opt.handleRequest
}

type GetHandlerOption struct {
	Registry      *Registry
	AppHomeOpened AppHomeOpenedHandler
	Message       MessageHandler
	Error         ErrorHandler
}

func (h *GetHandlerOption) handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		defer func() {
			if err := recover(); err != nil {
				if h.Error != nil {
					h.Error(&ErrorHandlerArgs{
						ResponseWriter: w,
						Request:        r,
						Err:            fmt.Errorf("panic: %#v", err),
					})
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%#v", err)
				}
			}
		}()

		if err := h.handlePostRequest(w, r); err != nil {
			if h.Error != nil {
				h.Error(&ErrorHandlerArgs{
					ResponseWriter: w,
					Request:        r,
					Err:            err,
				})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%#v", err)
			}
		}
	}
}

func (r *GetHandlerOption) handlePostRequest(rw http.ResponseWriter, req *http.Request) error {
	payload, err := r.getPayload(req)
	if err != nil {
		return fmt.Errorf("getPayload(%#v): %#v", req, err)
	}

	// TODO:
	// if err := r.verifyHeader(req.Header); err != nil {
	// 	return err
	// }

	event, err := slackevents.ParseEvent(payload, slackevents.OptionNoVerifyToken())
	if err != nil {
		return fmt.Errorf("slackevents.ParseEvent(%#v): %w", payload, err)
	}

	switch event.Type {
	case slackevents.URLVerification:
		if uve, ok := event.Data.(*slackevents.EventsAPIURLVerificationEvent); ok {
			return r.verifyURL(rw, uve)
		}
		return fmt.Errorf("event.Data is not *EventsAPIURLVerificationEvent: %#v", event.Data)

	case slackevents.CallbackEvent:
		if err := r.handleCallback(req, &event); err != nil {
			return err
		}
		return nil

	case string(slack.InteractionTypeBlockActions):
		callback := slack.InteractionCallback{}
		if err := json.Unmarshal(payload, &callback); err != nil {
			return fmt.Errorf("json.Unmarshal(%#v): %w", payload, err)
		}

		for _, action := range callback.ActionCallback.BlockActions {
			if handler, ok := r.Registry.blockAction[action.ActionID]; ok {
				args := &BlockActionHandlerArgs{
					Request:             req,
					InteractionCallback: &callback,
					BlockAction:         action,
				}
				if err := handler(args); err != nil {
					return fmt.Errorf("blockActions: %w", err)
				}
			} else {
				return fmt.Errorf("unknown actionID: %v", action.ActionID)
			}
		}
		return nil

	case string(slack.InteractionTypeViewSubmission):
		callback := slack.InteractionCallback{}
		if err := json.Unmarshal(payload, &callback); err != nil {
			return fmt.Errorf("json.Unmarshal(%#v): %w", payload, err)
		}
		if handler, ok := r.Registry.viewSubmission[callback.View.CallbackID]; ok {
			args := &ViewSubmissionHandlerArgs{
				Request:             req,
				InteractionCallback: &callback,
			}
			res, err := handler(args)
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
		return fmt.Errorf("unknown type: %#v", event.Type)
	}
}

func (h *GetHandlerOption) handleCallback(req *http.Request, event *slackevents.EventsAPIEvent) error {
	switch innerEvent := event.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		if h.AppHomeOpened != nil {
			return h.AppHomeOpened(&AppHomeOpenedHandlerArgs{Request: req, AppHomeOpenedEvent: innerEvent})
		}
		return nil
	case *slackevents.MessageEvent:
		if h.Message != nil {
			return h.Message(&MessageHandlerArgs{Request: req, MessageEvent: innerEvent})
		}
		return nil

	default:
		return fmt.Errorf("unknown type: %#v/%#v", event.Type, event.InnerEvent.Type)
	}
}

func (h *GetHandlerOption) getPayload(req *http.Request) ([]byte, error) {
	switch req.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		return []byte(req.Form.Get("payload")), nil
	case "application/json":
		return io.ReadAll(req.Body)
	default:
		return nil, fmt.Errorf("unsupported content-type: %v", req.Header.Get("Content-Type"))
	}
}

func (h *GetHandlerOption) verifyURL(rw http.ResponseWriter, uvEvent *slackevents.EventsAPIURLVerificationEvent) error {
	rw.Header().Set("Content-Type", "text/plain")
	_, err := rw.Write([]byte(uvEvent.Challenge))
	return err
}
