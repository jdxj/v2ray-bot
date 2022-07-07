package command

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestPrint(t *testing.T) {
	for {
		t.Logf("abc")

		time.Sleep(time.Second)

		t.Logf("\r")
	}
}

func TestJoinPath(t *testing.T) {
	res := filepath.Join("ii", "abc")
	fmt.Printf(res)
}
