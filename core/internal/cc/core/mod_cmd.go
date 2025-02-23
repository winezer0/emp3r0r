package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/agents"
	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
	"github.com/spf13/cobra"
)

// listModOptionsTable list currently available options for `set`, in a table
func listModOptionsTable() {
	if live.ActiveModule == nil {
		logging.Warningf("No module selected")
		return
	}
	opts := make(map[string]string)

	agent := agents.MustGetActiveAgent()
	opts["target"] = "<blank>"

	opts["module"] = live.ActiveModule.Name
	if agent != nil {
		shortName := strings.Split(agent.Tag, "-agent")[0]
		opts["target"] = shortName
	} else {
		opts["target"] = "<blank>"
	}

	for opt_name, opt := range live.ActiveModule.Options {
		if opt != nil {
			opts[opt_name] = opt.Name
		}
	}

	// build table rows
	rows := [][]string{}
	_, ok := def.Modules[live.ActiveModule.Name]
	if !ok {
		logging.Errorf("Module %s not found", live.ActiveModule.Name)
		return
	}
	for opt_name, opt_obj := range live.ActiveModule.Options {
		help := "N/A"
		if opt_obj == nil {
			continue
		}
		help = opt_obj.Desc
		switch opt_name {
		case "module":
			help = "Selected module"
		case "target":
			help = "Selected target"
		}
		val := ""
		currentOpt, ok := live.ActiveModule.Options[opt_name]
		if ok {
			val = currentOpt.Val
		}

		rows = append(rows,
			[]string{
				util.SplitLongLine(opt_name, 50),
				util.SplitLongLine(help, 50),
				util.SplitLongLine(val, 50),
			})
	}

	// reuse BuildTable helper
	tableStr := cli.BuildTable([]string{"Option", "Help", "Value"}, rows)
	cli.AdaptiveTable(tableStr)
	logging.Printf("\n%s", tableStr)
}

func executeModuleOperation(moduleName *string, opt *string, val *string) (*def.ModuleConfig, error) {
	agent := agents.MustGetActiveAgent()
	if agent == nil {
		logging.Errorf("No active agent")
		return nil, fmt.Errorf("no active agent")
	}
	operation := def.Operation{
		AgentTag:   agent.Tag,
		Action:     "module",
		ModuleName: moduleName,
		SetOption:  opt,
		SetValue:   val,
	}

	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorSendCommand)
	resp, err := sendJSONRequest(url, operation)
	if err != nil {
		logging.Errorf("Failed to execute module operation: %v", err)
	}
	// decode response
	mod := new(def.ModuleConfig)
	err = json.Unmarshal(resp, mod)

	return mod, err
}

func getModuleOptions() {
	if live.ActiveModule == nil {
		logging.Errorf("No module selected")
		return
	}
	mod, err := executeModuleOperation(&live.ActiveModule.Name, nil, nil)
	if err != nil {
		logging.Errorf("Failed to get module options: %v", err)
		return
	}
	live.ActiveModule = mod
	listModOptionsTable()
}

func cmdModuleRun(_ *cobra.Command, _ []string) {
	if live.ActiveModule == nil {
		logging.Errorf("No module selected")
		return
	}
	_, err := executeModuleOperation(&live.ActiveModule.Name, nil, nil)
	if err != nil {
		logging.Errorf("Failed to run module: %v", err)
	}
}

func cmdSetOptVal(cmd *cobra.Command, args []string) {
	if live.ActiveModule == nil {
		logging.Errorf("No module selected")
		return
	}
	opt := args[0]
	val := args[1]

	// hand to SetOption helper
	live.SetOption(opt, val)

	// send to C2 server to sync
	_, err := executeModuleOperation(&live.ActiveModule.Name, &opt, &val)
	if err != nil {
		logging.Errorf("Failed to set option: %v", err)
	}
	listModOptionsTable()
}

func cmdSetActiveModule(cmd *cobra.Command, args []string) {
	_, err := executeModuleOperation(&args[0], nil, nil)
	if err != nil {
		logging.Errorf("Failed to set active module: %v", err)
	}
}

func cmdListModules(_ *cobra.Command, _ []string) {
	executeModuleOperation(nil, nil, nil)
	// TODO: handle response
}

func cmdSearchModule(cmd *cobra.Command, args []string) {
	executeModuleOperation(&args[0], nil, nil)
	// TODO: handle response
}

func cmdModuleListOptions(_ *cobra.Command, _ []string) {
	executeModuleOperation(nil, nil, nil)
	// TODO: handle response
}
