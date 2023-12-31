package api

import (
	"context"
	"fmt"

	"github.com/xtls/xray-core/common/protocol"

	handlerService "github.com/xtls/xray-core/app/proxyman/command"
	cserial "github.com/xtls/xray-core/common/serial"

	"github.com/xtls/xray-core/main/commands/base"
)

var cmdRemoveUsers = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api rmu [--server=127.0.0.1:8080] <c.json> [c2.json] ...",
	Short:       "Remove users from inbounds",
	Long: `
Remove users from inbounds.
Arguments:
	-s, -server 
		The API server address. Default 127.0.0.1:8080
	-t, -timeout
		Timeout seconds to call API. Default 3
Example:
    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c.json [c2.json] ...
`,
	Run: executeRemoveUsers,
}

func executeRemoveUsers(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	unnamedArgs := cmd.Flag.Args()

	ins := loadInboundsFromConfigFiles(unnamedArgs)

	conn, ctx, close := dialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, in := range ins {
		fmt.Println("inbound:", in.Tag)
		i, err := in.Build()
		if err != nil {
			fmt.Println("failed to build conf:", err)
		} else if users := getUsersFromInbound(i); users != nil {
			removeUsers(ctx, client, in.Tag, users)
		}
		fmt.Println()
	}
}

func removeUsers(ctx context.Context, client handlerService.HandlerServiceClient, tag string, users []*protocol.User) {
	for _, user := range users {
		fmt.Println("remove user:", user.Email)
		_, err := client.AlterInbound(ctx, &handlerService.AlterInboundRequest{
			Tag: tag,
			Operation: cserial.ToTypedMessage(
				&handlerService.RemoveUserOperation{
					Email: user.Email,
				}),
		})
		if err != nil {
			fmt.Println("err:", err)
		} else {
			fmt.Println("ok.")
		}
	}
}
