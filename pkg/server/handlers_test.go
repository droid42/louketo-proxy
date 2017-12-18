/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/gambol99/keycloak-proxy/pkg/api"
	"github.com/gambol99/keycloak-proxy/pkg/constants"
)

func TestDebugHandler(t *testing.T) {
	c := newFakeKeycloakConfig()
	c.Resources = make([]*api.Resource, 0)
	c.EnableProfiling = true
	requests := []fakeRequest{
		{URI: "/debug/pprof/no_there", ExpectedCode: http.StatusNotFound},
		{URI: "/debug/pprof/heap", ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/goroutine", ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/block", ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/threadcreate", ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/cmdline", ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/trace", ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/symbol", ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/symbol", Method: http.MethodPost, ExpectedCode: http.StatusOK},
		{URI: "/debug/pprof/symbol", Method: http.MethodPost, ExpectedCode: http.StatusOK},
	}
	newFakeProxy(c).RunTests(t, requests)
}

func TestExpirationHandler(t *testing.T) {
	uri := constants.OauthURL + constants.ExpiredURL
	requests := []fakeRequest{
		{
			URI:          uri,
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			URI:          uri,
			HasToken:     true,
			Expires:      time.Duration(-48 * time.Hour),
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			URI:          uri,
			HasToken:     true,
			Expires:      time.Duration(14 * time.Hour),
			ExpectedCode: http.StatusOK,
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestOauthRequestNotProxying(t *testing.T) {
	requests := []fakeRequest{
		{URI: "/oauth/test"},
		{URI: "/oauth/..//oauth/test/"},
		{URI: "/oauth/expired", Method: http.MethodPost, ExpectedCode: http.StatusMethodNotAllowed},
		{URI: "/oauth/expiring", Method: http.MethodPost},
		{URI: "/oauth%2F///../test%2F%2Foauth"},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestLoginHandlerDisabled(t *testing.T) {
	c := newFakeKeycloakConfig()
	c.EnableLoginHandler = false
	requests := []fakeRequest{
		{URI: constants.OauthURL + constants.LoginURL, Method: http.MethodPost, ExpectedCode: http.StatusNotImplemented},
		{URI: constants.OauthURL + constants.LoginURL, ExpectedCode: http.StatusMethodNotAllowed},
	}
	newFakeProxy(c).RunTests(t, requests)
}

func TestLoginHandlerNotDisabled(t *testing.T) {
	c := newFakeKeycloakConfig()
	c.EnableLoginHandler = true
	requests := []fakeRequest{
		{URI: "/oauth/login", Method: http.MethodPost, ExpectedCode: http.StatusBadRequest},
	}
	newFakeProxy(c).RunTests(t, requests)
}

func TestLoginHandler(t *testing.T) {
	uri := constants.OauthURL + constants.LoginURL
	requests := []fakeRequest{
		{
			URI:          uri,
			Method:       http.MethodPost,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			URI:          uri,
			Method:       http.MethodPost,
			FormValues:   map[string]string{"username": "test"},
			ExpectedCode: http.StatusBadRequest,
		},
		{
			URI:          uri,
			Method:       http.MethodPost,
			FormValues:   map[string]string{"password": "test"},
			ExpectedCode: http.StatusBadRequest,
		},
		{
			URI:    uri,
			Method: http.MethodPost,
			FormValues: map[string]string{
				"password": "test",
				"username": "test",
			},
			ExpectedCode: http.StatusOK,
		},
		{
			URI:    uri,
			Method: http.MethodPost,
			FormValues: map[string]string{
				"password": "test",
				"username": "notmypassword",
			},
			ExpectedCode: http.StatusUnauthorized,
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestLogoutHandlerBadRequest(t *testing.T) {
	requests := []fakeRequest{
		{URI: constants.OauthURL + constants.LogoutURL, ExpectedCode: http.StatusBadRequest},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestLogoutHandlerBadToken(t *testing.T) {
	requests := []fakeRequest{
		{
			URI:          constants.OauthURL + constants.LogoutURL,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			URI:            constants.OauthURL + constants.LogoutURL,
			HasCookieToken: true,
			RawToken:       "this.is.a.bad.token",
			ExpectedCode:   http.StatusBadRequest,
		},
		{
			URI:          constants.OauthURL + constants.LogoutURL,
			RawToken:     "this.is.a.bad.token",
			ExpectedCode: http.StatusBadRequest,
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestLogoutHandlerGood(t *testing.T) {
	requests := []fakeRequest{
		{
			URI:          constants.OauthURL + constants.LogoutURL,
			HasToken:     true,
			ExpectedCode: http.StatusOK,
		},
		{
			URI:              constants.OauthURL + constants.LogoutURL + "?redirect=http://example.com",
			HasToken:         true,
			ExpectedCode:     http.StatusTemporaryRedirect,
			ExpectedLocation: "http://example.com",
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestTokenHandler(t *testing.T) {
	uri := constants.OauthURL + constants.TokenURL
	requests := []fakeRequest{
		{
			URI:          uri,
			HasToken:     true,
			ExpectedCode: http.StatusOK,
		},
		{
			URI:          uri,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			URI:          uri,
			RawToken:     "niothing",
			ExpectedCode: http.StatusBadRequest,
		},
		{
			URI:            uri,
			HasToken:       true,
			HasCookieToken: true,
			ExpectedCode:   http.StatusOK,
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestServiceRedirect(t *testing.T) {
	requests := []fakeRequest{
		{
			URI:              "/admin",
			Redirects:        true,
			ExpectedCode:     http.StatusTemporaryRedirect,
			ExpectedLocation: "/oauth/authorize?state=L2FkbWlu",
		},
		{
			URI:          "/admin",
			ExpectedCode: http.StatusUnauthorized,
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestAuthorizationURLWithSkipToken(t *testing.T) {
	c := newFakeKeycloakConfig()
	c.SkipTokenVerification = true
	newFakeProxy(c).RunTests(t, []fakeRequest{
		{
			URI:          constants.OauthURL + constants.AuthorizationURL,
			ExpectedCode: http.StatusNotAcceptable,
		},
	})
}

func TestAuthorizationURL(t *testing.T) {
	requests := []fakeRequest{
		{
			URI:              "/admin",
			Redirects:        true,
			ExpectedLocation: "/oauth/authorize?state=L2FkbWlu",
			ExpectedCode:     http.StatusTemporaryRedirect,
		},
		{
			URI:              "/admin/test",
			Redirects:        true,
			ExpectedLocation: "/oauth/authorize?state=L2FkbWluL3Rlc3Q=",
			ExpectedCode:     http.StatusTemporaryRedirect,
		},
		{
			URI:              "/help/../admin",
			Redirects:        true,
			ExpectedLocation: "/oauth/authorize?state=L2FkbWlu",
			ExpectedCode:     http.StatusTemporaryRedirect,
		},
		{
			URI:              "/admin?test=yes&test1=test",
			Redirects:        true,
			ExpectedLocation: "/oauth/authorize?state=L2FkbWluP3Rlc3Q9eWVzJnRlc3QxPXRlc3Q=",
			ExpectedCode:     http.StatusTemporaryRedirect,
		},
		{
			URI:          "/oauth/test",
			Redirects:    true,
			ExpectedCode: http.StatusNotFound,
		},
		{
			URI:          "/oauth/callback/..//test",
			Redirects:    true,
			ExpectedCode: http.StatusNotFound,
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}

func TestCallbackURL(t *testing.T) {
	cfg := newFakeKeycloakConfig()
	requests := []fakeRequest{
		{
			URI:          constants.OauthURL + constants.CallbackURL,
			Method:       http.MethodPost,
			ExpectedCode: http.StatusMethodNotAllowed,
		},
		{
			URI:          constants.OauthURL + constants.CallbackURL,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			URI:              constants.OauthURL + constants.CallbackURL + "?code=fake",
			ExpectedCookies:  map[string]string{cfg.CookieAccessName: ""},
			ExpectedLocation: "/",
			ExpectedCode:     http.StatusTemporaryRedirect,
		},
		{
			URI:              constants.OauthURL + constants.CallbackURL + "?code=fake&state=/admin",
			ExpectedCookies:  map[string]string{cfg.CookieAccessName: ""},
			ExpectedLocation: "/",
			ExpectedCode:     http.StatusTemporaryRedirect,
		},
		{
			URI:              constants.OauthURL + constants.CallbackURL + "?code=fake&state=L2FkbWlu",
			ExpectedCookies:  map[string]string{cfg.CookieAccessName: ""},
			ExpectedLocation: "/admin",
			ExpectedCode:     http.StatusTemporaryRedirect,
		},
	}
	newFakeProxy(cfg).RunTests(t, requests)
}

func TestHealthHandler(t *testing.T) {
	requests := []fakeRequest{
		{
			URI:             constants.OauthURL + constants.HealthURL,
			ExpectedCode:    http.StatusOK,
			ExpectedContent: "OK\n",
		},
		{
			URI:          constants.OauthURL + constants.HealthURL,
			Method:       http.MethodHead,
			ExpectedCode: http.StatusMethodNotAllowed,
		},
	}
	newFakeProxy(nil).RunTests(t, requests)
}
