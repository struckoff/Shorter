package configuration

import "github.com/BurntSushi/toml"

type Configuration struct {
	DBPath  string `toml:"DBPath"`
	Address string `toml:"Address"`
}

func (conf *Configuration) Read(path string) error {
	if _, err := toml.DecodeFile(path, conf); err != nil {
		return err
	}
	return nil
}
