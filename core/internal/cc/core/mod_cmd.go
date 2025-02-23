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
	agent := agents.MustGetActiveAgent()
	shortName := "none"
	if agent != nil {
		shortName = strings.Split(agent.Tag, "-agent")[0]
	}

	// build table rows
	rows := [][]string{}
	_, ok := def.Modules[live.ActiveModule.Name]
	if !ok {
		logging.Errorf("Module %s not found", live.ActiveModule.Name)
		return
	}
	rows = append(rows, []string{"module", "Selected module", live.ActiveModule.Name})
	rows = append(rows, []string{"target", "Selected target", shortName})
	for opt_name, opt_obj := range live.ActiveModule.Options {
		help := "N/A"
		if opt_obj == nil {
			continue
		}
		help = opt_obj.Desc
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

func executeModuleOperation(api string, moduleName *string, opt *string, val *string) ([]byte, error) {
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

	url := fmt.Sprintf("%s/%s", OperatorRootURL, api)
	resp, err := sendJSONRequest(url, operation)
	if err != nil {
		logging.Errorf("Failed to execute module operation: %v", err)
	}

	return resp, err
}

func cmdModuleRun(_ *cobra.Command, _ []string) {
	if live.ActiveModule == nil {
		logging.Errorf("No module selected")
		return
	}
	_, err := executeModuleOperation(transport.OperatorModuleRun, &live.ActiveModule.Name, nil, nil)
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
	_, err := executeModuleOperation(transport.OperatorSendCommand, &live.ActiveModule.Name, &opt, &val)
	if err != nil {
		logging.Errorf("Failed to set option: %v", err)
	}
	listModOptionsTable()
}

func cmdSetActiveModule(cmd *cobra.Command, args []string) {
	resp, err := executeModuleOperation(transport.OperatorSetActiveModule, &args[0], nil, nil)
	if err != nil {
		logging.Errorf("Failed to set active module: %v", err)
	}

	mod := new(def.ModuleConfig)
	err = json.Unmarshal(resp, mod)
	if err != nil {
		logging.Errorf("Failed to unmarshal: %v", err)
		return
	}
	live.ActiveModule = mod
	listModOptionsTable()
	logging.Successf("Using %s (by %s on %s): %s", mod.Name, mod.Author, mod.Date, mod.Comment)
}

func cmdListModules(_ *cobra.Command, _ []string) {
	resp, err := executeModuleOperation(transport.OperatorListModules, nil, nil, nil)
	if err != nil {
		logging.Errorf("Failed to list modules: %v", err)
		return
	}
	modules := []*def.ModuleConfig{}
	err = json.Unmarshal(resp, &modules)
	if err != nil {
		logging.Errorf("Failed to unmarshal: %v", err)
		return
	}
	for _, mod := range modules {
		def.Modules[mod.Name] = mod
	}

	// table output
	rows := [][]string{}
	for _, mod := range modules {
		rows = append(rows, []string{mod.Name, mod.Comment})
	}
	tableStr := cli.BuildTable([]string{"Module", "Description"}, rows)
	cli.AdaptiveTable(tableStr)
	logging.Printf("\n%s", tableStr)
}

func cmdSearchModule(cmd *cobra.Command, args []string) {
	resp, err := executeModuleOperation(transport.OperatorSearchModule, &args[0], nil, nil)
	if err != nil {
		logging.Errorf("Failed to search module: %v", err)
		return
	}
	modules := []*def.ModuleConfig{}
	err = json.Unmarshal(resp, &modules)
	if err != nil {
		logging.Errorf("Failed to unmarshal: %v", err)
		return
	}
	row := [][]string{}
	for _, mod := range modules {
		row = append(row, []string{mod.Name, mod.Comment})
	}
	tableStr := cli.BuildTable([]string{"Module", "Description"}, row)
	cli.AdaptiveTable(tableStr)
	logging.Printf("\n%s", tableStr)
}

func cmdModuleListOptions(_ *cobra.Command, _ []string) {
	listModOptionsTable()
}
