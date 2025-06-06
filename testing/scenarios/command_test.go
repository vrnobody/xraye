package scenarios

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/xtls/xray-core/app/commander"
	"github.com/xtls/xray-core/app/policy"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/app/proxyman/command"

	"github.com/xtls/xray-core/app/router"
	routercmd "github.com/xtls/xray-core/app/router/command"
	"github.com/xtls/xray-core/app/stats"
	statscmd "github.com/xtls/xray-core/app/stats/command"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/common/uuid"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/proxy/dokodemo"
	"github.com/xtls/xray-core/proxy/freedom"
	"github.com/xtls/xray-core/proxy/vmess"
	"github.com/xtls/xray-core/proxy/vmess/inbound"
	"github.com/xtls/xray-core/proxy/vmess/outbound"
	"github.com/xtls/xray-core/testing/servers/tcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestProxymanGetAddRemoveInboundUsers(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	cmdPort := tcp.PickPort()
	userID := protocol.NewID(uuid.New()).String()
	userEmail := "love@v2ray.com"
	vmessTag := "vmess"

	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&commander.Config{
				Tag: "api",
				Service: []*serial.TypedMessage{
					serial.ToTypedMessage(&command.Config{}),
				},
			}),
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						InboundTag: []string{"api"},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "api",
						},
					},
				},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "api",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(cmdPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
			{
				Tag: vmessTag,
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(tcp.PickPort())}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id: userID,
							}),
							Email: userEmail,
						},
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag:           "default-outbound",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(config)
	common.Must(err)
	defer CloseAllServers(servers)

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cmdPort), grpc.WithInsecure(), grpc.WithBlock())
	common.Must(err)
	defer cmdConn.Close()

	ctx := context.Background()
	hsClient := command.NewHandlerServiceClient(cmdConn)

	// get user
	getResp, err := hsClient.GetInboundUsers(ctx, &command.GetInboundUserRequest{
		Tag:   vmessTag,
		Email: "",
	})
	common.Must(err)
	if getResp == nil || len(getResp.Users) != 1 {
		t.Error("unexpected 1st nil response")
	}
	if getResp.Users[0].Email != userEmail {
		t.Error("unexpected user1 email")
	}
	if inst, err := getResp.Users[0].Account.GetInstance(); err != nil {
		t.Error("unexpected user1 account")
	} else if acc, ok := inst.(*vmess.Account); !ok || acc.Id != userID {
		t.Error("unexpected user1 id")
	}

	// add user
	user2ID := protocol.NewID(uuid.New()).String()
	user2Email := "user2@v2ray.com"
	addResp, err := hsClient.AlterInbound(ctx, &command.AlterInboundRequest{
		Tag: vmessTag,
		Operation: serial.ToTypedMessage(&command.AddUserOperation{
			User: &protocol.User{
				Account: serial.ToTypedMessage(&vmess.Account{
					Id: user2ID,
				}),
				Email: user2Email,
			},
		}),
	})
	common.Must(err)
	if addResp == nil {
		t.Error("unexpected nil response")
	}

	// get user
	getResp, err = hsClient.GetInboundUsers(ctx, &command.GetInboundUserRequest{
		Tag:   vmessTag,
		Email: "",
	})
	common.Must(err)
	if getResp == nil || len(getResp.Users) != 2 {
		t.Error("unexpected 2nd nil response")
	}
	if getResp.Users[1].Email != user2Email {
		t.Error("unexpected user2 email")
	}
	if inst, err := getResp.Users[1].Account.GetInstance(); err != nil {
		t.Error("unexpected user2 account")
	} else if acc, ok := inst.(*vmess.Account); !ok || acc.Id != user2ID {
		t.Error("unexpected user2 ID")
	}

	// remove user
	rmResp, err := hsClient.AlterInbound(ctx, &command.AlterInboundRequest{
		Tag: vmessTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{
			Email: userEmail,
		}),
	})
	common.Must(err)
	if rmResp == nil {
		t.Error("unexpected nil response")
	}

	// get user
	getResp, err = hsClient.GetInboundUsers(ctx, &command.GetInboundUserRequest{
		Tag:   vmessTag,
		Email: "",
	})
	common.Must(err)
	if getResp == nil || len(getResp.Users) != 1 {
		t.Error("unexpected 3rd nil response")
	}
}

func TestRouterGetSetRouting(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	oldConfig := &router.Config{
		Rule: []*router.RoutingRule{
			{
				InboundTag: []string{"api"},
				TargetTag: &router.RoutingRule_Tag{
					Tag: "api",
				},
			},
		},
	}

	cmdPort := tcp.PickPort()
	servConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&commander.Config{
				Tag: "api",
				Service: []*serial.TypedMessage{
					serial.ToTypedMessage(&routercmd.Config{}),
				},
			}),
			serial.ToTypedMessage(oldConfig),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "api",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(cmdPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag:           "default-outbound",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(servConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cmdPort), grpc.WithInsecure(), grpc.WithBlock())
	common.Must(err)
	defer cmdConn.Close()

	client := routercmd.NewRoutingServiceClient(cmdConn)

	ctx := context.Background()
	getRoutingRulesTest(t, client, 1, 0)

	newConfig := &router.Config{
		Rule: []*router.RoutingRule{
			{
				InboundTag: []string{"api"},
				TargetTag: &router.RoutingRule_Tag{
					Tag: "api",
				},
			},
			{
				InboundTag: []string{"test"},
				TargetTag: &router.RoutingRule_Tag{
					Tag: "default-outbound",
				},
			},
		},
		BalancingRule: []*router.BalancingRule{
			{
				Tag:              "pacman",
				OutboundSelector: []string{"agentout"},
				Strategy:         "random",
			},
		},
	}

	setResp, err := client.SetRoutingConfig(ctx, &routercmd.SetRoutingConfigRequest{
		Config: serial.ToTypedMessage(newConfig),
	})
	common.Must(err)
	if setResp == nil {
		t.Error("unexpected nil response")
	}
	getRoutingRulesTest(t, client, 2, 1)

	setResp, err = client.SetRoutingConfig(ctx, &routercmd.SetRoutingConfigRequest{
		Config: serial.ToTypedMessage(oldConfig),
	})
	common.Must(err)
	if setResp == nil {
		t.Error("unexpected nil response")
	}
	getRoutingRulesTest(t, client, 1, 0)
}

func getRoutingRulesTest(t *testing.T, client routercmd.RoutingServiceClient, ruleLen int, balancingRuleLen int) {
	getResp, err := client.GetRoutingConfig(context.Background(), &routercmd.GetRoutingConfigRequest{})
	if err != nil {
		t.Error(err)
	} else if getResp == nil {
		t.Error("unexpected nil response")
	} else if pb, err := getResp.Config.GetInstance(); err != nil {
		t.Error("instance typed message error")
	} else if config, ok := pb.(*router.Config); !ok {
		t.Error("protobuf assertion error")
	} else if len(config.Rule) != ruleLen {
		t.Error("unexpected routing rules length")
	} else if len(config.BalancingRule) != balancingRuleLen {
		t.Error("unexpected balancing roules length")
	}
}

func TestCommanderListenConfigurationItem(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	clientPort := tcp.PickPort()
	cmdPort := tcp.PickPort()
	clientConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&commander.Config{
				Tag:    "api",
				Listen: fmt.Sprintf("127.0.0.1:%d", cmdPort),
				Service: []*serial.TypedMessage{
					serial.ToTypedMessage(&command.Config{}),
				},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "d",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(clientPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag:           "default-outbound",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(clientConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testTCPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Fatal(err)
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cmdPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	common.Must(err)
	defer cmdConn.Close()

	hsClient := command.NewHandlerServiceClient(cmdConn)
	resp, err := hsClient.RemoveInbound(context.Background(), &command.RemoveInboundRequest{
		Tag: "d",
	})
	common.Must(err)
	if resp == nil {
		t.Error("unexpected nil response")
	}

	{
		_, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(clientPort),
		})
		if err == nil {
			t.Error("unexpected nil error")
		}
	}
}

func TestCommanderRemoveHandler(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	clientPort := tcp.PickPort()
	cmdPort := tcp.PickPort()
	clientConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&commander.Config{
				Tag: "api",
				Service: []*serial.TypedMessage{
					serial.ToTypedMessage(&command.Config{}),
				},
			}),
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						InboundTag: []string{"api"},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "api",
						},
					},
				},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "d",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(clientPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
			{
				Tag: "api",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(cmdPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag:           "default-outbound",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(clientConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testTCPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Fatal(err)
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cmdPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	common.Must(err)
	defer cmdConn.Close()

	hsClient := command.NewHandlerServiceClient(cmdConn)
	resp, err := hsClient.RemoveInbound(context.Background(), &command.RemoveInboundRequest{
		Tag: "d",
	})
	common.Must(err)
	if resp == nil {
		t.Error("unexpected nil response")
	}

	{
		_, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(clientPort),
		})
		if err == nil {
			t.Error("unexpected nil error")
		}
	}
}

func TestCommanderListHandlers(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	clientPort := tcp.PickPort()
	cmdPort := tcp.PickPort()
	clientConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&commander.Config{
				Tag: "api",
				Service: []*serial.TypedMessage{
					serial.ToTypedMessage(&command.Config{}),
				},
			}),
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						InboundTag: []string{"api"},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "api",
						},
					},
				},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "d",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(clientPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
			{
				Tag: "api",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(cmdPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag:            "default-outbound",
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{}),
				ProxySettings:  serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(clientConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testTCPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Fatal(err)
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cmdPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	common.Must(err)
	defer cmdConn.Close()

	hsClient := command.NewHandlerServiceClient(cmdConn)
	inboundResp, err := hsClient.ListInbounds(context.Background(), &command.ListInboundsRequest{})
	common.Must(err)
	if inboundResp == nil {
		t.Error("unexpected nil response")
	}

	if !cmp.Equal(inboundResp.Inbounds, clientConfig.Inbound, protocmp.Transform()) {
		t.Fatal("inbound response doesn't match config")
	}

	outboundResp, err := hsClient.ListOutbounds(context.Background(), &command.ListOutboundsRequest{})
	common.Must(err)
	if outboundResp == nil {
		t.Error("unexpected nil response")
	}

	if !cmp.Equal(outboundResp.Outbounds, clientConfig.Outbound, protocmp.Transform()) {
		t.Fatal("outbound response doesn't match config")
	}
}

func TestCommanderAddRemoveUser(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	u1 := protocol.NewID(uuid.New())
	u2 := protocol.NewID(uuid.New())

	cmdPort := tcp.PickPort()
	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&commander.Config{
				Tag: "api",
				Service: []*serial.TypedMessage{
					serial.ToTypedMessage(&command.Config{}),
				},
			}),
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						InboundTag: []string{"api"},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "api",
						},
					},
				},
			}),
			serial.ToTypedMessage(&policy.Config{
				Level: map[uint32]*policy.Policy{
					0: {
						Timeout: &policy.Policy_Timeout{
							UplinkOnly:   &policy.Second{Value: 0},
							DownlinkOnly: &policy.Second{Value: 0},
						},
					},
				},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "v",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(serverPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id: u1.String(),
							}),
						},
					},
				}),
			},
			{
				Tag: "api",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(cmdPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	clientPort := tcp.PickPort()
	clientConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&policy.Config{
				Level: map[uint32]*policy.Policy{
					0: {
						Timeout: &policy.Policy_Timeout{
							UplinkOnly:   &policy.Second{Value: 0},
							DownlinkOnly: &policy.Second{Value: 0},
						},
					},
				},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "d",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(clientPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: u2.String(),
										SecuritySettings: &protocol.SecurityConfig{
											Type: protocol.SecurityType_AES128_GCM,
										},
									}),
								},
							},
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testTCPConn(clientPort, 1024, time.Second*5)(); err != io.EOF &&
		/*We might wish to drain the connection*/
		(err != nil && !strings.HasSuffix(err.Error(), "i/o timeout")) {
		t.Fatal("expected error: ", err)
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cmdPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	common.Must(err)
	defer cmdConn.Close()

	hsClient := command.NewHandlerServiceClient(cmdConn)
	resp, err := hsClient.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: "v",
		Operation: serial.ToTypedMessage(
			&command.AddUserOperation{
				User: &protocol.User{
					Email: "test@example.com",
					Account: serial.ToTypedMessage(&vmess.Account{
						Id: u2.String(),
					}),
				},
			}),
	})
	common.Must(err)
	if resp == nil {
		t.Fatal("nil response")
	}

	if err := testTCPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Fatal(err)
	}

	resp, err = hsClient.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag:       "v",
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{Email: "test@example.com"}),
	})
	common.Must(err)
	if resp == nil {
		t.Fatal("nil response")
	}
}

func TestCommanderStats(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	cmdPort := tcp.PickPort()

	serverConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&commander.Config{
				Tag: "api",
				Service: []*serial.TypedMessage{
					serial.ToTypedMessage(&statscmd.Config{}),
				},
			}),
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						InboundTag: []string{"api"},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "api",
						},
					},
				},
			}),
			serial.ToTypedMessage(&policy.Config{
				Level: map[uint32]*policy.Policy{
					0: {
						Timeout: &policy.Policy_Timeout{
							UplinkOnly:   &policy.Second{Value: 0},
							DownlinkOnly: &policy.Second{Value: 0},
						},
					},
					1: {
						Stats: &policy.Policy_Stats{
							UserUplink:   true,
							UserDownlink: true,
						},
					},
				},
				System: &policy.SystemPolicy{
					Stats: &policy.SystemPolicy_Stats{
						InboundUplink: true,
					},
				},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "vmess",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(serverPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Level: 1,
							Email: "test",
							Account: serial.ToTypedMessage(&vmess.Account{
								Id: userID.String(),
							}),
						},
					},
				}),
			},
			{
				Tag: "api",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(cmdPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	clientPort := tcp.PickPort()
	clientConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortList: &net.PortList{Range: []*net.PortRange{net.SinglePortRange(clientPort)}},
					Listen:   net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(dest.Address),
					Port:     uint32(dest.Port),
					Networks: []net.Network{net.Network_TCP},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: userID.String(),
										SecuritySettings: &protocol.SecurityConfig{
											Type: protocol.SecurityType_AES128_GCM,
										},
									}),
								},
							},
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	if err != nil {
		t.Fatal("Failed to create all servers", err)
	}
	defer CloseAllServers(servers)

	if err := testTCPConn(clientPort, 10240*1024, time.Second*20)(); err != nil {
		t.Fatal(err)
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", cmdPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	common.Must(err)
	defer cmdConn.Close()

	const name = "user>>>test>>>traffic>>>uplink"
	sClient := statscmd.NewStatsServiceClient(cmdConn)

	sresp, err := sClient.GetStats(context.Background(), &statscmd.GetStatsRequest{
		Name:   name,
		Reset_: true,
	})
	common.Must(err)
	if r := cmp.Diff(sresp.Stat, &statscmd.Stat{
		Name:  name,
		Value: 10240 * 1024,
	}, cmpopts.IgnoreUnexported(statscmd.Stat{})); r != "" {
		t.Error(r)
	}

	sresp, err = sClient.GetStats(context.Background(), &statscmd.GetStatsRequest{
		Name: name,
	})
	common.Must(err)
	if r := cmp.Diff(sresp.Stat, &statscmd.Stat{
		Name:  name,
		Value: 0,
	}, cmpopts.IgnoreUnexported(statscmd.Stat{})); r != "" {
		t.Error(r)
	}

	sresp, err = sClient.GetStats(context.Background(), &statscmd.GetStatsRequest{
		Name:   "inbound>>>vmess>>>traffic>>>uplink",
		Reset_: true,
	})
	common.Must(err)
	if sresp.Stat.Value <= 10240*1024 {
		t.Error("value < 10240*1024: ", sresp.Stat.Value)
	}
}
