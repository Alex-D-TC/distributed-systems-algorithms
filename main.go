package main

import (
	"fmt"
	"time"

	"github.com/alex-d-tc/distributed-systems-algorithms/pfd"

	"github.com/alex-d-tc/distributed-systems-algorithms/environment"
	"github.com/joho/godotenv"
)

func main() {

	// Load the environment variables
	godotenv.Load()

	env, err := environment.NewNetworkEnvironment()
	if err != nil {
		panic(err)
	}

	// Startup the system
	fd := pfd.NewPerfectFailureDetector(env.GetPFDPort(), env.GetHosts(), 500*time.Millisecond)

	processCrashedListener := make(chan string, 1)

	fd.AddOnProcessCrashedListener(processCrashedListener)

	for {
		host := <-processCrashedListener
		fmt.Println("Host crashed: ", host)
	}
}
