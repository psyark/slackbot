package slackbot

import "github.com/GoogleCloudPlatform/functions-framework-go/functions"

func RegisterHandler_WillNotWork(name string, opt *GetHandlerOption) {
	functions.HTTP(name, GetHandler(opt))
}
