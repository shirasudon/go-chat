package server

import (
	"context"
	"testing"

	"github.com/labstack/echo"
)

func TestDefaultConfig(t *testing.T) {
	if err := DefaultConfig.Validate(); err != nil {
		t.Errorf("DefaultConfig.Validate: %v", err)
	}
}

func TestConfigValidate(t *testing.T) {
	for _, c := range []Config{
		{HTTP: "a"},
		{HTTP: "a:a:"},
		{HTTP: "a::"},
		{ChatAPIPrefix: "api/chat"},
		{StaticHandlerPrefix: "sta/tic"},
	} {
		if err := c.Validate(); err == nil {
			t.Errorf("It should be error but not, %#v", c)
		}
	}
}

func findRoute(routes []*echo.Route, query echo.Route) bool {
	for _, r := range routes {
		if *r == query {
			return true
		}
	}
	return false
}

func TestServerConfEnableServeStaticFile(t *testing.T) {
	var staticFileRoute = echo.Route{
		Name:   "staticContents",
		Path:   "/*",
		Method: echo.GET,
	}

	{ // enable serve static file
		conf := DefaultConfig
		conf.EnableServeStaticFile = true
		server1 := NewServer(chatCmd, chatQuery, chatHub, queryers.UserQueryer, &conf)
		defer server1.Shutdown(context.Background())
		if !findRoute(server1.echo.Routes(), staticFileRoute) {
			t.Errorf("staticContents route (%#v) is not found", staticFileRoute)
		}
	}

	{ // disable serve static file
		conf := DefaultConfig
		conf.EnableServeStaticFile = false
		server2 := NewServer(chatCmd, chatQuery, chatHub, queryers.UserQueryer, &conf)
		defer server2.Shutdown(context.Background())
		if findRoute(server2.echo.Routes(), staticFileRoute) {
			t.Errorf("staticContents route (%#v) should be not found", staticFileRoute)
		}
	}
}

func TestServerConfStaticHandlerPrefix(t *testing.T) {
	var Query = echo.Route{
		Name:   "staticContents",
		Path:   "/static/*",
		Method: echo.GET,
	}

	{ // set static prefix
		conf := DefaultConfig
		conf.StaticHandlerPrefix = "/static"
		server1 := NewServer(chatCmd, chatQuery, chatHub, queryers.UserQueryer, &conf)
		defer server1.Shutdown(context.Background())
		if !findRoute(server1.echo.Routes(), Query) {
			t.Errorf("staticContents route (%#v) is not found", Query)
		}
	}

	{ // no static prefix
		conf := DefaultConfig
		server2 := NewServer(chatCmd, chatQuery, chatHub, queryers.UserQueryer, &conf)
		defer server2.Shutdown(context.Background())
		if findRoute(server2.echo.Routes(), Query) {
			t.Errorf("staticContents route (%#v) should be not found", Query)
		}
	}
}

func TestServerConfChatAPIPrefix(t *testing.T) {
	var Query = echo.Route{
		Name:   "chat.createRoom",
		Path:   "/api/chat/rooms",
		Method: echo.POST,
	}

	{ // set chat prefix
		conf := DefaultConfig
		conf.ChatAPIPrefix = "/api"
		server1 := NewServer(chatCmd, chatQuery, chatHub, queryers.UserQueryer, &conf)
		defer server1.Shutdown(context.Background())
		if !findRoute(server1.echo.Routes(), Query) {
			t.Errorf("chat API route (%#v) is not found", Query)
		}
	}

	{ // no chat prefix
		conf := DefaultConfig
		server2 := NewServer(chatCmd, chatQuery, chatHub, queryers.UserQueryer, &conf)
		defer server2.Shutdown(context.Background())
		if findRoute(server2.echo.Routes(), Query) {
			t.Errorf("chat API route (%#v) should be not found", Query)
		}
	}
}
