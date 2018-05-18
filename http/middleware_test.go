package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/atlassian/logevent"
	"github.com/rs/xlog"
)

type fixtureHandler struct {
	instance     logevent.Logger
	xlogInstance xlog.Logger
}

func (h *fixtureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.instance = logevent.FromContext(r.Context())
	h.xlogInstance = xlog.FromContext(r.Context())
}

func TestMiddleware(t *testing.T) {
	var wrapped = &fixtureHandler{}
	var m, _ = NewMiddleware(
		OptionLevel("INFO"),
		OptionOutput(ioutil.Discard),
		OptionHumanReadable,
	)
	var h = m(wrapped)
	var w = httptest.NewRecorder()
	var r = httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(w, r)
	if wrapped.instance == nil {
		t.Error("did not find Logger in context")
	}
	if wrapped.xlogInstance == nil {
		t.Error("did not find xlog in context")
	}
}
