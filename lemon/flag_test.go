package lemon

import (
	"os"
	"reflect"
	"testing"
)

func TestCLIParse(t *testing.T) {
	assert := func(args []string, expected CLI) {
		expected.In = os.Stdin
		c := &CLI{In: os.Stdin}
		c.FlagParse(args, true)

		if !reflect.DeepEqual(expected, *c) {
			t.Errorf("Expected:\n %+v, but got\n %+v", expected, c)
		}
	}

	defaultPort := 2489
	defaultHost := "localhost"
	defaultAllow := "0.0.0.0/0,::/0"
	defaultLogLevel := 1

	assert([]string{"pbpaste", "--port", "1124"}, CLI{
		Type:     PASTE,
		Host:     defaultHost,
		Port:     1124,
		Allow:    defaultAllow,
		LogLevel: defaultLogLevel,
	})

	assert([]string{"/usr/bin/pbpaste", "--port", "1124"}, CLI{
		Type:     PASTE,
		Host:     defaultHost,
		Port:     1124,
		Allow:    defaultAllow,
		LogLevel: defaultLogLevel,
	})

	assert([]string{"vimonade", "paste"}, CLI{
		Type:     PASTE,
		Host:     defaultHost,
		Port:     defaultPort,
		Allow:    defaultAllow,
		LogLevel: defaultLogLevel,
	})

	assert([]string{"pbcopy", "hogefuga"}, CLI{
		Type:       COPY,
		Host:       defaultHost,
		Port:       defaultPort,
		Allow:      defaultAllow,
		DataSource: "hogefuga",
		LogLevel:   defaultLogLevel,
	})

	assert([]string{"/usr/bin/pbcopy", "hogefuga"}, CLI{
		Type:       COPY,
		Host:       defaultHost,
		Port:       defaultPort,
		Allow:      defaultAllow,
		DataSource: "hogefuga",
		LogLevel:   defaultLogLevel,
	})

	assert([]string{"vimonade", "copy", "hogefuga"}, CLI{
		Type:       COPY,
		Host:       defaultHost,
		Port:       defaultPort,
		Allow:      defaultAllow,
		DataSource: "hogefuga",
		LogLevel:   defaultLogLevel,
	})

	assert([]string{"vimonade", "send", "hogefuga.txt"}, CLI{
		Type:       SEND,
		Host:       defaultHost,
		Port:       defaultPort,
		Allow:      defaultAllow,
		DataSource: "hogefuga.txt",
		LogLevel:   defaultLogLevel,
	})

	assert([]string{"vimonade", "--allow", "192.168.0.0/24", "server", "--port", "1124"}, CLI{
		Type:     SERVER,
		Host:     defaultHost,
		Port:     1124,
		Allow:    "192.168.0.0/24",
		LogLevel: defaultLogLevel,
	})
}
