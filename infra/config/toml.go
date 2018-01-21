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

// it loads the configuration from file into dest.
// it returns load error if any.
func LoadFile(dest *server.Config, file string) error {
	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()
	return LoadByte(dest, fp)
}

// it loads the configuration from io.Reader into dest.
// it returns load error if any.
func LoadByte(dest *server.Config, r io.Reader) error {
	if err := decode(r, dest); err != nil {
		return fmt.Errorf("infra/config: %v", err)
	}
	if err := dest.Validate(); err != nil {
		return fmt.Errorf("infra/config: validation erorr: %v", err)
	}
	return nil
}

// decode from reader and store it to data.
func decode(r io.Reader, data interface{}) error {
	meta, err := toml.DecodeReader(r, data)
	if undecoded := meta.Undecoded(); undecoded != nil && len(undecoded) > 0 {
		fmt.Fprintln(os.Stderr, "infra/config.Decode:", "undecoded keys exist,", undecoded)
	}
	return err
}
