package command

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jdxj/v2ray-bot/model"
)

var parseCmd = &cobra.Command{
	Use:                        "parse",
	Aliases:                    nil,
	SuggestFor:                 nil,
	Short:                      "parse vmess subscription link",
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
	Run:                        parseRun,
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
	fromFile     string
	nameFromFile = "from-file"

	fromURL     string
	nameFromURL = "from-url"

	filter     []string
	nameFilter = "filter"
)

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().
		StringVar(&fromFile, nameFromFile, "v2ray.share", "parse v2ray share from file")
	parseCmd.Flags().
		StringVar(&fromURL, nameFromURL, "", "parse v2ray share from subscription url")
	parseCmd.MarkFlagsMutuallyExclusive(nameFromFile, nameFromURL)

	parseCmd.Flags().
		StringSliceVar(&filter, nameFilter, nil, "filter the specified vmess postscript, e.g.: k1,k2")
}

func parseRun(cmd *cobra.Command, args []string) {
	vmesses, err := tryParseVmess()
	if err != nil {
		cmd.PrintErrf("parse vmess err: %s", err)
		return
	}

	err = exportVmess(cmd, vmesses)
	if err != nil {
		cmd.PrintErrf("export vmess err: %s", err)
	}
}

func tryParseVmess() ([]*model.Vmess, error) {
	var (
		vmesses []*model.Vmess
		err     error
	)
	if fromURL != "" {
		vmesses, err = parseFromURL(fromURL)
	} else {
		vmesses, err = parseFromFile(fromFile)
	}
	if err != nil {
		return nil, err
	}
	if len(filter) == 0 {
		return vmesses, nil
	}

	var filtered []*model.Vmess
	for _, v := range vmesses {
		for _, keyword := range filter {
			if strings.Contains(v.Ps, keyword) {
				filtered = append(filtered, v)
				break
			}
		}
	}
	return filtered, nil
}

func exportVmess(cmd *cobra.Command, vmesses []*model.Vmess) error {
	var writer io.Writer
	if output == "" {
		writer = cmd.OutOrStdout()
	} else {
		f, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Sync()
			_ = f.Close()
		}()

		writer = f
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(vmesses)
}

func parseFromFile(filename string) ([]*model.Vmess, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseFromReader(f)
}

func parseFromURL(url string) ([]*model.Vmess, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return parseFromReader(rsp.Body)
}

func parseFromReader(r io.Reader) ([]*model.Vmess, error) {
	r = base64.NewDecoder(base64.StdEncoding, r)
	scanner := bufio.NewScanner(r)

	var result []*model.Vmess
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		v, err := parseVmess(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("share: %s, err: %s",
				scanner.Text(), err)
		}
		result = append(result, v)
	}

	return result, scanner.Err()
}

func parseVmess(share string) (*model.Vmess, error) {
	data := strings.TrimPrefix(share, "vmess://")
	jsonData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	v := &model.Vmess{}
	return v, json.Unmarshal(jsonData, v)
}
