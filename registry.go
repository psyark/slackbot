package slackbot

type HandlerRegistry struct {
	namespace      string
	blockAction    map[string]BlockActionHandler
	viewSubmission map[string]ViewSubmissionHandler
}

func NewRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		blockAction:    map[string]BlockActionHandler{},
		viewSubmission: map[string]ViewSubmissionHandler{},
	}
}

func (r *HandlerRegistry) resolve(name string) string {
	if r.namespace == "" {
		return name
	} else {
		return r.namespace + "." + name
	}
}

func (r *HandlerRegistry) Child(name string) *HandlerRegistry {
	return &HandlerRegistry{
		namespace:      r.resolve(name),
		blockAction:    r.blockAction,
		viewSubmission: r.viewSubmission,
	}
}

func (r *HandlerRegistry) GetActionID(name string, handler BlockActionHandler) string {
	id := r.resolve(name)
	if _, ok := r.blockAction[id]; ok {
		panic(id)
	}
	r.blockAction[id] = handler
	return id
}

func (r *HandlerRegistry) GetCallbackID(name string, handler ViewSubmissionHandler) string {
	id := r.resolve(name)
	if _, ok := r.viewSubmission[id]; ok {
		panic(id)
	}
	r.viewSubmission[id] = handler
	return id
}
