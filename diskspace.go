package main

import (
	"fmt"
	"github.com/golang/glog"
	"os/exec"
	"strings"
)

func GetDiskSpace() (output string) {
	cmd := exec.Command("df")
	out, err := cmd.Output()

	output = ""

	if err != nil {
		glog.Fatalf("df stderr: %s", err)
	}
	lines := strings.Split(string(out), "\n")

	for _, v := range lines {
		if v == "" {
			return
		}

		cols := strings.Fields(v)
		if cols[5] == "/home" || cols[5] == "/" {

		}
	}

	return
}

func main() {
	GetDiskSpace()
}
