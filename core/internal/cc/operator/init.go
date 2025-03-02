package operator

import (
	"fmt"

	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
)

// InitWG initializes operator's Wireguard connection with server
func InitWG(wg_server_ip string, wg_server_port int) {
	var err error
	OPERATOR_PORT = wg_server_port + 1
	OperatorHTTPClient, err = createMTLSHttpClient()
	if err != nil {
		logging.Fatalf("Failed to create HTTP client: %v", err)
	}
	OPERATOR_ADDR = fmt.Sprintf("%s:%d", netutil.WgServerIP, OPERATOR_PORT)
	logging.Infof("Operator's address: %s", OPERATOR_ADDR)

	// Update operator's IP to Wireguard IP
	OperatorRootURL = fmt.Sprintf("https://%s", OPERATOR_ADDR)
	logging.Infof("Operator's WireGuard address: %s", OperatorRootURL)
}
