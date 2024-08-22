package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asecurityteam/logevent/v2"
)

type instanceStoreTransport struct {
	instance logevent.Logger
}

func (t *instanceStoreTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.instance = FromRequest(r)
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
	}, nil
}

func TestNewTransport(t *testing.T) {
	logger := logevent.New(logevent.Config{Level: "INFO", Output: ioutil.Discard})
	wrapped := &instanceStoreTransport{}
	transport := NewTransport(logger)(wrapped)
	_, _ = transport.RoundTrip(httptest.NewRequest(http.MethodGet, "/", nil))
	assert.NotNil(t, wrapped.instance)
}
