package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/shirasudon/go-chat/server"
)

const (
	ExampleFile  = "./example/config.toml"
	NotFoundFile = "path/to/not/found"
)

func TestFileExists(t *testing.T) {
	t.Parallel()
	if !FileExists(ExampleFile) {
		t.Error("exist file is not detected")
	}

	if FileExists(NotFoundFile) {
		t.Error("not exist file is detected")
	}
}

func TestLoadFile(t *testing.T) {
	t.Parallel()
	conf := server.Config{}
	if err := LoadFile(&conf, ExampleFile); err != nil {
		t.Fatal(err)
	}
	if conf != server.DefaultConfig {
		t.Errorf("different config value, expect: %#v, got: %#v", server.DefaultConfig, conf)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	t.Parallel()
	conf := server.Config{}
	if err := LoadFile(&conf, NotFoundFile); err == nil {
		t.Fatal("not found file is given, but no error")
	}
	if conf != (server.Config{}) {
		t.Errorf("failed to load external config, but unexpected values are set: %#v", conf)
	}
}

func TestLoadByteInvalid(t *testing.T) {
	t.Parallel()
	conf := server.DefaultConfig
	conf.HTTP = "invalid string"

	buf := new(bytes.Buffer)
	if _, err := toml.DecodeReader(buf, &conf); err != nil {
		t.Fatal(err)
	}

	dest := server.Config{}
	if err := LoadByte(&dest, buf); err == nil {
		t.Errorf("invalid config.HTTP is given, but no error")
	}
}

func TestLoadByteOverwrite(t *testing.T) {
	t.Parallel()

	// over writes only Static* fields
	const ConfigBody = `
# HTTP = "localhost:8080"
# ChatAPIPrefix = ""
StaticHandlerPrefix = "/static"
StaticFileDir = "public"
# EnableServeStaticFile = true
# ShowRoutes = true`

	buf := strings.NewReader(ConfigBody)
	dest := server.DefaultConfig
	if err := LoadByte(&dest, buf); err != nil {
		t.Fatal(err)
	}

	defaultC := server.DefaultConfig
	if dest.HTTP != defaultC.HTTP ||
		dest.ShowRoutes != defaultC.ShowRoutes ||
		dest.EnableServeStaticFile != defaultC.EnableServeStaticFile ||
		dest.ChatAPIPrefix != defaultC.ChatAPIPrefix {
		t.Errorf("comment-outed values are overwritten, got: %#v", dest)
	}

	if dest.StaticFileDir != "public" || dest.StaticHandlerPrefix != "/static" {
		t.Errorf("overwritten values are not changed, got: %#v", dest)
	}
}
