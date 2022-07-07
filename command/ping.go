package command

import (
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var ping = &cobra.Command{
	Use:        "ping",
	Aliases:    nil,
	SuggestFor: nil,
	Short:      "http ping",
	Long: `example:
  ping https://www.google.com --proxy http://127.0.0.1:7891`,
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
	count     uint64
	nameCount = "count"

	interval     uint64
	nameInterval = "interval"
)

func init() {
	rootCmd.AddCommand(ping)

	ping.Flags().
		Uint64Var(&count, nameCount, 1, "ping count")

	ping.Flags().
		Uint64Var(&interval, nameInterval, 500, "ping interval(ms)")
}

func pingRun(cmd *cobra.Command, args []string) {
	c := getHttpClient()
	for i := uint64(0); i < count; i++ {
		tryPing(cmd, c, args[0])
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

func tryPing(cmd *cobra.Command, c *http.Client, host string) {
	var (
		rsp *http.Response
		err error
	)
	dur := delay(func() {
		rsp, err = c.Get(host)
	})
	if err != nil {
		cmd.PrintErrf("head %s err: %s, duration: %.3fs", host, err, dur.Seconds())
		return
	} else {
		cmd.Printf("duration: %.3fs\n", dur.Seconds())
	}

	_, _ = io.Copy(io.Discard, rsp.Body)
	_ = rsp.Body.Close()
}

func delay(f func()) time.Duration {
	if f == nil {
		return 0
	}

	start := time.Now()
	f()
	return time.Since(start)
}
