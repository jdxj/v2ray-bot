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
)

var parse = &cobra.Command{
	Use:                        "parse",
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

	filter     string
	nameFilter = "filter"
)

func init() {
	rootCmd.AddCommand(parse)

	parse.Flags().
		StringVar(&fromFile, nameFromFile, "v2ray.share", "parse v2ray share from file")
	parse.Flags().
		StringVar(&fromURL, nameFromURL, "", "parse v2ray share from subscription url")
	parse.MarkFlagsMutuallyExclusive(nameFromFile, nameFromURL)

	parse.Flags().
		StringVar(&filter, nameFilter, "", "filter the specified vmess postscript, prefix matching")
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

func tryParseVmess() ([]*vmess, error) {
	var (
		vmesses []*vmess
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
	if filter == "" {
		return vmesses, nil
	}

	var filtered []*vmess
	for _, v := range vmesses {
		if strings.Contains(v.Ps, filter) {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

func exportVmess(cmd *cobra.Command, vmesses []*vmess) error {
	var writer io.Writer
	if output == "" {
		writer = cmd.OutOrStdout()
	} else {
		f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0644)
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

func parseFromFile(filename string) ([]*vmess, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseFromReader(f)
}

func parseFromURL(url string) ([]*vmess, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return parseFromReader(rsp.Body)
}

func parseFromReader(r io.Reader) ([]*vmess, error) {
	r = base64.NewDecoder(base64.StdEncoding, r)
	scanner := bufio.NewScanner(r)

	var result []*vmess
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

type vmess struct {
	// 配置文件版本号,主要用来识别当前配置
	V string `json:"v"`
	// 备注或别名 postscript
	Ps string `json:"ps"`
	// UUID
	Id string `json:"id"`
	// 地址IP或域名
	Add string `json:"add"`
	// 端口号
	Port uint32 `json:"port"`
	// alterId
	Aid string `json:"aid"`
	// 传输协议(tcp\kcp\ws\h2\quic)
	Net string `json:"net"`
	// 伪装类型(none\http\srtp\utp\wechat-video) *tcp or kcp or QUIC
	Type string `json:"type"`
	// 伪装的域名
	Host string `json:"host"`
	// path
	Path string `json:"path"`
	// 底层传输安全(tls)
	Tls string `json:"tls"`
}

func parseVmess(share string) (*vmess, error) {
	data := strings.TrimPrefix(share, "vmess://")
	jsonData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	v := &vmess{}
	return v, json.Unmarshal(jsonData, v)
}
