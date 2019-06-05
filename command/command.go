package command

import (
	"context"
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/onar"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/uc"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/urb"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
)

type CommandService struct {
	log  *log.Logger
	beb  *beb.BestEffortBroadcast
	pfd  *pfd.PerfectFailureDetector
	uc   *uc.UniformConsensus
	urb  *urb.URB
	onar *onar.ONAR
}

func NewCommandService(beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector, uc *uc.UniformConsensus, urb *urb.URB, onar *onar.ONAR) *CommandService {

	return &CommandService{
		log:  log.New(os.Stdout, "[CommandService]", log.Ldate|log.Ltime),
		beb:  beb,
		pfd:  pfd,
		uc:   uc,
		urb:  urb,
		onar: onar,
	}
}

func (cm *CommandService) ONARRead(ctx context.Context, req *protocol.ONARReadRequest) (*protocol.ONARReadReply, error) {
	cm.log.Println("Received ONAR read command")
	err := cm.onar.Read()
	if err != nil {
		return &protocol.ONARReadReply{
			Result: &protocol.ONARReadReply_Error{
				Error: &protocol.Error{
					Error: err.Error(),
				},
			},
		}, nil
	}

	return &protocol.ONARReadReply{
		Result: &protocol.ONARReadReply_Ok{Ok: true},
	}, nil
}

func (cm *CommandService) ONARWrite(ctx context.Context, req *protocol.ONARWriteRequest) (*protocol.ONARWriteReply, error) {
	cm.log.Println("Received ONAR write command ", req)
	err := cm.onar.Write(req.Value)
	if err != nil {
		return &protocol.ONARWriteReply{
			Result: &protocol.ONARWriteReply_Error{
				Error: &protocol.Error{
					Error: err.Error(),
				},
			},
		}, nil
	}

	return &protocol.ONARWriteReply{
		Result: &protocol.ONARWriteReply_Ok{Ok: true},
	}, nil
}

func (cm *CommandService) URBBroadcast(ctx context.Context, req *protocol.URBRequest) (*protocol.URBReply, error) {
	cm.log.Println("Received URBBroadcast call", req)
	err := cm.urb.Broadcast(req.Data)
	if err != nil {
		return &protocol.URBReply{
			Result: &protocol.URBReply_Error{
				Error: &protocol.Error{
					Error: err.Error(),
				},
			},
		}, nil
	}

	return &protocol.URBReply{
		Result: &protocol.URBReply_Ok{Ok: true},
	}, nil
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
