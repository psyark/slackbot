package slackbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack"
)

func ExampleBot() {
	router := New()
	actionID := router.AddBlockActions("hoge", func(callback *slack.InteractionCallback, action *slack.BlockAction) error {
		fmt.Println("OK!")
		return nil
	})

	fmt.Println(actionID)

	router.Error = func(w http.ResponseWriter, r *http.Request, err error) {
		panic(err)
	}

	w := &responseWriter{}
	router.Route(w, createDummyRequest(actionID))

	// Output:
	// ba.hoge
	// OK!
}

func createDummyRequest(actionID string) *http.Request {
	callback := slack.InteractionCallback{
		Type: slack.InteractionTypeBlockActions,
		ActionCallback: slack.ActionCallbacks{
			BlockActions: []*slack.BlockAction{
				{
					ActionID: actionID,
				},
			},
		},
	}
	body, _ := json.Marshal(callback)
	return &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/json"}},
		Body:   io.NopCloser(bytes.NewBuffer(body)),
	}
}

type responseWriter struct {
}

func (rw *responseWriter) Header() http.Header {
	return http.Header{}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	fmt.Println(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	fmt.Println(data)
	return 0, nil
}
