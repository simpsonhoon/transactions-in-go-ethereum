package config

import (
	"fmt"
	"os"

	"github.com/naoina/toml"
)

type Config struct {
	Network struct {
		URL string
	}

	DB struct {
		Host string
	}
}

func GetConfig(fpath string) *Config {
	c := new(Config)

	if file, err := os.Open(fpath); err != nil {
		panic(err)
	} else {
		defer file.Close()
		if err := toml.NewDecoder(file).Decode(c); err != nil {
			panic(err)
		} else {
			fmt.Println(c)
			return c
		}
	}
}
