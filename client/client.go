package client

import (
	"context"
	"fmt"

	api "dedb"
)

type ClientConfig struct {
	Server        string
	Streams       []string
	ConsumerGroup string
	EventChannel  chan<- *api.Event
	ErrorChannel  chan<- error
}

type Client struct {
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	if config.Server == "" {
		return nil, fmt.Errorf("Server config entry required")
	}
	if len(config.Streams) == 0 {
		return nil, fmt.Errorf("Streams config entry required")
	}
	if config.ConsumerGroup == "" {
		return nil, fmt.Errorf("ConsumerGroup config entry required")
	}
	if len(domains) == 0 {
		return nil, fmt.Errorf("Streams config entry needs at least one domain to subscribe to")
	}

	c := &Client{config: config}
	return c, nil
}

func (c *Client) Save(ctx context.Context, request *api.SaveRequest) (*api.SaveResponse, error) {
	return c.server.Save(ctx, request)
}

func (c *Client) GetDomain(ctx context.Context, request *api.GetDomainRequest) (*api.GetResponse, error) {
	return c.server.GetDomain(ctx, request)
}

func (c *Client) GetDomainIds(ctx context.Context, request *api.GetDomainIdsRequest) (*api.GetDomainIdsResponse, error) {
	return c.server.GetDomainIds(ctx, request)
}

func (c *Client) Connect(ctx context.Context) error {
	return nil
}

func (c *Client) Close() {
}
