// Package main provides ...
package main

import (
	"./utils"
	"fmt"
	"github.com/pivotal-golang/bytefmt"
	"time"
)

func main() {

	// Cpu bw 1s
	// memory 3s
	// time, disc 30s
	// pacman + weather 1h

	var cpu string
	var bws string
	var mu string
	go func() {
		var previdle, prevtotal int
		previdle, prevtotal = 0, 0

		var bw utils.Bandwidth
		bw.New("eno1")

		for {
			cp := utils.GetCpuLoad(&previdle, &prevtotal)
			cpu = fmt.Sprintf("^fg(#ffffff)Cpu: %d%%", cp)
			bws = fmt.Sprintf("^fg(#ffffff)DL: %s UL:%s", bytefmt.ByteSize(bw.Download), bytefmt.ByteSize(bw.Upload))
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			mem := utils.GetMemUsage()
			mu = fmt.Sprintf("^fg(#ffffff)Mem: %d%%", mem)
			time.Sleep(time.Second * 3)
		}
	}()

	for {
		fmt.Println(cpu + mu + bws)
		time.Sleep(time.Millisecond * 300)
	}
}
