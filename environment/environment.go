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
	bebPort  int16
	onarPort int16
	pfdPort  int16
	ucPort   int16

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
		bebPort:  int16(bebPort),
		onarPort: int16(onarPort),
		pfdPort:  int16(pfdPort),
		ucPort:   int16(ucPort),
		logger:   log.New(os.Stdout, "[Network Manager]", log.Ltime|log.Ldate),
	}, nil
}

func (env *NetworkEnvironment) GetHosts() []string {
	return env.hosts
}

func (env *NetworkEnvironment) GetBebPort() int16 {
	return env.bebPort
}

func (env *NetworkEnvironment) GetPFDPort() int16 {
	return env.pfdPort
}

func (env *NetworkEnvironment) GetONARPort() int16 {
	return env.onarPort
}

func (env *NetworkEnvironment) GetUCPort() int16 {
	return env.ucPort
}

func (manager *NetworkEnvironment) SetLogger(logger *log.Logger) {
	manager.logger = logger
}
