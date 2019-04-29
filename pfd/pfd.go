package pfd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/alex-d-tc/distributed-systems-algorithms/network"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/golang/protobuf/proto"
)

// PerfectFailureDetector is a class which is responsible for analyzing the network and detecting possible failed nodes.
// Other services can register to the onProcessCrashed event by submitting a channel to send the crashed process hostname to,
//	through the AddOnProcessCrashedListener method
// The assumptions made in this class are that each node in the network starts the PerfectFailureDetector on the same servicePort
// 	and that each host in the hosts array starts a PerfectFailureDetector.
type PerfectFailureDetector struct {
	servicePort uint16
	hosts       []string

	// PFD Main functionality
	dead           map[string]bool
	replyFailed    map[string]uint8
	replied        map[string]bool
	delta          time.Duration
	maxReplyFailed uint8

	// PFD Goroutine sync
	deadMapAccessToken    chan bool
	repliedMapAccessToken chan bool

	// Outgoing event listeners
	onProcessCrashListeners []chan<- string

	// Logging
	logger *log.Logger
}

// NewPerfectFailureDetector creates a new perfect failure detector
// The PerfectFailureDetector class starts two goroutines. One performs periodic heartbeat calls, and another sets up a server
// 	to listen for incoming heartbeat requests
func NewPerfectFailureDetector(port uint16, hosts []string, delta time.Duration, maxReplyFailed uint8) *PerfectFailureDetector {

	pfd := &PerfectFailureDetector{
		servicePort:           port,
		hosts:                 hosts,
		delta:                 delta,
		deadMapAccessToken:    make(chan bool, 1),
		repliedMapAccessToken: make(chan bool, 1),
		dead:                  map[string]bool{},
		replied:               map[string]bool{},
		replyFailed:           map[string]uint8{},
		maxReplyFailed:        maxReplyFailed,
		logger:                log.New(os.Stdout, "[PFD]", log.Ldate|log.Ltime),
	}

	pfd.deadMapAccessToken <- true
	pfd.repliedMapAccessToken <- true

	// Start listening routine
	go network.Listen(port, pfd.handleConnection, pfd.logger)

	// Start the timer routine
	go pfd.periodicCheck()

	return pfd
}

// AddOnProcessCrashedListener adds a listener to the list of ProcessCrashed event listeners
// Each listener is a channel which will receive the hostname of the crashed process
// The listener will receive one hostname for each process that crashed
func (pfd *PerfectFailureDetector) AddOnProcessCrashedListener(listener chan<- string) {
	pfd.onProcessCrashListeners = append(pfd.onProcessCrashListeners, listener)

	// Send all processes currently known to be crashed to the listener
	<-pfd.deadMapAccessToken
	for host, dead := range pfd.dead {
		if dead {
			listener <- host
		}
	}
	pfd.deadMapAccessToken <- true
}

func (pfd *PerfectFailureDetector) periodicCheck() {

	for {

		// Send the new pfd requests
		pfd.pingLivingHosts()

		// Wait for the replies
		time.Sleep(pfd.delta)

		// Find out who died
		for _, host := range pfd.hosts {

			// Skip dead hosts
			if pfd.dead[host] {
				continue
			}

			<-pfd.repliedMapAccessToken
			hostReplied := pfd.replied[host]
			pfd.repliedMapAccessToken <- true

			if !hostReplied {

				pfd.replyFailed[host]++
				pfd.logger.Println(fmt.Sprintf("Reply failure count of host %s: %d", host, pfd.replyFailed[host]))

				if pfd.replyFailed[host] >= pfd.maxReplyFailed {

					// Notify the process crashed listeners
					<-pfd.deadMapAccessToken
					pfd.dead[host] = true
					pfd.deadMapAccessToken <- true

					for _, listener := range pfd.onProcessCrashListeners {
						listener <- host
					}
				}
			} else {
				// Reset the counter if the host has replied
				pfd.replyFailed[host] = 0
			}
		}
	}
}

func (pfd *PerfectFailureDetector) pingLivingHosts() {

	<-pfd.repliedMapAccessToken
	pfd.replied = map[string]bool{}
	pfd.repliedMapAccessToken <- true

	pfdRequestMessage := &protocol.PFDMessage{
		HeartbeatType: &protocol.PFDMessage_Request{
			Request: &protocol.HeartbeatRequest{},
		},
	}

	rawPfdMessage, err := proto.Marshal(pfdRequestMessage)
	if err != nil {
		pfd.logger.Panic(err.Error())
	}

	for _, host := range pfd.hosts {
		if !pfd.dead[host] {
			err = network.SendMessage(host, pfd.servicePort, rawPfdMessage)
			if err != nil {
				pfd.logger.Println(err.Error())
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
		pfd.logger.Println(err.Error())
	}
}

func (pfd *PerfectFailureDetector) onHeartbeatReply(host string) {

	pfd.logger.Println("Received Heartbeat Reply from host:", host)

	// Mark the host as having replied
	<-pfd.repliedMapAccessToken
	pfd.replied[host] = true
	pfd.repliedMapAccessToken <- true
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
