package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceRepo(t *testing.T) {
	// setup
	cases := []struct {
		name   string
		config Config
		err    error
	}{
		{
			name: "Repo not supported",
			config: Config{
				RepoImpl: "test",
			},
			err: fmt.Errorf("repository test not supported"),
		},
		{
			name: "Broker not supported",
			config: Config{
				RepoImpl:   "redis",
				BrokerImpl: "test",
				RedisDbConfig: RedisDbConfig{
					Server: "test_server",
				},
			},
			err: fmt.Errorf("broker test not supported"),
		},
	}

	// when / then
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := Service{}
			err := svc.Start(tc.config)
			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}
		})
	}
}
