package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/atlassian/logevent"
)

type fixtureHandler struct {
	instance logevent.Logger
}

func (h *fixtureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.instance = FromRequest(r)
}

func TestMiddleware(t *testing.T) {
	var wrapped = &fixtureHandler{}
	var logger = logevent.New(logevent.Config{Level: "INFO", Output: ioutil.Discard})
	var m = NewMiddleware(logger)
	var h = m(wrapped)
	var w = httptest.NewRecorder()
	var r = httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(w, r)
	if wrapped.instance == nil {
		t.Error("did not find Logger in context")
	}
}
