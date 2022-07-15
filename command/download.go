package command

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const (
	geoIpURL   = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat"
	geoSiteURL = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat"
)

var downloadCmd = &cobra.Command{
	Use:        "download",
	Aliases:    nil,
	SuggestFor: nil,
	Short:      "download resource",
	Long: fmt.Sprintf(`example:
  download %s
  download %s`, geoIpURL, geoSiteURL),
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
	Run:                        downloadRun,
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
	all     bool
	nameAll = "all"

	proxy     string
	nameProxy = "proxy"

	timeout     string
	nameTimeout = "timeout"
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().
		BoolVar(&all, nameAll, false, "download the resources listed in the example")

	downloadCmd.Flags().
		StringVar(&proxy, nameProxy, "", "http proxy addr")

	downloadCmd.Flags().
		StringVar(&timeout, nameTimeout, "30s", "timeout duration")
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

func downloadRun(cmd *cobra.Command, args []string) {
	if all {
		args = append(args, geoIpURL, geoSiteURL)
	}
	err := cobra.MinimumNArgs(1)(cmd, args)
	if err != nil {
		cmd.PrintErrf("Error: %s\n", err)
		return
	}

	c := getHttpClient(proxy)
	for _, dl := range args {
		if err := downloadFromURL(c, dl); err != nil {
			cmd.PrintErrf("download %s err: %s\n", dl, err)
			continue
		}
	}
}

func downloadFromURL(client *http.Client, dl string) error {
	filePath := filepath.Join(output, path.Base(dl))
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Sync()
		_ = f.Close()
	}()

	rsp, err := client.Get(dl)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	_, err = io.Copy(f, rsp.Body)
	return err
}
