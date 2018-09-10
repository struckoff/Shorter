package configuration

import "github.com/BurntSushi/toml"

// Configuration struckture
type Configuration struct {
	DBPath  string `toml:"DBPath"`  // Path to database file
	Address string `toml:"Address"` // Application network address
}

// Configuration.Read parse toml configuration file to struckture
func (conf *Configuration) Read(path string) error {
	if _, err := toml.DecodeFile(path, conf); err != nil {
		return err
	}
	return nil
}
