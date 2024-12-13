package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"z-election/zk"
	"z-election/zk/config"
)

func handlerHOF(callback func(cmd string)) func(w http.ResponseWriter, r *http.Request) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
		switch r.Method {
		case "POST":
			// Handle POST request and read the body message
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Unable to read the body", http.StatusInternalServerError)
				return
			}
			// Respond based on the posted message
			cmd := string(body)
			callback(cmd)

		default:
			// Respond to other HTTP methods
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	}

	return handler
}

func serverAsWebserver() {

	// Establishe connection with Zookeeper
	nodeId := int32(time.Now().UnixNano())
	zkConfig := config.NewZkConfig(
		nodeId,
		1*time.Minute,
		"/leader",
		"/consensus",
	)
	zkElectionCandidate := zk.NewElectionCandidate(zkConfig)
	zkElectionCandidate.Start()
	defer zkConfig.Stop()

	// Register the handler function for the "/" route
	http.HandleFunc("/", handlerHOF(func(cmd string) {
		_, error := zkElectionCandidate.Handle(cmd)
		if error != nil {
			fmt.Println(error)
		}
	}))

	//Get port from env
	port := os.Getenv("MY_PORT")

	// Start the web server on port
	fmt.Printf("Starting server on :%v...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func serverAsLoop() {

	instanceCount := 10
	var zkCandidates [10](*zk.ElectionCandidate)

	//Init all candidates config to connect to Zookeeper
	for i := 0; i < instanceCount; i++ {

		nodeId := int32(i)
		zkConfig := config.NewZkConfig(
			nodeId,
			1*time.Minute,
			"/leader",
			"/consensus",
		)
		zkElectionCandidate := zk.NewElectionCandidate(zkConfig)
		zkCandidates[i] = zkElectionCandidate
	}

	//Start all goroutine
	var wg sync.WaitGroup
	commands := []string{"INC", "INC", "INC", "INC", "INC", "INC", "INC", "INC", "INC", "INC"}
	for _, zkElectionCandidate := range zkCandidates {
		zkElectionCandidate.Start()
		defer zkElectionCandidate.Stop()
		wg.Add(1) // Increment the counter for each goroutine
		go func() {

			// Signal that this goroutine is done
			defer wg.Done()

			// Read commands from user
			for _, cmd := range commands {
				cmd = strings.ToLower(cmd)
				zkElectionCandidate.Handle(cmd)
			}
		}()
	}

	//Wait all goroutine to finish
	wg.Wait()
}

func main() {
	serverAsLoop()
}
