package model

import (
	"encoding/json"
	"os"
)

type Vmess struct {
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

func GetVmessFromFile(filename string) ([]*Vmess, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var vmesses []*Vmess
	decoder := json.NewDecoder(f)
	return vmesses, decoder.Decode(&vmesses)
}
