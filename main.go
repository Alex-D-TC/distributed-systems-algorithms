package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/alex-d-tc/distributed-systems-algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/command"
	"github.com/alex-d-tc/distributed-systems-algorithms/environment"
	"github.com/alex-d-tc/distributed-systems-algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"

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
	logger.Println(env.GetHosts())
	logger.Println(env.GetONARPort())
	logger.Println(env.GetPFDPort())
	logger.Println(env.GetUCPort())
	
	// Startup the system
	fd := pfd.NewPerfectFailureDetector(env.GetPFDPort(), env.GetHosts(), 3*time.Second, 3)

	beb := beb.NewBestEffortBroadcast(env.GetBebPort(), env.GetHosts())

	go commandListener(env.GetControlPort(), logger, fd, beb)

	bebListener := beb.AddOnDeliverListener()

	// Wait forever for now
	for {
		logger.Println("Awaiting broadcasts")
		broadcast := <-bebListener
		logger.Println("Received broadcast: ", broadcast)
	}
}

func commandListener(controlPort uint16, logger *log.Logger, fd *pfd.PerfectFailureDetector, beb *beb.BestEffortBroadcast) {
	logger.Println("Starting command listener")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", controlPort))
	if err != nil {
		logger.Fatal(err.Error())
	}

	commandService := command.NewCommandService(beb, fd)

	server := grpc.NewServer()
	protocol.RegisterControlServiceServer(server, commandService)

	server.Serve(lis)
}
