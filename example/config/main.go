package main

import (
	"log"
	"os"
	"time"

	"github.com/libgox/core/config/dynamic"
	"github.com/libgox/core/config/dynamic/parse"
	"github.com/libgox/core/config/dynamic/source"
	"github.com/libgox/core/config/dynamic/types"
)

func main() {
	fileSource, err := source.NewFileSource("./config.json", types.Dynamic)
	if err != nil {
		log.Fatal(err)
	}

	logger := Logger{level: "info"}

	manager := dynamic.NewConfigManager(fileSource, parse.JSONParseFunc[LogConfig])
	manager.RegisterListener(&logger)

	err = manager.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("./config.json", []byte(`{"level":"debug"}`), 0644)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	log.Println(logger.level)

	os.Remove("./config.json")
	time.Sleep(1 * time.Second)
	log.Println(logger.level)

	manager.Stop()
}

type LogConfig struct {
	Level string `json:"level"`
}

type Logger struct {
	level string
}

func (l *Logger) Update(config LogConfig) {
	if config.Level == "" {
		l.level = "info"
	} else {
		l.level = config.Level
	}
}
