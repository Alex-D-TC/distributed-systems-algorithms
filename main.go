package main

import (
	"log"
	"os"
	"time"

	"github.com/alex-d-tc/distributed-systems-algorithms/environment"
	"github.com/alex-d-tc/distributed-systems-algorithms/pfd"
	"github.com/joho/godotenv"
)

func main() {

	logger := log.New(os.Stdout, "[Main]", log.Ldate|log.Ltime)

	// Load the environment variables
	godotenv.Load()

	env, err := environment.NewNetworkEnvironment()
	if err != nil {
		panic(err)
	}

	// Startup the system
	fd := pfd.NewPerfectFailureDetector(env.GetPFDPort(), env.GetHosts(), 3*time.Second, 3)

	processCrashedListener := make(chan string, 1)

	fd.AddOnProcessCrashedListener(processCrashedListener)

	for {
		host := <-processCrashedListener
		logger.Println("Host crashed: ", host)
	}
}
