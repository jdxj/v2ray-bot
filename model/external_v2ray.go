package model

import (
	"context"
	"net/http"

	"github.com/v2fly/v2ray-core/v5/app/proxyman/command"
)

// NewExternalV2ray
// Note: inboundHost 要有 http(s):// 前缀
func NewExternalV2ray(dokodemoDoorAddr, inboundHost string, inboundPort uint32, outboundTag string) (*ExternalV2ray, error) {
	vc, err := getV2rayGrpcClient(dokodemoDoorAddr)
	if err != nil {
		return nil, err
	}

	return &ExternalV2ray{
		vc:          vc,
		hc:          getHttpClient(inboundHost, inboundPort),
		outboundTag: outboundTag,
	}, nil
}

type ExternalV2ray struct {
	vc          command.HandlerServiceClient
	curVmess    *Vmess
	hc          *http.Client
	outboundTag string
}

func (ev *ExternalV2ray) SetOutbound(v *Vmess) error {
	ev.curVmess = v
	_, err := ev.vc.AddOutbound(context.Background(), &command.AddOutboundRequest{
		Outbound: getOutboundHandlerConfig(v, ev.outboundTag),
	})
	return err
}

func (ev *ExternalV2ray) DeleteOutbound() {
	_, _ = ev.vc.RemoveOutbound(context.Background(), &command.RemoveOutboundRequest{
		Tag: ev.outboundTag,
	})
}

func (ev *ExternalV2ray) Ping(host string) *PingStat {
	ps := ping(ev.hc, host)
	ps.V = ev.curVmess
	return ps
}

func (ev *ExternalV2ray) Close() {}
