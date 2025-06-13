package cli

import (
	"strings"

	"github.com/fatih/color"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// BuildTable creates and renders a table with the given header and rows using the new tablewriter API.
func BuildTable(header []string, rows [][]string) string {
	builder := &strings.Builder{}

	// Configure colorized renderer with custom colors
	colorCfg := renderer.ColorizedConfig{
		Header: renderer.Tint{
			FG: renderer.Colors{color.FgHiMagenta, color.Bold},
		},
		Column: renderer.Tint{
			FG: renderer.Colors{color.FgHiBlue}, // Default column color
			Columns: []renderer.Tint{
				{FG: renderer.Colors{color.FgHiMagenta}},
				{FG: renderer.Colors{color.FgBlue}},
				{FG: renderer.Colors{color.FgHiWhite}},
				{FG: renderer.Colors{color.FgHiCyan}},
				{FG: renderer.Colors{color.FgYellow}},
			},
		},
		Border:    renderer.Tint{FG: renderer.Colors{color.FgWhite}},
		Separator: renderer.Tint{FG: renderer.Colors{color.FgWhite}},
	}

	table := tablewriter.NewTable(builder,
		tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal},
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				ColMaxWidths: tw.CellWidth{Global: 20},
			},
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.On},
				Alignment:  tw.CellAlignment{Global: tw.AlignCenter},
			},
		}),
	)

	table.Header(header)
	table.Bulk(rows)
	table.Render()
	return builder.String()
}

// automatically resize CommandPane according to table width
func AdaptiveTable(tableString string) {
	TmuxUpdatePanes()
	row_len := len(strings.Split(tableString, "\n")[0])
	if OutputPane.Width < row_len {
		logging.Debugf("Command Pane %d vs %d table width, resizing", CommandPane.Width, row_len)
		OutputPane.ResizePane("x", row_len)
	}
}

// CliPrettyPrint prints two-column help info
func CliPrettyPrint(header1, header2 string, map2write *map[string]string) {
	// build table rows using existing helper to split long lines
	rows := [][]string{}
	for c1, c2 := range *map2write {
		rows = append(rows, []string{
			util.SplitLongLine(c1, 50),
			util.SplitLongLine(c2, 50),
		})
	}
	// reuse BuildTable helper
	tableStr := BuildTable([]string{header1, header2}, rows)

	AdaptiveTable(tableStr)
	logging.Printf("\n%s", tableStr)
}
