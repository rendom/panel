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

	"github.com/golang/glog"
	"github.com/pivotal-golang/bytefmt"
	gcpu "github.com/shirou/gopsutil/cpu"
	gmem "github.com/shirou/gopsutil/mem"
	gnet "github.com/shirou/gopsutil/net"
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

const (
	TAGICON    string = "x"
	COLOR1     string = "#000000"
	COLOR2     string = "#212121"
	COLOR3     string = "#802828"
	COLOR4     string = "#9ca554"
	COLOR5     string = "#ddb62b"
	COLOR6     string = "#1e6a9a"
	TIMEFORMAT string = "06/01/02 15:04"
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
	ws := getHlwmtags("0")
	left := fmt.Sprintf("%s %s", ws, wTitle)
	right := fmt.Sprintf("%%{r}%s %s %s %s %s", cpu, memory, disk, bwStr, datetime)

	return fmt.Sprintf("%s %s\n", left, right)

}

func dzenOutput() string {
	var scr_width = 1920
	var dpi = 96
	var text_width = 5 * (dpi / 96)

	re := regexp.MustCompile("/\\^[^(]*[^)]*\\)/m")
	ws := getHlwmtags("0")
	left := fmt.Sprintf("%s %s", ws, wTitle)
	right := fmt.Sprintf("%s %s %s %s %s", cpu, memory, disk, bwStr, datetime)
	rtext := re.ReplaceAllString(right, "")

	pa := scr_width - (len(rtext) * text_width)
	return fmt.Sprintf("%s^pa(%d)%s\n", left, pa, right)
}

func getHlwmtags(monitor string) (output string) {
	out, err := exec.Command("herbstclient", "tag_status", monitor).Output()
	if err != nil {
		glog.Fatalf("hlc stderr %s", err)
	}

	tags := strings.Split(string(out), "\t")

	for _, v := range tags {
		if v == "" {
			continue
		}
		switch v[:1] {
		case "%":
			output = output + formatStr(TAGICON, COLOR6)
		case "#":
			output = output + formatStr(TAGICON, COLOR5)
		case "+":
			output = output + formatStr(TAGICON, COLOR5)
		case "-":
			output = output + formatStr(TAGICON, COLOR6)
		case ":":
			output = output + formatStr(TAGICON, COLOR3)
		case "!":
			output = output + formatStr(TAGICON, COLOR2)
		case ".":
			output = output + formatStr(TAGICON, COLOR5)
		}
	}

	return
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

func getColor(max float64, val float64) string {
	if max == 0 {
		return ""
	}

	p := val / max
	if p >= 0.8 {
		return "#802828"
	} else if p >= 0.5 {
		return "#ddb62b"
	} else {
		return "#9ca554"
	}
}

func main() {
	flag.BoolVar(&dzen, "dzen", false, "Dzen output?")
	flag.IntVar(&monitor, "monitor", 0, "Monitor 0 = Worker, Monitor 1+ = Just listen")
	flag.Parse()

	if monitor == 0 {

		interval(func() {
			v, err := gmem.VirtualMemory()
			if err == nil {
				sendEvent("memory", formatStr(fmt.Sprintf("mem: %.0f%%", v.UsedPercent), getColor(100, v.UsedPercent)))
			}

			c, err := gcpu.Percent(0, false)
			if err == nil {
				sendEvent("cpu", formatStr(fmt.Sprintf("cpu: %.0f%%", c[0]), getColor(100, c[0])))
			}
		}, time.Second*5)

		var oldUp uint64
		var oldDown uint64

		interval(func() {
			ios, err := gnet.IOCounters(false)
			if err == nil {
				down := ios[0].BytesRecv - oldDown
				oldDown = ios[0].BytesRecv
				downColor := getColor(1000000.0, float64(down))

				up := ios[0].BytesSent - oldUp
				oldUp = ios[0].BytesSent
				upColor := getColor(1000000.0, float64(up))

				sendEvent("bw", formatStr("D:"+bytefmt.ByteSize(down), downColor)+formatStr(" U:"+bytefmt.ByteSize(up), upColor))
			}
		}, time.Second*1)
		// time, disk
		interval(func() {
			t := time.Now().Local()
			datetime := t.Format(TIMEFORMAT)

			sendEvent("datetime", datetime)
			// sendEvent("disk", dspace)
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
		if len(row) < 3 {
			continue
		}
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
