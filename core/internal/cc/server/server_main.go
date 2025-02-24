package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jm33-m0/emp3r0r/core/internal/cc/modules"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
)

var (
	Emp3r0rWorkingDir string
	ConfigTar         string
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
	StartMTLSServer(port)
}

func tarConfig() {
	shm := "/dev/shm"
	// copy working dir
	dst := fmt.Sprintf("%s/%s", shm, filepath.Base(Emp3r0rWorkingDir))
	if util.IsExist(dst) {
		os.RemoveAll(dst)
	}
	os.MkdirAll(dst, 0755)
	err := util.Copy(Emp3r0rWorkingDir, dst)
	if err != nil {
		logging.Fatalf("Failed to copy working dir to %s: %v", shm, err)
	}
	os.Chdir(shm)
	defer os.Chdir(Emp3r0rWorkingDir)

	// tar all config files
	err = util.TarXZ(filepath.Base(dst), ConfigTar)
	if err != nil {
		logging.Errorf("Failed to tar config files: %v", err)
	}
}
