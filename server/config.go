package server

import (
	"fmt"
	"strings"
)

// Configuration for server behavior.
// it must construct by LoadConfig() or LoadConfigFile().
type Config struct {
	// HTTP service address for the server.
	// The format is `[host]:[port]`, e.g. localhost:8080.
	HTTP string

	// Prefix of URI for the chat API.
	// e.g. given ChatAPIPrefix = `/api` and chat API `/chat/rooms`,
	// the prefixed chat API is `/api/chat/rooms`.
	ChatAPIPrefix string

	// Prefix of URI for the static file server.
	//
	// Example: given local html file `/www/index.html`,
	// StaticHandlerPrefix = "/www" and StaticHandlerPrefix = "/static",
	// the requesting the server with URI `/static/index.html` responds
	// the html content of `/www/index.html`.
	StaticHandlerPrefix string

	// root directory to serve static files.
	StaticFileDir string

	// indicates whether serving static files is enable.
	// if false, StaticHandlerPrefix and StaticFileDir do not
	// affect the server.
	EnableServeStaticFile bool

	// show all of URI routes at server starts.
	ShowRoutes bool
}

// DefaultConfig is default configuration for the server.
var DefaultConfig = Config{
	HTTP:                  "localhost:8080",
	ChatAPIPrefix:         "",
	StaticHandlerPrefix:   "",
	StaticFileDir:         "", // current directory
	EnableServeStaticFile: true,
	ShowRoutes:            true,
}

// Validate checks whether the all of field values are correct format.
func (c *Config) Validate() error {
	if ss := strings.Split(c.HTTP, ":"); len(ss) != 2 {
		return fmt.Errorf("config: HTTP should be [host]:[port], but %v", c.HTTP)
	}
	if len(c.ChatAPIPrefix) > 0 && !strings.HasPrefix(c.ChatAPIPrefix, "/") {
		return fmt.Errorf("config: ChatAPIPrefix should start with \"/\" but %v", c.ChatAPIPrefix)
	}
	if len(c.StaticHandlerPrefix) > 0 && !strings.HasPrefix(c.StaticHandlerPrefix, "/") {
		return fmt.Errorf("config: StaticHandlerPrefix should start with \"/\" but %v", c.StaticHandlerPrefix)
	}
	return nil
}
