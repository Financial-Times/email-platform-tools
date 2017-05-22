package config

import (
	"bufio"
	"errors"
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

type Config struct {
	UsersAuth string `json:"usersauth"`
}

func Bind(path string, config interface{}) error {
	var buf []byte
	if _, err := os.Stat(path); err != nil {
		return errors.New("config path not valid")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	bufr := bufio.NewReader(f)

	for run := true; run; {
		line, err := bufr.ReadString('\n')
		switch err {
		case io.EOF:
			run = false
		case nil:
		default:
			return err
		}
		line = os.ExpandEnv(line)
		buf = append(buf, []byte(line)...)
	}

	return yaml.Unmarshal(buf, config)
}
