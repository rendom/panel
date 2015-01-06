package main

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strconv"
	"strings"
)

func GetLoad() int {
	f, err := os.Open("/proc/stat")
	if err != nil {
		glog.Fatalf("Initialization failed: %s", err)
		return -1
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	idle, previdle := 0, 0
	total, prevtotal := 0, 0
	for scanner.Scan() {
		row := scanner.Bytes()
		c := strings.Split(string(row), " ")
		if string(c[0]) == "cpu" {
			for i := 2; i < len(c); i++ {
				if c[i] == "" {
					continue
				}
				v, err := strconv.Atoi(c[i])
				if err != nil {
					glog.Fatalf("Failed to convert str to int: %s, err", err)
					return -1
				}

				total = total + v
				if i == 5 {
					idle = v
				}
			}
			break
		}
	}
	diffidle := idle - previdle
	difftotal := total - prevtotal

	diffusage := (1000 * (difftotal - diffidle) / (difftotal + 5)) / 10
	return diffusage
}

func main() {
	load := GetLoad()
	fmt.Println(load)
}
