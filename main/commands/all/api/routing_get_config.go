package api

import (
	"fmt"

	routerService "github.com/xtls/xray-core/app/router/command"
	"github.com/xtls/xray-core/main/commands/base"

	creflect "github.com/xtls/xray-core/common/reflect"
)

var cmdGetRoutingConfig = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api getr [--server=127.0.0.1:8080]",
	Short:       "Get routing config",
	Long: `
Get routing config from Xray.
Arguments:
	-s, -server
		The API server address. Default 127.0.0.1:8080
	-t, -timeout
		Timeout seconds to call API. Default 3
Example:
    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080
`,
	Run: executeGetRoutingConfig,
}

func executeGetRoutingConfig(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	conn, ctx, close := dialAPIServer()
	defer close()

	client := routerService.NewRoutingServiceClient(conn)
	req := &routerService.GetRoutingConfigRequest{}
	resp, err := client.GetRoutingConfig(ctx, req)
	if err != nil {
		base.Fatalf("failed to get routing config: %s", err)
	}
	if j, ok := creflect.MarshalToJson(resp.Config); !ok {
		base.Fatalf("failed to marshal configs to json")
	} else {
		fmt.Print(j)
	}
}
