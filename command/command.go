package command

import (
	"context"
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/uc"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
)

type CommandService struct {
	log *log.Logger
	beb *beb.BestEffortBroadcast
	pfd *pfd.PerfectFailureDetector
	uc  *uc.UniformConsensus
}

func NewCommandService(beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector, uc *uc.UniformConsensus) *CommandService {

	return &CommandService{
		log: log.New(os.Stdout, "[CommandService]", log.Ldate|log.Ltime),
		beb: beb,
		pfd: pfd,
		uc:  uc,
	}
}

func (cm *CommandService) BEBBroadcast(ctx context.Context, req *protocol.BEBRequest) (*protocol.BEBConfirm, error) {
	cm.log.Println("Received BEBBroadcast call", req)
	cm.beb.Broadcast(req.GetMessage())
	return &protocol.BEBConfirm{
		Result: &protocol.BEBConfirm_Ok{Ok: true},
	}, nil
}

func (cm *CommandService) UCProposal(ctx context.Context, req *protocol.UCRequest) (*protocol.UCReply, error) {
	cm.log.Println("Received UCRequest call", req)
	err := cm.uc.Propose(req.Value)
	if err != nil {
		return &protocol.UCReply{
			Result: &protocol.UCReply_Error{
				Error: &protocol.Error{
					Error: err.Error(),
				},
			},
		}, nil
	}

	return &protocol.UCReply{
		Result: &protocol.UCReply_Ok{Ok: true},
	}, nil
}

func (cm *CommandService) MarcoPolo(ctx context.Context, req *protocol.Empty) (*protocol.Empty, error) {
	cm.log.Println("Received MarcoPolo call")
	return &protocol.Empty{}, nil
}
