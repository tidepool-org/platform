package highwater

import (
	"log"
)

type HighwaterMockClient struct{}

func NewMock() *HighwaterMockClient {
	return &HighwaterMockClient{}
}

func (client *HighwaterMockClient) PostServer(eventName, token string, params map[string]string) {

	if eventName == "" || token == "" {
		log.Panicf("missing required eventName[%s] token[%s] params[%v]", eventName, token, params)
	}

	return
}

func (client *HighwaterMockClient) PostThisUser(eventName, token string, params map[string]string) {

	if eventName == "" || token == "" {
		log.Panicf("missing required eventName[%s] token[%s]", eventName, token)
	}

	return
}

func (client *HighwaterMockClient) PostWithUser(userId, eventName, token string, params map[string]string) {
	if userId == "" || eventName == "" || token == "" {
		log.Panicf("missing required userId[%s] eventName[%s] token[%s]", userId, eventName, token)
	}

	return
}
