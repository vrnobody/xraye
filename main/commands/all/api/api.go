package api

import (
	"github.com/xtls/xray-core/main/commands/base"
)

// CmdAPI calls an API in an Xray process
var CmdAPI = &base.Command{
	UsageLine: "{{.Exec}} api",
	Short:     "Call an API in an Xray process",
	Long: `{{.Exec}} {{.LongName}} provides tools to manipulate Xray via its API.
`,
	Commands: []*base.Command{
		cmdRestartLogger,
		cmdGetStats,
		cmdQueryStats,
		cmdSysStats,
		cmdGetAllInbounds,
		cmdGetAllOutbounds,
		cmdBalancerInfo,
		cmdBalancerOverride,
		cmdAddInbounds,
		cmdAddOutbounds,
		cmdRemoveInbounds,
		cmdRemoveOutbounds,
		cmdGetRoutingConfig,
		cmdSetRoutingConfig,
		cmdGetUsers,
		cmdAddUsers,
		cmdRemoveUsers,
	},
}
