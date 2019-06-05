package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/uc"
	"github.com/alex-d-tc/distributed-systems-algorithms/command"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/alex-d-tc/distributed-systems-algorithms/util/environment"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {

	logger := log.New(os.Stdout, "[Main]", log.Ldate|log.Ltime)

	// Load the environment variables
	godotenv.Load()

	env, err := environment.NewNetworkEnvironment()
	if err != nil {
		panic(err)
	}

	logger.Println(env.GetBebPort())
	logger.Println(env.GetControlPort())
	logger.Println(env.GetHostname())
	logger.Println(env.GetHosts())
	logger.Println(env.GetONARPort())
	logger.Println(env.GetPFDPort())
	logger.Println(env.GetUCPort())

	// Startup the system
	pfd := pfd.NewPerfectFailureDetector(env.GetPFDPort(), env.GetHosts(), 3*time.Second, 3)

	beb := beb.NewBestEffortBroadcast(env.GetBebPort(), env.GetHosts())

	uc := uc.NewUniformConsensus(env, beb, pfd)

	go commandListener(env.GetControlPort(), logger, pfd, beb, uc)

	ucListener := uc.AddOnDecidedListener()

	// Wait forever for now
	for {
		logger.Println("Awaiting broadcasts")
		decision := <-ucListener
		logger.Println("Received UC decision: ", decision)
	}
}

func commandListener(controlPort uint16, logger *log.Logger, pfd *pfd.PerfectFailureDetector, beb *beb.BestEffortBroadcast, uc *uc.UniformConsensus) {
	logger.Println("Starting command listener")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", controlPort))
	if err != nil {
		logger.Fatal(err.Error())
	}

	commandService := command.NewCommandService(beb, pfd, uc)

	server := grpc.NewServer()
	protocol.RegisterControlServiceServer(server, commandService)

	server.Serve(lis)
}
