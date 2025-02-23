package server

import (
	"fmt"
	"os"

	"github.com/jm33-m0/emp3r0r/core/internal/cc/modules"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
)

var (
	Emp3r0rWorkingDir string
	ConfigTar         string
)

func init() {
	// logging
	logging.SetOutput(os.Stderr)

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
	go StartTLSServer()
	go KCPC2ListenAndServe()
	go modules.InitModules()
	go tarConfig()
	StartMTLSServer(port)
}

func tarConfig() {
	// tar all config files
	err := util.TarXZ(Emp3r0rWorkingDir, ConfigTar)
	if err != nil {
		logging.Errorf("Failed to tar config files: %v", err)
	}
}
