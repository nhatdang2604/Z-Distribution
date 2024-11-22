package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nhatdang2604/z-distribution/engine/zk"
	"github.com/nhatdang2604/z-distribution/engine/zk/config"
)

func main() {

	// Establishe connection with Zookeeper
	zkConfig := config.NewZkConfig(
		1*time.Minute,
		"/counter",
		"/lock",
	)
	zkEngine := zk.NewZkEngine(zkConfig)
	zkEngine.Start()
	defer zkConfig.Stop()

	// Read commands from user
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command (GET/INC/SLEEPINC): ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		cmd = strings.ToLower(cmd)
		zkEngine.Handle(cmd)
	}
}
