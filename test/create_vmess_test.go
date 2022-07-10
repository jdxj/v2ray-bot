package test

import (
	"testing"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	"github.com/v2fly/v2ray-core/v5/app/log"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/outbound"
	"github.com/v2fly/v2ray-core/v5/app/router"
	comLog "github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/http"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	transHttp "github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/jdxj/v2ray-bot/config"
)

func TestCreateV2rayByManual(t *testing.T) {
	cfg := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			&core.InboundHandlerConfig{
				Tag: "http",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: &net.PortRange{
						From: 7891,
						To:   7891,
					},
					Listen: &net.IPOrDomain{
						Address: &net.IPOrDomain_Ip{Ip: []byte{0, 0, 0, 0}},
					},
					AllocationStrategy:         nil,
					StreamSettings:             nil,
					ReceiveOriginalDestination: false,
					DomainOverride:             nil,
					SniffingSettings:           nil,
				}),
				ProxySettings: serial.ToTypedMessage(&http.ServerConfig{
					Timeout:          0,
					Accounts:         nil,
					AllowTransparent: false,
					UserLevel:        0,
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			&core.OutboundHandlerConfig{
				Tag: "proxy",
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
					Via: nil,
					StreamSettings: &internet.StreamConfig{
						Protocol:     internet.TransportProtocol_TCP,
						ProtocolName: "tcp",
						TransportSettings: []*internet.TransportConfig{
							&internet.TransportConfig{
								Protocol:     internet.TransportProtocol_TCP,
								ProtocolName: "tcp",
								Settings: serial.ToTypedMessage(&tcp.Config{
									HeaderSettings: serial.ToTypedMessage(&transHttp.Config{
										Request: &transHttp.RequestConfig{
											Version: &transHttp.Version{Value: "1.1"},
											Method:  &transHttp.Method{Value: "GET"},
											Uri:     []string{"/"},
											Header: []*transHttp.Header{
												&transHttp.Header{
													Name:  "Accept-Encoding",
													Value: []string{"gzip,deflate"},
												},
												&transHttp.Header{
													Name:  "Connection",
													Value: []string{"keep-alive"},
												},
												&transHttp.Header{
													Name:  "Host",
													Value: []string{"xhazsglt5oqpgsyzamn2wojzh.sina.cn"},
												},
												&transHttp.Header{
													Name:  "Pragma",
													Value: []string{"no-cache"},
												},
												&transHttp.Header{
													Name: "User-Agent",
													Value: []string{
														"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
														"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.1.46",
													},
												},
											},
										},
										Response: &transHttp.ResponseConfig{
											Version: nil,
											Status:  nil,
											Header: []*transHttp.Header{
												&transHttp.Header{
													Name: "Content-Type",
													Value: []string{
														"application/octet-stream",
														"video/mpeg",
													},
												},
												&transHttp.Header{
													Name:  "Transfer-Encoding",
													Value: []string{"chunked"},
												},
												&transHttp.Header{
													Name:  "Connection",
													Value: []string{"keep-alive"},
												},
												&transHttp.Header{
													Name:  "Pragma",
													Value: []string{"no-cache"},
												},
												&transHttp.Header{
													Name: "Cache-Control",
													Value: []string{
														"private",
														"no-cache",
													},
												},
											},
										},
									}),
									AcceptProxyProtocol: false,
								}),
							},
						},
						SecurityType:     "",
						SecuritySettings: nil,
						SocketSettings:   nil,
					},
					ProxySettings:     nil,
					MultiplexSettings: nil,
				}),
				ProxySettings: serial.ToTypedMessage(&outbound.Config{Receiver: []*protocol.ServerEndpoint{
					&protocol.ServerEndpoint{
						Address: &net.IPOrDomain{Address: &net.IPOrDomain_Domain{Domain: config.Domain}},
						Port:    config.Port,
						User: []*protocol.User{
							&protocol.User{
								Level: 0,
								Email: "",
								Account: serial.ToTypedMessage(&vmess.Account{
									Id:               config.Id,
									AlterId:          0,
									SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
									TestsEnabled:     "",
								}),
							},
						},
					},
				}}),
				Expire:  0,
				Comment: "",
			},
		},
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{
					Type:  log.LogType_Console,
					Level: comLog.Severity_Warning,
					Path:  "",
				},
				Access: &log.LogSpecification{
					Type:  log.LogType_None,
					Level: comLog.Severity_Unknown,
					Path:  "",
				},
			}),
			serial.ToTypedMessage(&dispatcher.Config{
				Settings: nil,
			}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(&router.Config{
				DomainStrategy: router.DomainStrategy_IpIfNonMatch,
				Rule: []*router.RoutingRule{
					&router.RoutingRule{
						TargetTag:      &router.RoutingRule_Tag{Tag: "proxy"},
						Domain:         nil,
						Cidr:           nil,
						Geoip:          nil,
						PortRange:      nil,
						PortList:       nil,
						NetworkList:    nil,
						Networks:       nil,
						SourceCidr:     nil,
						SourceGeoip:    nil,
						SourcePortList: nil,
						UserEmail:      nil,
						InboundTag:     []string{"http"},
						Protocol:       nil,
						Attributes:     "",
						DomainMatcher:  "mph",
						GeoDomain:      nil,
					},
				},
				BalancingRule: nil,
			}),
		},
		Transport: nil,
		Extension: nil,
	}
	ins, err := core.New(cfg)
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	if err := ins.Start(); err != nil {
		t.Fatalf("%s\n", err)
	}

	time.Sleep(time.Hour)

	if err := ins.Close(); err != nil {
		t.Fatalf("%s\n", err)
	}

}
