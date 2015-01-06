package main

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strconv"
	"strings"
)

func GetMemUsage() int {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		glog.Fatalf("Failed to open meminfo: %s", err)
		return -1
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	memused := 0
	memtotal := 0

	for scanner.Scan() {
		row := scanner.Bytes()
		//c := strings.Split(string(row), "       ")
		c := strings.Fields(string(row))
		switch c[0] {
		case "MemFree:", "Buffers:", "Cached:":
			v, err := strconv.Atoi(c[1])
			if err != nil {
				glog.Fatalf("Failed to convert str to int: %s", err)
			}
			memused = memused + v

		case "MemTotal:":
			v, err := strconv.Atoi(c[1])
			if err != nil {
				glog.Fatalf("Failed to convert str to int: %s", err)
			}
			memtotal = v
		}
	}

	mp := (float64(memused) / float64(memtotal)) * 100
	return int(mp)
}

func main() {
	load := GetMemUsage()
	fmt.Println(load)
}
