package chat

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// Configuration for server behavior.
// it must construct by LoadConfig() or LoadConfigFile().
type Config struct {
	// Http service address for the server.
	HTTP string

	// root path for the websocket server.
	WebSocketPath string
}

const (
	DefaultHTTP          = "localhost:8080"
	DefaultWebSocketPath = "/chat/ws"
)

var DefaultConfig = Config{
	HTTP:          DefaultHTTP,
	WebSocketPath: DefaultWebSocketPath,
}

// it loads the configuration from file.
// it returns loaded config and load error.
func LoadConfigFile(file string) (*Config, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return LoadConfig(fp)
}

// it loads the configuration from io.Reader.
// it returns loaded config and load error.
func LoadConfig(r io.Reader) (*Config, error) {
	conf := &Config{}
	if err := decode(r, conf); err != nil {
		return nil, fmt.Errorf("LoadConfig: %v", err)
	}
	return conf, nil
}

// decode from reader and store it to data.
func decode(r io.Reader, data interface{}) error {
	meta, err := toml.DecodeReader(r, data)
	if undecoded := meta.Undecoded(); undecoded != nil && len(undecoded) > 0 {
		log.Println("Config.Decode:", "undecoded keys exist,", undecoded)
	}
	return err
}
