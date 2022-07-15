package command

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/jdxj/v2ray-bot/model"
)

var pingCmd = &cobra.Command{
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

	inboundTag     string
	nameInboundTag = "inbound-tag"

	outboundTag     string
	nameOutboundTag = "outbound-tag"

	vmessFile     string
	nameVmessFile = "vmess-file"
)

func init() {
	rootCmd.AddCommand(pingCmd)

	pingCmd.Flags().
		BoolVar(&externalV2ray, nameExternalV2ray, false, "use external v2ray instance to ping")

	pingCmd.Flags().
		StringVar(&inboundHost, nameInboundHost, "http://127.0.0.1", "v2ray http inbound listen addr")

	pingCmd.Flags().
		Uint32Var(&inboundPort, nameInboundPort, 7891, "v2ray http inbound listen port")

	pingCmd.Flags().
		StringVarP(&dokodemoDoorAddr, nameDokodemoDoorAddr, "A", "127.0.0.1:10085", "dokodemo-door listen addr")

	pingCmd.Flags().
		BoolVar(&setFastest, nameSetFastest, false, "set the fastest vmess config")

	pingCmd.Flags().
		StringVar(&inboundTag, nameInboundTag, "http", "use tag to associate inbound and outbound in routing")

	pingCmd.Flags().
		StringVar(&outboundTag, nameOutboundTag, "proxy", "use tag to associate inbound and outbound in routing")

	pingCmd.Flags().
		StringVar(&vmessFile, nameVmessFile, "vmess.txt", "parsed vmess config (parse cmd)")
}

func ping(vc model.V2rayClient, vmesses []*model.Vmess, host string) ([]*model.PingStat, []*model.PingStat, error) {
	var (
		pingStatsNormal []*model.PingStat
		pingStatsErr    []*model.PingStat
	)
	for _, v := range vmesses {
		if err := vc.SetOutbound(v); err != nil {
			return nil, nil, err
		}

		if ps := vc.Ping(host); ps.Err != nil {
			pingStatsErr = append(pingStatsErr, ps)
		} else {
			pingStatsNormal = append(pingStatsNormal, ps)
		}

		vc.DeleteOutbound()
	}

	return pingStatsNormal, pingStatsErr, nil
}

func printPingStat(cmd *cobra.Command, pingStatsNormal, pingStatsErr []*model.PingStat) {
	if len(pingStatsNormal) != 0 {
		sort.Slice(pingStatsNormal, func(i, j int) bool {
			return pingStatsNormal[i].Dur < pingStatsNormal[j].Dur
		})

		cmd.Println("normal:")
		for i, v := range pingStatsNormal {
			cmd.Printf("%3d: %-s %dms\n", i+1, v.V.Ps, v.Dur.Milliseconds())
		}
	}

	if len(pingStatsErr) != 0 {
		cmd.Println("error:")
		for i, v := range pingStatsErr {
			cmd.PrintErrf("%3d: %-s %s\n", i+1, v.V.Ps, v.Err)
		}
	}
}

func pingRun(cmd *cobra.Command, args []string) {
	var (
		vc  model.V2rayClient
		err error
	)
	if externalV2ray {
		vc, err = model.NewExternalV2ray(dokodemoDoorAddr, inboundHost, inboundPort, outboundTag)
	} else {
		vc, err = model.NewInternalV2ray(inboundPort, inboundTag, outboundTag)
	}
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	defer vc.Close()

	vmesses, err := model.GetVmessFromFile(vmessFile)
	if err != nil {
		cmd.PrintErrf("get vmess err: %s", err)
		return
	}

	pingStatsNormal, pingStatsErr, err := ping(vc, vmesses, args[0])
	if err != nil {
		cmd.PrintErrf("ping err: %s", err)
		return
	}
	printPingStat(cmd, pingStatsNormal, pingStatsErr)

	if externalV2ray && setFastest {
		if len(pingStatsNormal) > 0 {
			fastest := pingStatsNormal[0].V
			if err := vc.SetOutbound(fastest); err != nil {
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
