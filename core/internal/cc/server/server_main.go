package server

import "github.com/jm33-m0/emp3r0r/core/internal/cc/modules"

func ServerMain() {
	// start all services
	go StartTLSServer()
	go KCPC2ListenAndServe()
	go modules.InitModules()
}
