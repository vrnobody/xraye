package api

import (
	"fmt"

	routerService "github.com/xtls/xray-core/app/router/command"
	cserial "github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/infra/conf/serial"
	"github.com/xtls/xray-core/main/commands/base"
)

var cmdSetRoutingConfig = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api setr [--server=127.0.0.1:8080] <c1.json>",
	Short:       "Set routing config",
	Long: `
Change routing config for Xray.
Arguments:
	-s, -server 
		The API server address. Default 127.0.0.1:8080
	-t, -timeout
		Timeout seconds to call API. Default 3
Example:
    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c1.json
`,
	Run: executeSetRoutingConfig,
}

func executeSetRoutingConfig(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	unnamedArgs := cmd.Flag.Args()
	if len(unnamedArgs) == 0 {
		fmt.Println("Reading from STDIN")
		unnamedArgs = []string{"stdin:"}
	}

	routes := make([]conf.RouterConfig, 0)
	for _, arg := range unnamedArgs {
		r, err := loadArg(arg)
		if err != nil {
			base.Fatalf("failed to load %s: %s", arg, err)
		}
		conf, err := serial.DecodeJSONConfig(r)
		if err != nil {
			base.Fatalf("failed to decode %s: %s", arg, err)
		}
		routes = append(routes, *conf.RouterConfig)
	}
	if len(routes) == 0 {
		base.Fatalf("no valid routing found")
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	client := routerService.NewRoutingServiceClient(conn)

	fmt.Println("replacing routing config")
	o, err := routes[0].Build()
	if err != nil {
		base.Fatalf("failed to build conf: %s", err)
	}
	tmsg := cserial.ToTypedMessage(o)
	if tmsg == nil {
		base.Fatalf("failed to format config to TypedMessage.")
	}
	r := &routerService.SetRoutingConfigRequest{
		Config: tmsg,
	}
	resp, err := client.SetRoutingConfig(ctx, r)
	if err != nil {
		base.Fatalf("failed to set routing config: %s", err)
	}
	showJSONResponse(resp)
}
