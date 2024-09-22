package httpio

import "net/http"

func BuildTransport() http.RoundTripper {
	return http.DefaultTransport
}
