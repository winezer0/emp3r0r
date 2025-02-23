package core

import (
	"encoding/json"

	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/spf13/cobra"
)

// CmdLsModules list all available modules
func CmdLsModules(_ *cobra.Command, _ []string) {
	mod_comment_map := make(map[string]string)
	for mod_name, mod := range def.Modules {
		mod_comment_map[mod_name] = mod.Comment
	}
	cli.CliPrettyPrint("Module Name", "Help", &mod_comment_map)
}

func initModules() {
	// request operator server for available modules
	// and store them in def.Modules
	resp, err := executeModuleOperation(transport.OperatorListModules, nil, nil, nil)
	if err != nil {
		logging.Errorf("Failed to list modules: %v", err)
		return
	}
	modules := make([]def.ModuleConfig, 0)
	err = json.Unmarshal(resp, &modules)
	if err != nil {
		logging.Errorf("Failed to unmarshal modules: %v", err)
		return
	}

	// store modules in def.Modules
	for _, mod := range modules {
		def.Modules[mod.Name] = &mod
	}
}
