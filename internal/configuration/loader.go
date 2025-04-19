package configuration

import "github.com/BurntSushi/toml"

func Load() (*Configuration, error) {
	f := "./config.toml"

	var c Configuration
	if _, err := toml.DecodeFile(f, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
