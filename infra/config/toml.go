package config

import (
	"fmt"
	"io"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/shirasudon/go-chat/server"
)

// FileExists returns whether given file path is exist?
func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

// it loads the configuration from file.
// it returns loaded config and load error.
func LoadFile(file string) (*server.Config, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return LoadByte(fp)
}

// it loads the configuration from io.Reader.
// it returns loaded config and load error.
func LoadByte(r io.Reader) (*server.Config, error) {
	conf := &server.Config{}
	if err := decode(r, conf); err != nil {
		return nil, fmt.Errorf("infra/config: %v", err)
	}
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("infra/config: validation erorr: %v", err)
	}
	return conf, nil
}

// decode from reader and store it to data.
func decode(r io.Reader, data interface{}) error {
	meta, err := toml.DecodeReader(r, data)
	if undecoded := meta.Undecoded(); undecoded != nil && len(undecoded) > 0 {
		fmt.Fprintln(os.Stderr, "infra/config.Decode:", "undecoded keys exist,", undecoded)
	}
	return err
}
