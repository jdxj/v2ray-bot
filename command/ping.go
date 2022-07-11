package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
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
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	coreProxyHttp "github.com/v2fly/v2ray-core/v5/proxy/http"
	coreProxyVmess "github.com/v2fly/v2ray-core/v5/proxy/vmess"
	coreProxyVmessOutbound "github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	transHttp "github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"google.golang.org/protobuf/types/known/anypb"
)

var ping = &cobra.Command{
	Use:        "ping",
	Aliases:    nil,
	SuggestFor: nil,
	Short:      "http ping",
	Long: `example:
  ping https://www.google.com --vmess-file vmess.txt`,
	Example:                    "",
	ValidArgs:                  nil,
	ValidArgsFunction:          nil,
	Args:                       cobra.MinimumNArgs(1),
	ArgAliases:                 nil,
	BashCompletionFunction:     "",
	Deprecated:                 "",
	Annotations:                nil,
	Version:                    "",
	PersistentPreRun:           nil,
	PersistentPreRunE:          nil,
	PreRun:                     nil,
	PreRunE:                    nil,
	Run:                        pingRun,
	RunE:                       nil,
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	CompletionOptions:          cobra.CompletionOptions{},
	TraverseChildren:           false,
	Hidden:                     false,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
}

var (
	inboundPort     uint32
	nameInboundPort = "inbound-port"

	vmessFile     string
	nameVmessFile = "vmess-file"
)

func init() {
	rootCmd.AddCommand(ping)

	ping.Flags().
		Uint32Var(&inboundPort, nameInboundPort, 7891, "port for v2ray http inbound")

	ping.Flags().
		StringVar(&vmessFile, nameVmessFile, "vmess.txt", "parsed vmess config (parse cmd)")
}

func getHttpClient(httpProxy string) *http.Client {
	dur, _ := time.ParseDuration(timeout)
	c := &http.Client{
		Timeout: dur,
	}
	if httpProxy == "" {
		return c
	}

	c.Transport = &http.Transport{
		Proxy: func(r *http.Request) (*url.URL, error) {
			return url.Parse(httpProxy)
		},
	}
	return c
}

func getVmessFromFile() ([]*vmess, error) {
	f, err := os.Open(vmessFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var vmesses []*vmess
	decoder := json.NewDecoder(f)
	return vmesses, decoder.Decode(&vmesses)
}

const (
	routingTag = "proxy"
)

func getV2rayConfig(inboundPort uint32) *core.Config {
	return &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "http",
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
						TargetTag:     &router.RoutingRule_Tag{Tag: routingTag},
						InboundTag:    []string{"http"},
						DomainMatcher: "mph",
					},
				},
			}),
		},
	}
}

func startV2ray() (*core.Instance, error) {
	ins, err := core.New(getV2rayConfig(inboundPort))
	if err != nil {
		return nil, fmt.Errorf("new v2ray core err: %s", err)
	}

	if err := ins.Start(); err != nil {
		return nil, fmt.Errorf("start v2ray err: %s", err)
	}
	return ins, nil
}

func addOutboundHandler(ins *core.Instance, vmess *vmess) error {
	outboundConfig := &core.OutboundHandlerConfig{
		Tag: routingTag,
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
									Uri:     []string{vmess.Path},
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
											Value: []string{vmess.Host},
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
				Address: &net.IPOrDomain{Address: &net.IPOrDomain_Domain{Domain: vmess.Add}},
				Port:    vmess.Port,
				User: []*protocol.User{
					{
						Account: serial.ToTypedMessage(&coreProxyVmess.Account{
							Id:               vmess.Id,
							SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
						}),
					},
				},
			},
		}}),
	}

	err := core.AddOutboundHandler(ins, outboundConfig)
	if err != nil {
		return fmt.Errorf("add outbound handler err: %s", err)
	}
	return nil
}

func removeOutboundHandler(ins *core.Instance) error {
	outboundManager := ins.GetFeature(outbound.ManagerType()).(outbound.Manager)
	err := outboundManager.RemoveHandler(context.Background(), routingTag)
	if err != nil {
		return fmt.Errorf("remove handler %s err: %s", routingTag, err)
	}
	return nil
}

func tryPing(c *http.Client, host string) (time.Duration, error) {
	start := time.Now()
	rsp, err := c.Get(host)
	dur := time.Since(start)

	if err != nil {
		return dur, err
	}

	_, _ = io.Copy(io.Discard, rsp.Body)
	_ = rsp.Body.Close()
	return dur, nil
}

type pingStat struct {
	v   *vmess
	dur time.Duration
}

func pingRun(cmd *cobra.Command, args []string) {
	host := args[0]
	c := getHttpClient(fmt.Sprintf("http://127.0.0.1:%d", inboundPort))
	vmesses, err := getVmessFromFile()
	if err != nil {
		cmd.PrintErrf("get vmess err: %s", err)
		return
	}

	ins, err := startV2ray()
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	defer ins.Close()

	var pingStats []pingStat
	for _, v := range vmesses {
		err := addOutboundHandler(ins, v)
		if err != nil {
			cmd.PrintErrln(err)
			return
		}

		dur, err := tryPing(c, host)
		if err != nil {
			cmd.PrintErrf("ping err: %s", err)
		}

		pingStats = append(pingStats, pingStat{
			v:   v,
			dur: dur,
		})

		err = removeOutboundHandler(ins)
		if err != nil {
			cmd.PrintErrln(err)
			return
		}
	}

	sort.Slice(pingStats, func(i, j int) bool {
		return pingStats[i].dur < pingStats[j].dur
	})

	for i, v := range pingStats {
		cmd.Printf("%3d. %-s %4dms\n", i+1, v.v.Ps, v.dur.Milliseconds())
	}
}
