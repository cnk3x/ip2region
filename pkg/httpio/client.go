package httpio

import (
	"net/http"
)

type ClientOption func(*http.Client)

func BuildClient(opts ...ClientOption) (*http.Client, error) {
	var client = &http.Client{}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}
