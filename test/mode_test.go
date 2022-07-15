package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jdxj/v2ray-bot/config"
	"github.com/jdxj/v2ray-bot/model"
)

func TestNewInternalV2ray(t *testing.T) {
	iv, err := model.NewInternalV2ray(7891, "http", "proxy")
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	err = iv.SetOutbound(config.V)
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	ps := iv.Ping("https://www.google.com")
	fmt.Printf("%+v\n", ps)

	iv.DeleteOutbound()
	iv.Close()
	time.Sleep(time.Hour)
}

func TestNewExternalV2ray(t *testing.T) {
	dokodemo := "127.0.0.1:10085"
	inboundHost := "http://127.0.0.1"
	ev, err := model.NewExternalV2ray(dokodemo, inboundHost, 7891, "proxy")
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	err = ev.SetOutbound(config.V)
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	ps := ev.Ping("https://www.google.com")
	fmt.Printf("%+v\n", ps)

	ev.DeleteOutbound()
	ev.Close()
	time.Sleep(time.Hour)
}
