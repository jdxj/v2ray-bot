package model

type Vmess struct {
}

func (v *Vmess) Encode() (string, error) {

}

func (v *Vmess) Decode(url string) error {

}

func NewVmessSubscription(url string) *VmessSubscription {
	return &VmessSubscription{
		url: url,
	}
}

type VmessSubscription struct {
	url     string
	vmesses []Vmess
}
