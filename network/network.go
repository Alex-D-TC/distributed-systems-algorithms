package network

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type NetworkManager struct {
	hosts    []string
	bebPort  int16
	onarPort int16
	pfdPort  int16
	ucPort   int16

	logger *log.Logger
}

func NewNetworkManager() (*NetworkManager, error) {

	// Validate environment variable data
	errs := []string{}

	hosts := strings.Split(os.Getenv("HOSTS"), ";")

	bebPort, err := strconv.Atoi(os.Getenv("BEB_PORT"))
	if err != nil {
		errs = append(errs, err.Error())
	}

	onarPort, err := strconv.Atoi(os.Getenv("ONAR_PORT"))
	if err != nil {
		errs = append(errs, err.Error())
	}

	pfdPort, err := strconv.Atoi(os.Getenv("PFD_PORT"))
	if err != nil {
		errs = append(errs, err.Error())
	}

	ucPort, err := strconv.Atoi(os.Getenv("PFD_PORT"))
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) != 0 {
		return nil, errors.New(strings.Join(errs, "\n"))
	}

	return &NetworkManager{
		hosts:    hosts,
		bebPort:  int16(bebPort),
		onarPort: int16(onarPort),
		pfdPort:  int16(pfdPort),
		ucPort:   int16(ucPort),
		logger:   log.New(os.Stdout, "[Network Manager]", log.Ltime|log.Ldate),
	}, nil
}

func (manager *NetworkManager) SetLogger(logger *log.Logger) {
	manager.logger = logger
}

func (manager *NetworkManager) SendPFDMessage(hostID int, message []byte) error {
	return manager.sendMessage(hostID, "pfd", message)
}

func (manager *NetworkManager) SendBebMessage(hostID int, message []byte) error {
	return manager.sendMessage(hostID, "beb", message)
}

func (manager *NetworkManager) StartBEBListener(handler func(net.Conn)) {
	listen(manager.bebPort, handler, manager.logger)
}

func (manager *NetworkManager) StartPFDListener(handler func(net.Conn)) {
	listen(manager.pfdPort, handler, manager.logger)
}

func (manager *NetworkManager) sendMessage(hostID int, moduleName string, message []byte) error {

	// Validate the input parameter
	if hostID < 0 || hostID > len(manager.hosts) {
		return fmt.Errorf("The host with the id %d does not exist", hostID)
	}

	modulePort := int16(0)

	switch moduleName {
	case "pfd":
		modulePort = manager.pfdPort
		break
	case "uc":
		modulePort = manager.ucPort
		break
	case "onar":
		modulePort = manager.onarPort
		break
	case "beb":
		modulePort = manager.bebPort
		break
	}

	if modulePort == 0 {
		return errors.New("The destination module name is invalid")
	}

	// Initiate the connection
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", manager.hosts[hostID], modulePort))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send the data
	return sendWithGuarantee(conn, message)
}
