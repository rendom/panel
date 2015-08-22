// Package main provides ...
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pivotal-golang/bytefmt"

	"./utils"
)

var (
	bwStr     string
	wTitle    string
	cpu       string
	disk      string
	datetime  string
	memory    string
	previdle  int
	prevtotal int

	dzen    bool
	monitor int
)

func formatStr(str string, color string) string {
	if dzen {
		return fmt.Sprintf("^fg(%s)%s^bg()", color, str)
	} else {
		// Lemonbar format
		return fmt.Sprintf("%%{F%s}%s", color, str)
	}
}

func lemonbarOutput() string {
	ws := utils.GetHlwmtags("0")
	left := fmt.Sprintf("%s %s", ws, wTitle)
	right := fmt.Sprintf("%%{r}%s %s %s %s %s", cpu, memory, disk, bwStr, datetime)

	return fmt.Sprintf("%s %s\n", left, right)

}

func dzenOutput() string {
	var scr_width = 1920
	var dpi = 96
	var text_width = 5 * (dpi / 96)

	re := regexp.MustCompile("/\\^[^(]*[^)]*\\)/m")
	ws := utils.GetHlwmtags("0")
	left := fmt.Sprintf("%s %s", ws, wTitle)
	right := fmt.Sprintf("%s %s %s %s %s", cpu, memory, disk, bwStr, datetime)
	rtext := re.ReplaceAllString(right, "")

	pa := scr_width - (len(rtext) * text_width)
	return fmt.Sprintf("%s^pa(%d)%s\n", left, pa, right)
}

func sendEvent(t string, val string) {
	data := fmt.Sprintf("%s\t0\t%s", t, val)
	cmd := exec.Command("herbstclient", "emit_hook", data)
	cmd.Run()
}

func interval(fn func(), t time.Duration) {
	go func() {
		for {
			fn()
			time.Sleep(t)
		}
	}()
}

func main() {
	flag.BoolVar(&dzen, "dzen", false, "Dzen output?")
	flag.IntVar(&monitor, "monitor", 0, "Monitor 0 = Worker, Monitor 1+ = Just listen")
	flag.Parse()

	if monitor == 0 {
		previdle, prevtotal := 0, 0
		// CPU/BW/MEMORY
		var bw utils.Bandwidth
		bw.New("eno1")
		interval(func() {
			c := fmt.Sprintf("Cpu: %d%%", utils.GetCpuLoad(&previdle, &prevtotal))
			m := fmt.Sprintf("Mem: %d%%", utils.GetMemUsage())
			sendEvent("cpu", c)
			sendEvent("bw", fmt.Sprintf("D: %s U: %s", bytefmt.ByteSize(bw.Download), bytefmt.ByteSize(bw.Upload)))
			sendEvent("memory", m)
		}, time.Second*5)

		// time, disk
		interval(func() {
			dspace := utils.GetDiskSpace()
			datetime := utils.GetDatetime()
			sendEvent("datetime", datetime)
			sendEvent("disk", dspace)
		}, time.Second*30)

	}

	cmd := exec.Command("herbstclient", "--idle")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(stdout)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		row := strings.Split(scanner.Text(), "\t")
		switch row[0] {
		case "focus_changed", "window_title_changed":
			wTitle = formatStr(row[2], "#FFFFFF")
		case "datetime":
			datetime = formatStr(row[2], "#FFFFFF")
		case "disk":
			disk = formatStr(row[2], "#FFFFFF")
		case "memory":
			memory = formatStr(row[2], "#FFFFFF")
		case "cpu":
			cpu = formatStr(row[2], "#FFFFFF")
		case "bw":
			bwStr = formatStr(row[2], "#FFFFFF")
		}

		if dzen {
			fmt.Print(dzenOutput())
		} else {
			fmt.Print(lemonbarOutput())
		}
	}
}
