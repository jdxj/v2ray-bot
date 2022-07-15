package model

import (
	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	"github.com/v2fly/v2ray-core/v5/app/log"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/outbound"
	"github.com/v2fly/v2ray-core/v5/app/router"
	comLog "github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	coreProxyHttp "github.com/v2fly/v2ray-core/v5/proxy/http"
	coreProxyVmess "github.com/v2fly/v2ray-core/v5/proxy/vmess"
	coreProxyVmessOutbound "github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	transHttp "github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"
)

// getV2rayPingConfig 获取用于延迟测试的配置,
// 预配置 Inbound, router. outbound 在测试时动态生成.
func getV2rayPingConfig(inboundPort uint32, inboundTag, outboundTag string) *core.Config {
	return &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: inboundTag,
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: &net.PortRange{
						From: inboundPort,
						To:   inboundPort,
					},
					Listen: &net.IPOrDomain{
						Address: &net.IPOrDomain_Ip{Ip: []byte{0, 0, 0, 0}},
					},
				}),
				ProxySettings: serial.ToTypedMessage(&coreProxyHttp.ServerConfig{}),
			},
		},
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{
					Type:  log.LogType_Console,
					Level: comLog.Severity_Warning,
				},
				Access: &log.LogSpecification{
					Type:  log.LogType_None,
					Level: comLog.Severity_Unknown,
				},
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(&router.Config{
				DomainStrategy: router.DomainStrategy_IpIfNonMatch,
				Rule: []*router.RoutingRule{
					{
						TargetTag:     &router.RoutingRule_Tag{Tag: outboundTag},
						InboundTag:    []string{inboundTag},
						DomainMatcher: "mph",
					},
				},
			}),
		},
	}
}

// getOutboundHandlerConfig 根据 vmess 生成 outbound 配置.
func getOutboundHandlerConfig(v *Vmess, outboundTag string) *core.OutboundHandlerConfig {
	return &core.OutboundHandlerConfig{
		Tag: outboundTag,
		SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
			StreamSettings: &internet.StreamConfig{
				Protocol:     internet.TransportProtocol_TCP,
				ProtocolName: "tcp",
				TransportSettings: []*internet.TransportConfig{
					{
						Protocol:     internet.TransportProtocol_TCP,
						ProtocolName: "tcp",
						Settings: serial.ToTypedMessage(&tcp.Config{
							HeaderSettings: serial.ToTypedMessage(&transHttp.Config{
								Request: &transHttp.RequestConfig{
									Version: &transHttp.Version{Value: "1.1"},
									Method:  &transHttp.Method{Value: "GET"},
									Uri:     []string{v.Path},
									Header: []*transHttp.Header{
										{
											Name:  "Accept-Encoding",
											Value: []string{"gzip,deflate"},
										},
										{
											Name:  "Connection",
											Value: []string{"keep-alive"},
										},
										{
											Name:  "Host",
											Value: []string{v.Host},
										},
										{
											Name:  "Pragma",
											Value: []string{"no-cache"},
										},
										{
											Name: "User-Agent",
											Value: []string{
												"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
												"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.1.46",
											},
										},
									},
								},
								Response: &transHttp.ResponseConfig{
									Header: []*transHttp.Header{
										{
											Name: "Content-Type",
											Value: []string{
												"application/octet-stream",
												"video/mpeg",
											},
										},
										{
											Name:  "Transfer-Encoding",
											Value: []string{"chunked"},
										},
										{
											Name:  "Connection",
											Value: []string{"keep-alive"},
										},
										{
											Name:  "Pragma",
											Value: []string{"no-cache"},
										},
										{
											Name: "Cache-Control",
											Value: []string{
												"private",
												"no-cache",
											},
										},
									},
								},
							}),
						}),
					},
				},
			},
		}),
		ProxySettings: serial.ToTypedMessage(&coreProxyVmessOutbound.Config{Receiver: []*protocol.ServerEndpoint{
			{
				Address: &net.IPOrDomain{Address: &net.IPOrDomain_Domain{Domain: v.Add}},
				Port:    v.Port,
				User: []*protocol.User{
					{
						Account: serial.ToTypedMessage(&coreProxyVmess.Account{
							Id:               v.Id,
							SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
						}),
					},
				},
			},
		}}),
	}
}

func getV2rayGrpcClient(dokodemoDoorAddr string) (command.HandlerServiceClient, error) {
	conn, err := grpc.Dial(dokodemoDoorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return command.NewHandlerServiceClient(conn), nil
}
