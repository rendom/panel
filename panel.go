// Package main provides ...
package main

import (
	"fmt"
	"regexp"
	"time"

	"./utils"
	"github.com/pivotal-golang/bytefmt"
)

func main() {
	// Cpu bw 1s
	// memory 3s
	// time, disc 30s
	// pacman + weather 1h

	var cpu string
	var bws string
	var mu string
	var ws string
	var wtitle string
	var dspace string
	var datetime string

	var scr_width = 1920
	var dpi = 96
	var text_width = 5 * (dpi / 96)

	go func() {
		var previdle, prevtotal int
		previdle, prevtotal = 0, 0

		var bw utils.Bandwidth
		bw.New("eno1")

		for {
			cp := utils.GetCpuLoad(&previdle, &prevtotal)
			cpu = fmt.Sprintf("^fg(#ffffff)Cpu: %d%%", cp)
			bws = fmt.Sprintf("^fg(#ffffff)DL: %s UL:%s", bytefmt.ByteSize(bw.Download), bytefmt.ByteSize(bw.Upload))
			time.Sleep(time.Second * 1)
		}
	}()

	go func() {
		for {
			mem := utils.GetMemUsage()
			mu = fmt.Sprintf("^fg(#ffffff)Mem: %d%%", mem)
			time.Sleep(time.Second * 3)
		}
	}()

	go func() {
		for {
			dspace = utils.GetDiskSpace()
			datetime = utils.GetDatetime()
			time.Sleep(time.Second * 30)
		}
	}()
	re := regexp.MustCompile("/\\^[^(]*[^)]*\\)/m")
	for {
		ws = utils.GetHlwmtags("0")
		wtitle = fmt.Sprintf("^fg(#ffffff)%s^bg()", utils.GetActiveWindowName())

		left := fmt.Sprintf("%s %s", ws, wtitle)
		right := fmt.Sprintf("%s %s %s %s %s", cpu, mu, dspace, bws, datetime)
		rtext := re.ReplaceAllString(right, "")

		pa := scr_width - (len(rtext) * text_width)
		fmt.Printf("%s^pa(%d)%s\n", left, pa, right)

		// Sleep 300ms
		time.Sleep(time.Millisecond * 300)
	}
}
