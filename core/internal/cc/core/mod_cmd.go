package core

import (
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
func listModOptionsTable(_ *cobra.Command, _ []string) {
	if live.ActiveModule == "none" {
		logging.Warningf("No module selected")
		return
	}
	live.AgentControlMapMutex.RLock()
	defer live.AgentControlMapMutex.RUnlock()
	opts := make(map[string]string)

	agent := agents.MustGetActiveAgent()

	opts["module"] = live.ActiveModule
	_, exist := live.AgentControlMap[agent]
	if exist {
		shortName := strings.Split(agent.Tag, "-agent")[0]
		opts["target"] = shortName
	} else {
		opts["target"] = "<blank>"
	}
	if agent == nil {
		opts["target"] = "<blank>"
	}

	for opt_name, opt := range live.AvailableModuleOptions {
		if opt != nil {
			opts[opt_name] = opt.Name
		}
	}

	// build table rows
	rows := [][]string{}
	module_obj := def.Modules[live.ActiveModule]
	if module_obj == nil {
		logging.Errorf("Module %s not found", live.ActiveModule)
		return
	}
	for opt_name, opt_obj := range live.AvailableModuleOptions {
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
		currentOpt, ok := live.AvailableModuleOptions[opt_name]
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

func executeModuleOperation(action string, moduleName *string, opt *string, val *string) {
	agent := agents.MustGetActiveAgent()
	operation := def.Operation{
		AgentTag:   agent.Tag,
		Action:     action,
		ModuleName: moduleName,
		SetOption:  opt,
		SetValue:   val,
	}

	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorSendCommand)
	if err := sendJSONRequest(url, operation); err != nil {
		logging.Errorf("Failed to execute module operation: %v", err)
	}
}

func cmdModuleRun(_ *cobra.Command, _ []string) {
	executeModuleOperation("module", &live.ActiveModule, nil, nil)
}

func cmdSetOptVal(cmd *cobra.Command, args []string) {
	opt := args[0]
	val := args[1]

	// hand to SetOption helper
	live.SetOption(opt, val)
	listModOptionsTable(cmd, args)

	// send to C2 server to sync
	executeModuleOperation("module", &live.ActiveModule, &opt, &val)
}

func cmdSetActiveModule(cmd *cobra.Command, args []string) {
	executeModuleOperation("module", &args[0], nil, nil)
}

func cmdListModules(_ *cobra.Command, _ []string) {
	executeModuleOperation("module", nil, nil, nil)
	// TODO: handle response
}

func cmdSearchModule(cmd *cobra.Command, args []string) {
	executeModuleOperation("module", &args[0], nil, nil)
	// TODO: handle response
}

func cmdModuleListOptions(_ *cobra.Command, _ []string) {
	executeModuleOperation("module", nil, nil, nil)
	// TODO: handle response
}
