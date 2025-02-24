package transport

import (
	"fmt"
	"os"
)

var (
	CaCrtFile             string // Path to CA cert file
	CaKeyFile             string // Path to CA key file
	ServerCrtFile         string // Path to server cert file
	ServerKeyFile         string // Path to server key file
	ServerPubKey          string // PEM encoded server public key
	OperatorCaCrtFile     string // Path to operator CA cert file
	OperatorCaKeyFile     string // Path to operator CA key file
	OperatorServerCrtFile string // Path to operator cert file
	OperatorServerKeyFile string // Path to operator key file
	OperatorClientCrtFile string // operator client mTLS cert
	OperatorClientKeyFile string // operator client mTLS key
	EmpWorkSpace          string // Path to emp3r0r workspace
	CACrtPEM              []byte // CA cert in PEM format
)

func init() {
	// Paths
	EmpWorkSpace = fmt.Sprintf("%s/.emp3r0r", os.Getenv("HOME"))
	CaCrtFile = fmt.Sprintf("%s/ca-cert.pem", EmpWorkSpace)
	CaKeyFile = fmt.Sprintf("%s/ca-key.pem", EmpWorkSpace)
	ServerCrtFile = fmt.Sprintf("%s/server-cert.pem", EmpWorkSpace)
	ServerKeyFile = fmt.Sprintf("%s/server-key.pem", EmpWorkSpace)
	OperatorCaCrtFile = fmt.Sprintf("%s/operator-ca-cert.pem", EmpWorkSpace)
	OperatorCaKeyFile = fmt.Sprintf("%s/operator-ca-key.pem", EmpWorkSpace)
	OperatorServerCrtFile = fmt.Sprintf("%s/operator-server-cert.pem", EmpWorkSpace)
	OperatorServerKeyFile = fmt.Sprintf("%s/operator-server-key.pem", EmpWorkSpace)
	OperatorClientCrtFile = fmt.Sprintf("%s/operator-client-cert.pem", EmpWorkSpace)
	OperatorClientKeyFile = fmt.Sprintf("%s/operator-client-key.pem", EmpWorkSpace)
}

// LoadCACrt load CA cert from file
func LoadCACrt() error {
	ca_data, err := os.ReadFile(CaCrtFile)
	if err != nil {
		return err
	}
	CACrtPEM = ca_data
	return nil
}
