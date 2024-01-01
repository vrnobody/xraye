package api

import (
	"fmt"
	"strings"

	handlerService "github.com/xtls/xray-core/app/proxyman/command"
	cserial "github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/main/commands/base"
)

var cmdGetUsers = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api getu [--server=127.0.0.1:8080] tag",
	Short:       "Get all IDs from inbound",
	Long: `
Get all users info from inbound by tag.
Arguments:
	-s, -server 
		The API server address. Default 127.0.0.1:8080
	-t, -timeout
		Timeout seconds to call API. Default 3
Example:
    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 "vmessin"
`,
	Run: executeGetUsers,
}

func executeGetUsers(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	unnamedArgs := cmd.Flag.Args()
	if len(unnamedArgs) < 1 {
		base.Fatalf("please provide an inbound tag")
	}
	tag := unnamedArgs[0]
	conn, ctx, close := dialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	resp, err := client.QueryInbound(ctx, &handlerService.QueryInboundRequest{
		Tag:       tag,
		Operation: cserial.ToTypedMessage(&handlerService.GetUsersOperation{}),
	})

	if err != nil {
		base.Fatalf("%s\n", err)
	} else {
		fmt.Printf("[%s]", strings.Join(resp.Content, ",\n"))
	}
}
