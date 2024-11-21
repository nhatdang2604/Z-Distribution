package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nhatdang2604/z-distribution/config"
)

func main() {

	// Establishe connection with Zookeeper
	zkConfig := config.NewZkConfig()
	_, err := zkConfig.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer zkConfig.Close()

	// Initalize data for Zookeeper
	var counterPath = "/counter"
	var lockPath = "/lock"
	err = zkConfig.Init(counterPath, lockPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read commands from user
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command (GET/INC): ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		zkConfig.Handle(cmd)
	}
}
