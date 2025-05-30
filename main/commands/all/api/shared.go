package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/xtls/xray-core/common/protocol"

	"github.com/xtls/xray-core/core"

	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/infra/conf/serial"

	trojanin "github.com/xtls/xray-core/proxy/trojan"
	vlessin "github.com/xtls/xray-core/proxy/vless/inbound"
	vmessin "github.com/xtls/xray-core/proxy/vmess/inbound"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/xtls/xray-core/common/buf"
	creflect "github.com/xtls/xray-core/common/reflect"
	"github.com/xtls/xray-core/main/commands/base"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type serviceHandler func(ctx context.Context, conn *grpc.ClientConn, cmd *base.Command, args []string) string

var (
	apiServerAddrPtr string
	apiTimeout       int
	apiJSON          bool
)

func setSharedFlags(cmd *base.Command) {
	cmd.Flag.StringVar(&apiServerAddrPtr, "s", "127.0.0.1:8080", "")
	cmd.Flag.StringVar(&apiServerAddrPtr, "server", "127.0.0.1:8080", "")
	cmd.Flag.IntVar(&apiTimeout, "t", 3, "")
	cmd.Flag.IntVar(&apiTimeout, "timeout", 3, "")
	cmd.Flag.BoolVar(&apiJSON, "json", false, "")
}

func dialAPIServer() (conn *grpc.ClientConn, ctx context.Context, close func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	close = func() {
		cancel()
		conn.Close()
	}
	return
}

// loadArg loads one arg, maybe an remote url, or local file path
func loadArg(arg string) (out io.Reader, err error) {
	var data []byte
	switch {
	case strings.HasPrefix(arg, "http://"), strings.HasPrefix(arg, "https://"):
		data, err = fetchHTTPContent(arg)

	case arg == "stdin:":
		data, err = io.ReadAll(os.Stdin)

	default:
		data, err = os.ReadFile(arg)
	}

	if err != nil {
		return
	}
	out = bytes.NewBuffer(data)
	return
}

// fetchHTTPContent dials https for remote content
func fetchHTTPContent(target string) ([]byte, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	if s := strings.ToLower(parsedTarget.Scheme); s != "http" && s != "https" {
		return nil, fmt.Errorf("invalid scheme: %s", parsedTarget.Scheme)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    parsedTarget,
		Close:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial to %s", target)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	content, err := buf.ReadAllToBytes(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read HTTP response")
	}

	return content, nil
}

func showJSONResponse(m proto.Message) {
	if isNil(m) {
		return
	}
	if j, ok := creflect.MarshalToJson(m, true); ok {
		fmt.Println(j)
	} else {
		fmt.Fprintf(os.Stdout, "%v\n", m)
		base.Fatalf("error encode json")
	}
}

func isNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return i == nil
}

func getUsersFromInbound(i *core.InboundHandlerConfig) []*protocol.User {
	if i == nil {
		return nil
	}
	inst, err := i.ProxySettings.GetInstance()
	if err != nil || inst == nil {
		fmt.Println("failed to get inbound instance:", err)
		return nil
	}
	switch ty := inst.(type) {
	case *vmessin.Config:
		return ty.User
	case *vlessin.Config:
		return ty.Clients
	case *trojanin.ServerConfig:
		return ty.Users
	default:
		fmt.Println("unsupported inbound type")
	}
	return nil
}

func loadInboundsFromConfigFiles(unnamedArgs []string) []conf.InboundDetourConfig {
	ins := make([]conf.InboundDetourConfig, 0)
	for _, arg := range unnamedArgs {
		r, err := loadArg(arg)
		if err != nil {
			base.Fatalf("failed to load %s: %s", arg, err)
		}
		conf, err := serial.DecodeJSONConfig(r)
		if err != nil {
			base.Fatalf("failed to decode %s: %s", arg, err)
		}
		ins = append(ins, conf.InboundConfigs...)
	}
	if len(ins) == 0 {
		base.Fatalf("no valid inbound found")
	}
	return ins
}
