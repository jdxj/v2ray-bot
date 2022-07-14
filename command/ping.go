package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	"github.com/v2fly/v2ray-core/v5/app/log"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/app/router"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"

	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/outbound"

	core "github.com/v2fly/v2ray-core/v5"
	comLog "github.com/v2fly/v2ray-core/v5/common/log"
	coreProxyHttp "github.com/v2fly/v2ray-core/v5/proxy/http"
	coreProxyVmess "github.com/v2fly/v2ray-core/v5/proxy/vmess"
	coreProxyVmessOutbound "github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
	transHttp "github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
)

var (
	ErrGetVmess = errors.New("get vmess err")
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
	externalV2ray     bool
	nameExternalV2ray = "external-v2ray"

	inboundHost     string
	nameInboundHost = "inbound-host"

	inboundPort     uint32
	nameInboundPort = "inbound-port"

	dokodemoDoorAddr     string
	nameDokodemoDoorAddr = "dokodemo-door-addr"

	setFastest     bool
	nameSetFastest = "set-fastest"

	outboundTag     string
	nameOutboundTag = "outbound-tag"

	vmessFile     string
	nameVmessFile = "vmess-file"
)

func init() {
	rootCmd.AddCommand(ping)

	ping.Flags().
		BoolVar(&externalV2ray, nameExternalV2ray, false, "use external v2ray instance to ping")

	ping.Flags().
		StringVar(&inboundHost, nameInboundHost, "http://127.0.0.1", "v2ray http inbound listen addr")

	ping.Flags().
		Uint32Var(&inboundPort, nameInboundPort, 7891, "v2ray http inbound listen port")

	ping.Flags().
		StringVarP(&dokodemoDoorAddr, nameDokodemoDoorAddr, "A", "127.0.0.1:10085", "dokodemo-door listen addr")

	ping.Flags().
		BoolVar(&setFastest, nameSetFastest, false, "set the fastest vmess config")

	ping.Flags().
		StringVar(&outboundTag, nameOutboundTag, "proxy", "use tag to associate inbound and outbound in routing")

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
						TargetTag:     &router.RoutingRule_Tag{Tag: outboundTag},
						InboundTag:    []string{"http"},
						DomainMatcher: "mph",
					},
				},
			}),
		},
	}
}

func getOutboundHandlerConfig(vmess *vmess) *core.OutboundHandlerConfig {
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
}

func addOutboundHandler(ins *core.Instance, vmess *vmess) error {
	err := core.AddOutboundHandler(ins, getOutboundHandlerConfig(vmess))
	if err != nil {
		return fmt.Errorf("add outbound handler err: %s", err)
	}
	return nil
}

func removeOutboundHandler(ins *core.Instance) {
	outboundManager := ins.GetFeature(outbound.ManagerType()).(outbound.Manager)
	_ = outboundManager.RemoveHandler(context.Background(), outboundTag)
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

func getV2ray() (command.HandlerServiceClient, error) {
	conn, err := grpc.Dial(dokodemoDoorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return command.NewHandlerServiceClient(conn), nil
}

func doPing(c *http.Client, host string) (time.Duration, error) {
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
	err error
}

func pingByInternalV2ray(pingHost string) ([]pingStat, []pingStat, error) {
	c := getHttpClient(fmt.Sprintf("http://127.0.0.1:%d", inboundPort))
	vmesses, err := getVmessFromFile()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrGetVmess, err)
	}

	ins, err := startV2ray()
	if err != nil {
		return nil, nil, err
	}
	defer ins.Close()

	var (
		pingStatsNormal []pingStat
		pingStatsErr    []pingStat
	)
	for _, v := range vmesses {
		err := addOutboundHandler(ins, v)
		if err != nil {
			return nil, nil, err
		}

		dur, err := doPing(c, pingHost)
		stat := pingStat{
			v:   v,
			dur: dur,
		}
		if err != nil {
			stat.err = fmt.Errorf("ping err: %s, ps: %s", err, v.Ps)
			pingStatsErr = append(pingStatsErr, stat)
		} else {
			pingStatsNormal = append(pingStatsNormal, stat)
		}

		removeOutboundHandler(ins)
	}

	return pingStatsNormal, pingStatsErr, nil
}

func pingByExternalV2ray(client command.HandlerServiceClient, pingHost string) ([]pingStat, []pingStat, error) {
	c := getHttpClient(fmt.Sprintf("%s:%d", inboundHost, inboundPort))
	vmesses, err := getVmessFromFile()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrGetVmess, err)
	}

	var (
		pingStatsNormal []pingStat
		pingStatsErr    []pingStat
		ctx             = context.Background()
	)
	for _, v := range vmesses {
		_, err = client.AddOutbound(ctx, &command.AddOutboundRequest{Outbound: getOutboundHandlerConfig(v)})
		if err != nil {
			return nil, nil, err
		}

		dur, err := doPing(c, pingHost)
		stat := pingStat{
			v:   v,
			dur: dur,
		}
		if err != nil {
			stat.err = fmt.Errorf("ping err: %s, ps: %s", err, v.Ps)
			pingStatsErr = append(pingStatsErr, stat)
		} else {
			pingStatsNormal = append(pingStatsNormal, stat)
		}

		_, _ = client.RemoveOutbound(context.Background(), &command.RemoveOutboundRequest{Tag: outboundTag})
	}
	return pingStatsNormal, pingStatsErr, nil
}

func setFastestVmess(client command.HandlerServiceClient, v *vmess) error {
	_, err := client.AddOutbound(context.Background(), &command.AddOutboundRequest{Outbound: getOutboundHandlerConfig(v)})
	if err != nil {
		return fmt.Errorf("set fastest err: %s", err)
	}
	return nil
}

func pingRun(cmd *cobra.Command, args []string) {
	var (
		pingHost        = args[0]
		pingStatsNormal []pingStat
		pingStatsErr    []pingStat
		err             error

		v2rayClient command.HandlerServiceClient
	)
	if !externalV2ray {
		pingStatsNormal, pingStatsErr, err = pingByInternalV2ray(pingHost)
	} else {
		v2rayClient, err = getV2ray()
		if err != nil {
			cmd.PrintErrf("get v2ray err: %s", err)
			return
		}
		pingStatsNormal, pingStatsErr, err = pingByExternalV2ray(v2rayClient, pingHost)
	}

	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	sort.Slice(pingStatsNormal, func(i, j int) bool {
		return pingStatsNormal[i].dur < pingStatsNormal[j].dur
	})

	if len(pingStatsNormal) != 0 {
		cmd.Println("normal:")
		for i, v := range pingStatsNormal {
			cmd.Printf("%d: %-s %dms\n", i+1, v.v.Ps, v.dur.Milliseconds())
		}
	}

	if len(pingStatsErr) != 0 {
		cmd.Println("error:")
		for i, v := range pingStatsErr {
			cmd.PrintErrf("%d: %-s %s\n", i+1, v.v.Ps, v.err)
		}
	}

	if externalV2ray && setFastest {
		if len(pingStatsNormal) > 0 {
			fastest := pingStatsNormal[0].v
			err := setFastestVmess(v2rayClient, fastest)
			if err != nil {
				cmd.PrintErrln(err)
			} else {
				cmd.Printf("set fastest %s success", fastest.Ps)
			}
		} else {
			cmd.PrintErrln("no available vmess")
		}
	} else if !externalV2ray && setFastest {
		cmd.PrintErrln("can not set internal v2ray")
	}
}
