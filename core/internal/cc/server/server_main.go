package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jm33-m0/arc"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/modules"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
)

var (
	Emp3r0rWorkingDir string
	ConfigTar         string
	OPERATOR_ADDR     string
	OPERATOR_PORT     int
)

func init() {
	// paths
	home, err := os.UserHomeDir()
	if err != nil {
		logging.Fatalf("Failed to get user home directory: %v", err)
	}
	Emp3r0rWorkingDir = fmt.Sprintf("%s/.emp3r0r", home)
	ConfigTar = fmt.Sprintf("%s/emp3r0r.tar.xz", home)
}

func ServerMain(port int) {
	// start all services
	go StartC2TLSServer()
	go KCPC2ListenAndServe()
	go modules.InitModules()
	go tarConfig()
	OPERATOR_PORT = port
	StartMTLSServer(port)
}

func tarConfig() {
	// tar all config files
	filter := func(path string) bool {
		return strings.HasSuffix(path, ".log")
	}
	os.Chdir(filepath.Dir(Emp3r0rWorkingDir))
	defer os.Chdir(Emp3r0rWorkingDir)

	err := arc.ArchiveWithFilter(filepath.Base(Emp3r0rWorkingDir), ConfigTar, arc.CompressionMap["xz"], arc.ArchivalMap["tar"], filter)
	if err != nil {
		logging.Errorf("Failed to tar config files: %v", err)
	}
}
