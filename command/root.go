package command

import (
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

func Execute() {
	_ = rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:                        "v2raybot",
	Aliases:                    nil,
	SuggestFor:                 nil,
	Short:                      "",
	Long:                       "",
	Example:                    "",
	ValidArgs:                  nil,
	ValidArgsFunction:          nil,
	Args:                       nil,
	ArgAliases:                 nil,
	BashCompletionFunction:     "",
	Deprecated:                 "",
	Annotations:                nil,
	Version:                    "",
	PersistentPreRun:           nil,
	PersistentPreRunE:          nil,
	PreRun:                     nil,
	PreRunE:                    nil,
	Run:                        rootRun,
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
	output     string
	nameOutput = "output"

	proxy     string
	nameProxy = "proxy"
)

func init() {
	rootCmd.PersistentFlags().
		StringVarP(&output, nameOutput, "o", "", "output directory")

	rootCmd.PersistentFlags().
		StringVar(&proxy, nameProxy, "", "http proxy address")
}

func rootRun(cmd *cobra.Command, args []string) {
	cmd.Printf("hello\n")
}

func getHttpClient() *http.Client {
	c := &http.Client{}
	if proxy == "" {
		return c
	}

	c.Transport = &http.Transport{
		Proxy: func(r *http.Request) (*url.URL, error) {
			return url.Parse(proxy)
		},
	}
	return c
}
