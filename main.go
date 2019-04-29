package main

import (
	"fmt"

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

	fmt.Println(env.GetHosts())

	fmt.Println("Hello World!")
}
