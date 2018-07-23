package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/providers"
)

// Tests log middleware.
func TestLogMiddleware(t *testing.T) {
	in := []struct {
		url string
	}{
		{
			url: "/api/v1/test",
		},
		{
			url: "/pub/ping",
		},
	}

	nextCalled := false
	logCalled := false
	s := &GoHomeServer{
		Logger: mocks.FakeNewLogger(func(s string) {
			logCalled = true
		}),
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nextCalled = true
	})
	ts := httptest.NewServer(s.logMiddleware(handler))

	for _, v := range in {
		nextCalled = false
		logCalled = false
		var u bytes.Buffer
		u.WriteString(string(ts.URL))
		u.WriteString(v.url)

		_, err := http.Get(u.String())
		if err != nil {
			t.FailNow()
		}

		if !nextCalled {
			t.Error("Next not called " + v.url)
			t.Fail()
		}

		if !logCalled {
			t.Error("Expected log " + v.url)
			t.Fail()
		}
	}
}

// Tests authorization middleware.
func TestAuthMiddleware(t *testing.T) {
	prepareCidrs()
	in := []struct {
		url          string
		nextExpected bool
		security     providers.ISecurityProvider
		headers      map[string]string
	}{
		{
			url: "/api/v2/test/1",
			security:mocks.FakeNewSecurityProvider(true),
			headers:map[string]string{"X-Real-Ip" : "512.0.0.0"},
			nextExpected:false,
		},
		{
			url: "/api/v2/test/2",
			security:mocks.FakeNewSecurityProvider(true),
			headers:map[string]string{"X-Forwarded-For" : "245.0.0.0"},
			nextExpected:false,
		},
		{
			url: "/api/v2/test/3",
			security:mocks.FakeNewSecurityProvider(true),
			headers:map[string]string{"X-Forwarded-For" : "10.0.0.0", "X-Real-IP": "245.0.0.0"},
			nextExpected:false,
		},
		{
			url: "/api/v2/test/4",
			security:mocks.FakeNewSecurityProvider(true),
			headers:map[string]string{"X-Forwarded-For" : "10.0.0.1"},
			nextExpected:true,
		},
		{
			url: "/api/v2/test/5",
			security:mocks.FakeNewSecurityProvider(true),
			headers:map[string]string{},
			nextExpected:true,
		},
		{
			url: "/api/v1/test/6",
			security:mocks.FakeNewSecurityProvider(true),
			headers:map[string]string{},
			nextExpected:true,
		},
		{
			url: "/api/v1/test/7",
			security:mocks.FakeNewSecurityProvider(false),
			headers:map[string]string{},
			nextExpected:false,
		},
	}

	nextCalled := false
	s := &GoHomeServer{
		Logger: mocks.FakeNewLogger(nil),
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nextCalled = true
	})

	ts := httptest.NewServer(s.authMiddleware(handler))
	for _, v := range in {
		s.Settings = mocks.FakeNewSettingsWithUserStorage(v.security)

		nextCalled = false
		var u bytes.Buffer
		u.WriteString(string(ts.URL))
		u.WriteString(v.url)

		client := &http.Client{}
		req, _ := http.NewRequest("GET", u.String(), nil)

		for k, h := range v.headers {
			req.Header.Add(k, h)
		}

		_, err := client.Do(req)
		if err != nil {
			t.FailNow()
		}

		if nextCalled != v.nextExpected {
			t.Error("Next call not passed " + v.url)
			t.Fail()
		}
	}
}