package core

import (
	"fmt"
	"strings"

	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// cmdListModOptionsTable list currently available options for `set`, in a table
func cmdListModOptionsTable(_ *cobra.Command, _ []string) {
	if live.ActiveModule == "none" {
		logging.Warningf("No module selected")
		return
	}
	live.AgentControlMapMutex.RLock()
	defer live.AgentControlMapMutex.RUnlock()
	opts := make(map[string]string)

	opts["module"] = live.ActiveModule
	if live.ActiveAgent != nil {
		_, exist := live.AgentControlMap[live.ActiveAgent]
		if exist {
			shortName := strings.Split(live.ActiveAgent.Tag, "-agent")[0]
			opts["target"] = shortName
		} else {
			opts["target"] = "<blank>"
		}
	} else {
		opts["target"] = "<blank>"
	}

	for opt_name, opt := range live.AvailableModuleOptions {
		if opt != nil {
			opts[opt_name] = opt.Name
		}
	}

	// build table
	tdata := [][]string{}
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Option", "Help", "Value"})
	table.SetBorder(true)
	table.SetRowLine(true)
	table.SetAutoWrapText(true)
	table.SetColWidth(50)

	// color
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor})
	table.SetColumnColor(tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.FgBlueColor})

	// fill table
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

		tdata = append(tdata,
			[]string{
				util.SplitLongLine(opt_name, 50),
				util.SplitLongLine(help, 50),
				util.SplitLongLine(val, 50),
			})
	}
	table.AppendBulk(tdata)
	table.Render()
	out := tableString.String()
	cli.AdaptiveTable(out)
	logging.Printf("\n%s", out)
}

func cmdModuleRun(_ *cobra.Command, _ []string) {
	operation := def.Operation{
		AgentTag:   live.ActiveAgent.Tag,
		Action:     "module",
		ModuleName: &live.ActiveModule,
	}

	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorModuleRun)
	if err := sendJSONRequest(url, operation); err != nil {
		logging.Errorf("Failed to run module: %v", err)
	}
}

func cmdSetOptVal(cmd *cobra.Command, args []string) {
	opt := args[0]
	val := args[1]

	// hand to SetOption helper
	live.SetOption(opt, val)
	cmdListModOptionsTable(cmd, args)

	operation := def.Operation{
		AgentTag:   live.ActiveAgent.Tag,
		Action:     "module",
		ModuleName: &live.ActiveModule,
		SetOption:  &opt,
		SetValue:   &val,
	}

	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorSendCommand)
	if err := sendJSONRequest(url, operation); err != nil {
		logging.Errorf("Failed to set option value: %v", err)
	}
}

func cmdSetActiveModule(cmd *cobra.Command, args []string) {
	operation := def.Operation{
		AgentTag:   live.ActiveAgent.Tag,
		Action:     "module",
		ModuleName: &args[0],
	}

	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorSetActiveModule)
	if err := sendJSONRequest(url, operation); err != nil {
		logging.Errorf("Failed to set active module: %v", err)
	}
}
