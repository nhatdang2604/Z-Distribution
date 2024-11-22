package zk

import (
	"fmt"
	"strings"

	"github.com/nhatdang2604/z-distribution/engine/zk/config"
	"github.com/nhatdang2604/z-distribution/engine/zk/handler"
)

type ZkEngine struct {
	config     *config.ZkConfig
	getHandler *handler.GetHandler
	incHandler *handler.IncHandler
}

func (e *ZkEngine) Start() error {

	// Init the Zookeeper config
	err := e.config.Start()
	if err != nil {
		return err
	}

	// Init handlers
	e.getHandler = handler.NewGetHandler(e.config)
	e.incHandler = handler.NewIncHandler(e.config, e.getHandler)

	return nil
}

func (e *ZkEngine) Stop() {
	e.config.Stop()
}

// Handle commands from the user
func (e *ZkEngine) Handle(cmd string) {

	switch cmd {
	case strings.ToLower(e.getHandler.Key()):
		counter, _, err := e.getHandler.Handle()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Current counter value: %d\n", counter)
		}
	case strings.ToLower(e.incHandler.Key()):
		err := e.incHandler.Handle()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Increase counter successfully")
		}
	default:
		fmt.Println("Invalid command. Use 'GET' or 'INC'.")
	}
}

func NewZkEngine(config *config.ZkConfig) *ZkEngine {
	return &ZkEngine{config: config}
}
