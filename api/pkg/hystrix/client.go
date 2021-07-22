package hystrix

import (
	"fmt"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
)

type Doer interface {
	Do(request *http.Request) (*http.Response, error)
}

type Client struct {
	client      Doer
	commandName string
}

func NewClient(client Doer, circuitConfig *hystrix.CommandConfig, commandName string) *Client {
	config := &hystrix.CommandConfig{}
	if circuitConfig != nil {
		config = circuitConfig
	}

	hystrix.ConfigureCommand(commandName, *config)
	return &Client{
		client:      client,
		commandName: commandName,
	}
}

func (cl *Client) Do(request *http.Request) (*http.Response, error) {
	var response *http.Response
	var err error

	err = hystrix.Do(cl.commandName, func() error {
		response, err = cl.client.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("got 5xx response code: %d", response.StatusCode)
		}
		return nil
	}, nil)
	return response, err
}
