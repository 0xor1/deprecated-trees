package mail

import (
	"fmt"
	"github.com/0xor1/panic"
	sp "github.com/SparkPost/gosparkpost"
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

func NewSparkPostClient(from, apiKey string) Client {
	spClient := &sp.Client{}
	panic.IfNotNil(spClient.Init(&sp.Config{
		BaseUrl:    "https://api.eu.sparkpost.com",
		ApiKey:     apiKey,
		ApiVersion: 1,
	}))
	return &sparkPostClient{
		from:     from,
		spClient: spClient,
	}
}

type sparkPostClient struct {
	from     string
	spClient *sp.Client
}

func (c *sparkPostClient) Send(sendTo []string, content string) {
	f := false
	_, _, e := c.spClient.Send(&sp.Transmission{
		Options: &sp.TxOptions{
			TmplOptions: sp.TmplOptions{
				OpenTracking:  &f,
				ClickTracking: &f,
			},
		},
		Recipients: sendTo,
		Content: sp.Content{
			HTML:    content,
			From:    c.from,
			Subject: "project-trees.com registration",
		},
	})
	panic.IfNotNil(e)
}
