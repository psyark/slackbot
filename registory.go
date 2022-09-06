package slackbot

type HandlerRegistory interface {
	Child(name string) HandlerRegistory
	GetActionID(name string, handler BlockActionHandler) string
	GetCallbackID(name string, handler ViewSubmissionHandler) string
}

var _ HandlerRegistory = &handlerRegistory{}

type handlerRegistory struct {
	namespace      string
	blockAction    map[string]BlockActionHandler
	viewSubmission map[string]ViewSubmissionHandler
}

func newHandlerRegistory() *handlerRegistory {
	return &handlerRegistory{
		blockAction:    map[string]BlockActionHandler{},
		viewSubmission: map[string]ViewSubmissionHandler{},
	}
}

func (r *handlerRegistory) resolve(name string) string {
	if r.namespace == "" {
		return name
	} else {
		return r.namespace + "." + name
	}
}

func (r *handlerRegistory) Child(name string) HandlerRegistory {
	return &handlerRegistory{
		namespace:      r.resolve(name),
		blockAction:    r.blockAction,
		viewSubmission: r.viewSubmission,
	}
}

func (r *handlerRegistory) GetActionID(name string, handler BlockActionHandler) string {
	id := r.resolve(name)
	if _, ok := r.blockAction[id]; ok {
		panic(id)
	}
	r.blockAction[id] = handler
	return id
}

func (r *handlerRegistory) GetCallbackID(name string, handler ViewSubmissionHandler) string {
	id := r.resolve(name)
	if _, ok := r.viewSubmission[id]; ok {
		panic(id)
	}
	r.viewSubmission[id] = handler
	return id
}
