package config

import (
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/shirasudon/go-chat/server"
)

const (
	ExampleFile  = "./example/config.toml"
	NotFoundFile = "path/to/not/found"
)

func TestLoadFile(t *testing.T) {
	conf, err := LoadFile(ExampleFile)
	if err != nil {
		t.Fatal(err)
	}
	if *conf != server.DefaultConfig {
		t.Errorf("different config value, expect: %#v, got: %#v", server.DefaultConfig, conf)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile(NotFoundFile)
	if err == nil {
		t.Fatal("not found file is given, but no error")
	}
}

func TestLoadByteInvalid(t *testing.T) {
	conf := server.DefaultConfig
	conf.HTTP = "invalid string"

	buf := new(bytes.Buffer)
	if _, err := toml.DecodeReader(buf, &conf); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadByte(buf); err == nil {
		t.Errorf("invalid config.HTTP is given, but no error")
	}
}
