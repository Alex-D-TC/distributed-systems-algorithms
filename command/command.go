package command

import (
	"context"
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
)

type CommandService struct {
	log *log.Logger
	beb *beb.BestEffortBroadcast
	pfd *pfd.PerfectFailureDetector
}

func NewCommandService(beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector) *CommandService {

	return &CommandService{
		log: log.New(os.Stdout, "[CommandService]", log.Ldate|log.Ltime),
		beb: beb,
		pfd: pfd,
	}
}

func (cm *CommandService) BEBBroadcast(ctx context.Context, req *protocol.BEBRequest) (*protocol.BEBConfirm, error) {
	cm.log.Println("Received BEBBroadcast call", req)
	cm.beb.Broadcast(req.GetMessage())
	return &protocol.BEBConfirm{
		Result: &protocol.BEBConfirm_Ok{Ok: true},
	}, nil
}

func (cm *CommandService) MarcoPolo(ctx context.Context, req *protocol.Empty) (*protocol.Empty, error) {
	cm.log.Println("Received MarcoPolo call")
	return &protocol.Empty{}, nil
}
