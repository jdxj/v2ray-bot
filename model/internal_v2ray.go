package model

import (
	"context"
	"net/http"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
)

func NewInternalV2ray(inboundPort uint32, inboundTag, outboundTag string) (*InternalV2ray, error) {
	ins, err := core.New(getV2rayPingConfig(inboundPort, inboundTag, outboundTag))
	if err != nil {
		return nil, err
	}
	err = ins.Start()
	if err != nil {
		return nil, err
	}

	c := &InternalV2ray{
		ins:         ins,
		hc:          getHttpClient("http://127.0.0.1", inboundPort),
		outboundTag: outboundTag,
	}
	return c, nil
}

type InternalV2ray struct {
	ins         *core.Instance
	curVmess    *Vmess
	hc          *http.Client
	outboundTag string
}

func (iv *InternalV2ray) SetOutbound(v *Vmess) error {
	iv.curVmess = v
	return core.AddOutboundHandler(iv.ins, getOutboundHandlerConfig(v, iv.outboundTag))
}

func (iv *InternalV2ray) DeleteOutbound() {
	outboundManager := iv.ins.GetFeature(outbound.ManagerType()).(outbound.Manager)
	_ = outboundManager.RemoveHandler(context.Background(), iv.outboundTag)
}

func (iv *InternalV2ray) Ping(host string) *PingStat {
	ps := ping(iv.hc, host)
	ps.V = iv.curVmess
	return ps
}

func (iv *InternalV2ray) Close() {
	_ = iv.ins.Close()
}
