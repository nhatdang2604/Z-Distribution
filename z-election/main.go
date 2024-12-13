package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"z-election/zk"
	"z-election/zk/config"
)

func main() {

	// Establishe connection with Zookeeper
	zkConfig := config.NewZkConfig(
		1*time.Minute,
		"/leader",
		"/consensus",
	)
	zkEngine := zk.NewElectionCandidate(zkConfig)
	zkEngine.Start()
	defer zkConfig.Stop()

	// Read commands from user
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command (GET/INC): ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		cmd = strings.ToLower(cmd)
		_, error := zkEngine.Handle(cmd)
		if error != nil {
			fmt.Println(error)
		}
	}
}
