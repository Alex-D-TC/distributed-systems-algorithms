package environment

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
)

type NetworkEnvironment struct {
	hosts    []string
	bebPort  uint16
	onarPort uint16
	pfdPort  uint16
	ucPort   uint16

	logger *log.Logger
}

func NewNetworkEnvironment() (*NetworkEnvironment, error) {

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

	ucPort, err := strconv.Atoi(os.Getenv("UC_PORT"))
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) != 0 {
		return nil, errors.New(strings.Join(errs, "\n"))
	}

	return &NetworkEnvironment{
		hosts:    hosts,
		bebPort:  uint16(bebPort),
		onarPort: uint16(onarPort),
		pfdPort:  uint16(pfdPort),
		ucPort:   uint16(ucPort),
		logger:   log.New(os.Stdout, "[Network Manager]", log.Ltime|log.Ldate),
	}, nil
}

func (env *NetworkEnvironment) GetHosts() []string {
	return env.hosts
}

func (env *NetworkEnvironment) GetBebPort() uint16 {
	return env.bebPort
}

func (env *NetworkEnvironment) GetPFDPort() uint16 {
	return env.pfdPort
}

func (env *NetworkEnvironment) GetONARPort() uint16 {
	return env.onarPort
}

func (env *NetworkEnvironment) GetUCPort() uint16 {
	return env.ucPort
}

func (manager *NetworkEnvironment) SetLogger(logger *log.Logger) {
	manager.logger = logger
}
