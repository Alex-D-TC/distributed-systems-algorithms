package pfd

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/alex-d-tc/distributed-systems-algorithms/network"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/golang/protobuf/proto"
)

type PerfectFailureDetector struct {
	servicePort uint16
	hosts       []string

	// PFD Main functionality
	dead    map[string]bool
	replied map[string]bool
	delta   time.Duration

	// Outgoing event listeners
	onProcessCrashListeners []chan<- string

	// Logging
	logger *log.Logger
}

func NewPerfectFailureDetector(port uint16, hosts []string, delta time.Duration) *PerfectFailureDetector {

	pfd := &PerfectFailureDetector{
		servicePort: port,
		hosts:       hosts,
		delta:       delta,
		dead:        map[string]bool{},
		replied:     map[string]bool{},
		logger:      log.New(os.Stdout, "[PFD]", log.Ldate|log.Ltime),
	}

	// Start the timer routine
	go pfd.periodicCheck()

	// Start listening routine
	go network.Listen(port, pfd.handleConnection, pfd.logger)

	return pfd
}

func (pfd *PerfectFailureDetector) AddOnProcessCrashedListener(listener chan<- string) {
	pfd.onProcessCrashListeners = append(pfd.onProcessCrashListeners, listener)
}

func (pfd *PerfectFailureDetector) periodicCheck() {

	for {

		// Send the new pfd requests
		pfd.replied = map[string]bool{}

		pfdRequestMessage := &protocol.PFDMessage{
			HeartbeatType: &protocol.PFDMessage_Request{
				Request: &protocol.HeartbeatRequest{},
			},
		}

		rawPfdMessage, err := proto.Marshal(pfdRequestMessage)
		if err != nil {
			pfd.logger.Panic(err.Error())
		}

		for i := 0; i < len(pfd.hosts); i++ {
			network.SendMessage(pfd.hosts[i], pfd.servicePort, rawPfdMessage)
		}

		// Wait for the replies
		time.Sleep(pfd.delta)

		// Find out who died
		for _, host := range pfd.hosts {
			if !pfd.replied[host] && !pfd.dead[host] {
				// Notify the process crashed listeners
				pfd.dead[host] = true
				for _, listener := range pfd.onProcessCrashListeners {
					listener <- host
				}
			}
		}
	}
}

func (pfd *PerfectFailureDetector) onHeartbeatRequest(host string) {

	pfd.logger.Println("Received Heartbeat Request from host:", host)

	reply := &protocol.HeartbeatReply{}

	replyMessage := &protocol.PFDMessage{
		HeartbeatType: &protocol.PFDMessage_Reply{Reply: reply},
	}

	raw, err := proto.Marshal(replyMessage)
	if err != nil {
		// If marshaling fails, stop
		pfd.logger.Println(err.Error())
		return
	}

	// Reply with a heartbeat reply
	err = network.SendMessage(host, pfd.servicePort, raw)
	if err != nil {
		pfd.logger.Panic(err.Error())
	}
}

func (pfd *PerfectFailureDetector) onHeartbeatReply(host string) {

	pfd.logger.Println("Received Heartbeat Reply from host:", host)

	// Mark the host as having replied
	pfd.replied[host] = true
}

func (pfd *PerfectFailureDetector) handleConnection(conn net.Conn) {

	// Close the connection at the end
	defer conn.Close()

	// Get the PFD Message
	rawData, err := ioutil.ReadAll(conn)
	if err != nil {
		pfd.logger.Println(err.Error())
		return
	}

	pfdMessage := &protocol.PFDMessage{}

	err = proto.Unmarshal(rawData, pfdMessage)
	if err != nil {
		pfd.logger.Println(err.Error())
		return
	}

	addr := conn.RemoteAddr().(*net.TCPAddr)

	pfd.logger.Println("Received message from host:", addr)

	// Handle the different message types
	switch pfdMessage.HeartbeatType.(type) {
	case *protocol.PFDMessage_Reply:
		pfd.onHeartbeatReply(addr.IP.String())
		break
	case *protocol.PFDMessage_Request:
		pfd.onHeartbeatRequest(addr.IP.String())
		break
	}
}
