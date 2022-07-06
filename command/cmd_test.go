package command

import (
	"fmt"
	"os"
	"testing"
)

func TestParseFromReader(t *testing.T) {
	f, err := os.Open("../config/v2ray.share")
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	defer f.Close()

	result, err := parseFromReader(f)
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	for _, v := range result {
		fmt.Printf("%+v\n", v)
	}
}
