package slackbot

type Registry struct {
	namespace      string
	blockAction    map[string]BlockActionHandler
	viewSubmission map[string]ViewSubmissionHandler
}

func NewRegistry() *Registry {
	return &Registry{
		blockAction:    map[string]BlockActionHandler{},
		viewSubmission: map[string]ViewSubmissionHandler{},
	}
}

func (r *Registry) resolve(name string) string {
	if r.namespace == "" {
		return name
	} else {
		return r.namespace + "." + name
	}
}

func (r *Registry) Child(name string) *Registry {
	return &Registry{
		namespace:      r.resolve(name),
		blockAction:    r.blockAction,
		viewSubmission: r.viewSubmission,
	}
}

func (r *Registry) GetActionID(name string, handler BlockActionHandler) string {
	id := r.resolve(name)
	if _, ok := r.blockAction[id]; ok {
		panic(id)
	}
	r.blockAction[id] = handler
	return id
}

func (r *Registry) GetCallbackID(name string, handler ViewSubmissionHandler) string {
	id := r.resolve(name)
	if _, ok := r.viewSubmission[id]; ok {
		panic(id)
	}
	r.viewSubmission[id] = handler
	return id
}
