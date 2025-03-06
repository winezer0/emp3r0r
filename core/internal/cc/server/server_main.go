package server

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jm33-m0/arc"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/relay"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
)

func ServerMain(wg_port int, hosts string, numOperators int) {
	// start all services
	go KCPC2ListenAndServe()
	go tarConfig(hosts)
	wg(wg_port, numOperators)
	time.Sleep(3 * time.Second)
	go StartC2AgentTLSServer()
	StartOperatorMTLSServer(wg_port + 1)
}

type OperatorConfig struct {
	PrivateKey string
	PublicKey  string
	IP         string
}

func wg(wg_port int, numOperators int) {
	server_privkey, err := netutil.GeneratePrivateKey()
	if err != nil {
		logging.Fatalf("Failed to generate server private key: %v", err)
	}
	server_pubkey, err := netutil.PublicKeyFromPrivate(server_privkey)
	if err != nil {
		logging.Fatalf("Failed to generate server public key: %v", err)
	}

	// network address
	subnet := netutil.GenerateRandomPrivateSubnet24()
	netutil.WgServerIP, _ = netutil.GenerateRandomIPInSubnet24(subnet)

	// Generate operator configs
	operators := make([]OperatorConfig, numOperators)
	peers := make([]netutil.PeerConfig, numOperators)

	for i := range numOperators {
		operator_privkey, err := netutil.GeneratePrivateKey()
		if err != nil {
			logging.Fatalf("Failed to generate operator private key: %v", err)
		}
		operator_pubkey, err := netutil.PublicKeyFromPrivate(operator_privkey)
		if err != nil {
			logging.Fatalf("Failed to generate operator public key: %v", err)
		}
		operatorIP, _ := netutil.GenerateRandomIPInSubnet24(subnet)

		// Save for the first operator (backward compatibility)
		if i == 0 {
			netutil.WgOperatorIP = operatorIP
		}

		operators[i] = OperatorConfig{
			PrivateKey: operator_privkey,
			PublicKey:  operator_pubkey,
			IP:         operatorIP,
		}

		peers[i] = netutil.PeerConfig{
			PublicKey:  operator_pubkey,
			AllowedIPs: operatorIP + "/32",
		}
	}

	wgConfig := netutil.WireGuardConfig{
		IPAddress:     netutil.WgServerIP + "/24",
		InterfaceName: "emp_server",
		ListenPort:    wg_port,
		PrivateKey:    server_privkey,
		Peers:         peers,
	}
	go func() {
		netutil.WgServer, err = netutil.WireGuardMain(wgConfig)
		if err != nil {
			logging.Fatalf("Failed to start WireGuard server: %v", err)
		}
	}()

	// Create server config table
	headers := []string{"Parameter", "Value"}
	rows := [][]string{
		{"C2 Server IP (WG)", netutil.WgServerIP},
		{"C2 Server Port", strconv.Itoa(wg_port)},
		{"C2 Public Key", server_pubkey},
	}

	// Build the server table
	serverTableStr := cli.BuildTable(headers, rows)

	// Create operator config table
	opHeaders := []string{"Operator ID", "IP Address", "Private Key", "Public Key"}
	opRows := make([][]string, numOperators)

	for i, op := range operators {
		opRows[i] = []string{
			strconv.Itoa(i + 1),
			op.IP,
			op.PrivateKey,
			op.PublicKey,
		}
	}

	// Build the operators table
	operatorsTableStr := cli.BuildTable(opHeaders, opRows)

	// Print the tables with titles
	logging.Successf("\n══════════════════ WireGuard Server Configuration ════════════════════════════\n\n%s\n", serverTableStr)
	logging.Successf("\n══════════════════ WireGuard Operator Configurations ════════════════════════════\n\n%s\n", operatorsTableStr)
}

func tarConfig(hosts string) {
	err := live.GenC2Certs(hosts)
	if err != nil {
		logging.Fatalf("Failed to generate C2 certs: %v", err)
	}
	err = os.Chdir(live.EmpWorkSpace)
	if err != nil {
		logging.Fatalf("Failed to change directory: %v", err)
	}
	// tar all config files
	filter := func(path string) bool {
		return strings.HasSuffix(path, ".log") || strings.HasPrefix(path, "stub") || strings.HasSuffix(path, ".history")
	}
	os.Chdir(filepath.Dir(live.EmpWorkSpace))
	defer os.Chdir(live.EmpWorkSpace)

	err = arc.ArchiveWithFilter(filepath.Base(live.EmpWorkSpace), live.EmpConfigTar, arc.CompressionMap["xz"], arc.ArchivalMap["tar"], filter)
	if err != nil {
		logging.Errorf("Failed to tar config files: %v", err)
	}
	err = relay.WgFileServer(live.EmpConfigTar)
	if err != nil {
		logging.Errorf("Failed to start file server to serve config tarball: %v", err)
	}
}
