package command

import (
	"context"

	"github.com/alex-d-tc/distributed-systems-algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
)

type CommandService struct {
	beb *beb.BestEffortBroadcast
	pfd *pfd.PerfectFailureDetector
}

func NewCommandService(beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector) *CommandService {

	return &CommandService{
		beb: beb,
		pfd: pfd,
	}
}

func (cm *CommandService) BEBBroadcast(ctx context.Context, req *protocol.BEBRequest) (*protocol.BEBConfirm, error) {
	cm.beb.Broadcast(req.GetMessage())
	return &protocol.BEBConfirm{
		Result: &protocol.BEBConfirm_Ok{Ok: true},
	}, nil
}
