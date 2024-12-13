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

	instanceCount := 5
	raceConditionCount := 20
	var zkCandidates [5](*zk.ElectionCandidate)

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
		zkCandidates[i].Start()
		defer zkCandidates[i].Stop()
	}

	//Start all goroutine
	for roundIdx := 0; roundIdx < raceConditionCount; roundIdx++ {
		cmd := "INC"
		fmt.Printf("Round %v start\n", roundIdx)
		var wg sync.WaitGroup

		for _, zkElectionCandidate := range zkCandidates {
			wg.Add(1) // Increment the counter for each goroutine
			go func() {

				// Signal that this goroutine is done
				defer wg.Done()

				cmd = strings.ToLower(cmd)
				zkElectionCandidate.Handle(cmd)
			}()
		}

		//Wait all goroutine to finish
		wg.Wait()

		fmt.Printf("Round %v end\n", roundIdx)
	}
}

func main() {
	serverAsLoop()
}
