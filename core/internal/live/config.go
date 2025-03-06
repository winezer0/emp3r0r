package live

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jm33-m0/arc"
	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
)

var (
	// IsServer is true if we are running as server
	IsServer = false

	// HOME is the user's home directory
	HOME = ""

	// ActiveAgent selected target
	ActiveAgent *def.Emp3r0rAgent

	// Save the configuration of the current session
	RuntimeConfig = &def.Config{}
	// TmuxPersistence enable debug (-debug)
	TmuxPersistence = false
	// Prefix /usr or /usr/local, can be set through $EMP3R0R_PREFIX
	Prefix = ""
	// EmpWorkSpace workspace directory of emp3r0r
	EmpWorkSpace = ""
	// EmpDataDir prefix/lib/emp3r0r
	EmpDataDir = ""
	// EmpBuildDir prefix/lib/emp3r0r/build
	EmpBuildDir = ""
	// FileGetDir where we save #get files
	FileGetDir = ""
	// EmpConfigFile emp3r0r.json
	EmpConfigFile = ""
	// EmpLogFile ~/.emp3r0r/emp3r0r.log, initialized in logging package
	EmpLogFile = ""

	// EmpConfigTar emp3r0r_operator_config.tar.xz
	EmpConfigTar = ""

	// emp3r0r-cat
	CAT = ""
)

const (
	// Temp where we save temp files
	Temp = "/tmp/emp3r0r/"

	// WWWRoot host static files for agent
	WWWRoot = Temp + "www/"

	// UtilsArchive host utils.tar.xz for agent
	UtilsArchive = WWWRoot + "utils.tar.xz"
)

func cleanupConfig() (err error) {
	dents, err := os.ReadDir(EmpWorkSpace)
	if err != nil {
		return
	}
	for _, d := range dents {
		if strings.HasSuffix(d.Name(), ".json") ||
			strings.HasSuffix(d.Name(), ".pem") ||
			strings.HasSuffix(d.Name(), ".history") {
			err = os.Remove(EmpWorkSpace + "/" + d.Name())
			if err != nil {
				return
			}
		}
	}
	return
}

func DownloadExtractConfig(url string, downloader func(string, string) error) (err error) {
	logging.Infof("Downloading and extracting config from %s to %s", url, EmpConfigTar)
	// download config tarball from server
	err = downloader(url, EmpConfigTar)
	if err != nil {
		return
	}
	// remove existing config files for a clean start
	err = cleanupConfig()
	if err != nil {
		return
	}
	// re-create workspace
	err = SetupFilePaths()
	if err != nil {
		return
	}
	// unarchive config files to workspace
	return arc.Unarchive(EmpConfigTar, HOME)
}

func SetupFilePaths() (err error) {
	HOME, err = os.UserHomeDir()
	if err != nil {
		return
	}
	EmpConfigTar = HOME + "/emp3r0r_operator_config.tar.xz"
	// prefix
	Prefix = os.Getenv("EMP3R0R_PREFIX")
	if Prefix == "" {
		Prefix = "/usr/local"
	}
	// eg. /usr/local/lib/emp3r0r
	EmpDataDir = Prefix + "/lib/emp3r0r"
	EmpBuildDir = EmpDataDir + "/build"
	CAT = EmpDataDir + "/emp3r0r-cat"

	if !util.IsExist(EmpDataDir) {
		return fmt.Errorf("emp3r0r is not installed correctly: %s not found", EmpDataDir)
	}
	if !util.IsExist(CAT) {
		return fmt.Errorf("emp3r0r is not installed correctly: %s not found", CAT)
	}

	// set workspace to ~/.emp3r0r
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("get current user: %v", err)
	}
	EmpWorkSpace = u.HomeDir + "/.emp3r0r"
	FileGetDir = EmpWorkSpace + "/file-get/"
	EmpConfigFile = EmpWorkSpace + "/emp3r0r.json"
	EmpLogFile = EmpWorkSpace + "/emp3r0r.log"
	if !util.IsDirExist(EmpWorkSpace) {
		err = os.MkdirAll(FileGetDir, 0o700)
		if err != nil {
			return fmt.Errorf("mkdir %s: %v", EmpWorkSpace, err)
		}
	}

	// cd to workspace
	err = os.Chdir(EmpWorkSpace)
	if err != nil {
		return fmt.Errorf("cd to workspace %s: %v", EmpWorkSpace, err)
	}

	// Module directories
	ModuleDirs = []string{EmpDataDir + "/modules", EmpWorkSpace + "/modules"}

	return nil
}

// CopyStubs copy agent stubs to ~/.emp3r0r, must be run after SetupFilePaths
func CopyStubs() (err error) {
	// copy stub binaries to ~/.emp3r0r
	stubFiles, err := filepath.Glob(fmt.Sprintf("%s/stub*", EmpBuildDir))
	if err != nil {
		return fmt.Errorf("finding agent stubs: %v", err)
	}
	for _, stubFile := range stubFiles {
		copyErr := util.Copy(stubFile, EmpWorkSpace)
		if copyErr != nil {
			return fmt.Errorf("copying agent stubs: %v", copyErr)
		}
	}
	return nil
}

func ReadJSONConfig() (err error) {
	// read JSON
	jsonData, err := os.ReadFile(EmpConfigFile)
	if err != nil {
		return
	}

	return def.ReadJSONConfig(jsonData, RuntimeConfig)
}

// InitCertsAndConfig generate certs if not found, then generate config file
func InitCertsAndConfig() error {
	// if we are not running as server, return, the certs are already generated
	if !IsServer {
		return nil
	}

	if _, err := os.Stat(transport.CaCrtFile); os.IsNotExist(err) {
		logging.Warningf("CA cert not found, generating a new one")
		_, err := transport.GenCerts(nil, transport.CaCrtFile, transport.CaKeyFile, "", "", true)
		if err != nil {
			return fmt.Errorf("GenCerts: %v", err)
		}
	}

	// generate mTLS cert for operator
	if _, err := os.Stat(transport.OperatorCaCrtFile); os.IsNotExist(err) {
		logging.Warningf("mTLS cert not found, generating a new one")
		// CA cert
		_, err := transport.GenCerts(nil, transport.OperatorCaCrtFile, transport.OperatorCaKeyFile, "", "", true)
		if err != nil {
			return fmt.Errorf("generating operator CA: %v", err)
		}

		// client cert signed by CA
		_, err = transport.GenCerts(nil, transport.OperatorClientCrtFile, transport.OperatorClientKeyFile, transport.OperatorCaKeyFile, transport.OperatorCaCrtFile, false)
		if err != nil {
			return fmt.Errorf("generating operator cert: %v", err)
		}
	}

	return nil
}

func GenC2Certs(hosts_str string) error {
	// generate C2 TLS cert for given host names
	var hosts []string
	hosts = strings.Fields(hosts_str)
	// if C2 server TLS cert not found, generate new ones
	logging.Warningf("C2 TLS cert not found, generating a new one")
	hosts = append(hosts, "127.0.0.1") // sometimes we need to connect to a relay that listens on localhost
	hosts = append(hosts, "localhost") // sometimes we need to connect to a relay that listens on localhost

	// validate host names
	for _, host := range hosts {
		if !netutil.ValidateHostName(host) {
			return fmt.Errorf("invalid host name: %s", host)
		}
	}

	// generate C2 TLS cert
	_, certErr := transport.GenCerts(hosts, transport.ServerCrtFile, transport.ServerKeyFile, transport.CaKeyFile, transport.CaCrtFile, false)
	if certErr != nil {
		return fmt.Errorf("generating C2 TLS cert: %v", certErr)
	}
	// generate operator mTLS cert
	hosts = append(hosts, netutil.WgServerIP)   // add wireguard IP for operator
	hosts = append(hosts, netutil.WgOperatorIP) // add wireguard IP for operator
	_, certErr = transport.GenCerts(hosts, transport.OperatorServerCrtFile, transport.OperatorServerKeyFile, transport.OperatorCaKeyFile, transport.OperatorCaCrtFile, false)
	if certErr != nil {
		return fmt.Errorf("generating operator cert: %v", certErr)
	}

	return nil
}

// LoadConfig load config JSON file
func LoadConfig() error {
	err := LoadCACrt2RuntimeConfig()
	if err != nil {
		return fmt.Errorf("failed to load CA to RuntimeConfig: %v", err)
	}

	if util.IsFileExist(EmpConfigFile) {
		return ReadJSONConfig()
	}
	// init config file using the first host name
	return InitConfigFile("127.0.0.1")
}
