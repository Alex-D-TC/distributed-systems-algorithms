package main

import (
	"fmt"

	"github.com/joho/godotenv"
)

func main() {

	// Load the environment variables
	godotenv.Load()

	fmt.Println("Hello World!")
}
