package service

import (
	"github.com/xtls/xray-core/main/commands/all/service/latency"
	"github.com/xtls/xray-core/main/commands/base"
)

// CmdConvert do config convertion
var CmdService = &base.Command{
	UsageLine: "{{.Exec}} service",
	Short:     "Provide all kinds of services.",
	Long: `{{.Exec}} {{.LongName}} run as server to provide services.
`,
	Commands: []*base.Command{
		latency.CmdProber,
		latency.CmdLatency,
	},
}
