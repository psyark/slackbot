package slackbot

type HandlerRegistry interface {
	Child(name string) HandlerRegistry
	GetActionID(name string, handler BlockActionHandler) string
	GetCallbackID(name string, handler ViewSubmissionHandler) string
}

var _ HandlerRegistry = &handlerRegistry{}

type handlerRegistry struct {
	namespace      string
	blockAction    map[string]BlockActionHandler
	viewSubmission map[string]ViewSubmissionHandler
}

func newHandlerRegistry() *handlerRegistry {
	return &handlerRegistry{
		blockAction:    map[string]BlockActionHandler{},
		viewSubmission: map[string]ViewSubmissionHandler{},
	}
}

func (r *handlerRegistry) resolve(name string) string {
	if r.namespace == "" {
		return name
	} else {
		return r.namespace + "." + name
	}
}

func (r *handlerRegistry) Child(name string) HandlerRegistry {
	return &handlerRegistry{
		namespace:      r.resolve(name),
		blockAction:    r.blockAction,
		viewSubmission: r.viewSubmission,
	}
}

func (r *handlerRegistry) GetActionID(name string, handler BlockActionHandler) string {
	id := r.resolve(name)
	if _, ok := r.blockAction[id]; ok {
		panic(id)
	}
	r.blockAction[id] = handler
	return id
}

func (r *handlerRegistry) GetCallbackID(name string, handler ViewSubmissionHandler) string {
	id := r.resolve(name)
	if _, ok := r.viewSubmission[id]; ok {
		panic(id)
	}
	r.viewSubmission[id] = handler
	return id
}
