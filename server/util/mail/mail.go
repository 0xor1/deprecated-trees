package mail

import (
	"fmt"
)

type Client interface {
	Send(sendTo []string, content string)
}

func NewLocalClient() Client {
	return &localClient{}
}

type localClient struct{}

func (c *localClient) Send(sendTo []string, content string) {
	fmt.Println(sendTo, content)
}
