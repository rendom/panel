// Package main provides ...
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
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
	BLACK      string = "#282a2e"
	GREY       string = "#373b41"
	RED        string = "#cc6666"
	GREEN      string = "#b5bd68"
	YELLOW     string = "#de935f"
	BLUE       string = "#81A2BE"
	CYAN       string = "#8abeb7"
	WHITE      string = "#c5c8c6"
	TIMEFORMAT string = "06/01/02 15:04"
)

func colorStr(str string, color string) string {
	if dzen {
		return fmt.Sprintf("^fg(%s)%s^bg()", color, str)
	} else {
		// Lemonbar format
		return fmt.Sprintf("%%{F%s}%s", color, str)
	}
}

func lemonbarOutput() string {
	ws := getHlwmtags(strconv.Itoa(monitor))
	left := fmt.Sprintf("%s %s", ws, wTitle)
	right := fmt.Sprintf("%%{r}%s %s %s %s %s", cpu, memory, disk, bwStr, datetime)

	return fmt.Sprintf("%s %s\n", left, right)

}

func dzenOutput() string {
	var scr_width = 1920
	var dpi = 96
	var text_width = 5 * (dpi / 96)

	re := regexp.MustCompile("/\\^[^(]*[^)]*\\)/m")
	ws := getHlwmtags(strconv.Itoa(monitor))
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
			output = output + " " + colorStr(TAGICON, BLUE)
		case "#":
			output = output + " " + colorStr(TAGICON, BLUE)
		case "+":
			output = output + " " + colorStr(TAGICON, YELLOW)
		case "-":
			output = output + " " + colorStr(TAGICON, GREY)
		case ":":
			output = output + " " + colorStr(TAGICON, GREY)
		case "!":
			output = output + " " + colorStr(TAGICON, GREY)
		case ".":
			output = output + " " + colorStr(TAGICON, BLACK)
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
		return RED
	} else if p >= 0.5 {
		return YELLOW
	} else {
		return GREEN
	}
}

func labelStr(label string, value string) string {
	return colorStr(label, BLUE) + ":" + value
}

func main() {
	flag.BoolVar(&dzen, "dzen", false, "Dzen output?")
	flag.IntVar(&monitor, "monitor", 0, "Monitor 0 = Worker, Monitor 1+ = Just listen")
	flag.Parse()

	if monitor == 0 {

		interval(func() {
			v, err := gmem.VirtualMemory()
			if err == nil {
				sendEvent("memory", labelStr("mem", colorStr(fmt.Sprintf("%.0f%%", v.UsedPercent), getColor(100, v.UsedPercent))))
			}

			c, err := gcpu.Percent(0, false)
			if err == nil {
				sendEvent("cpu", labelStr("cpu", colorStr(fmt.Sprintf("%.0f%%", c[0]), getColor(100, c[0]))))
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

				sendEvent("bw", labelStr("d", colorStr(bytefmt.ByteSize(down), downColor))+labelStr(" u", colorStr(bytefmt.ByteSize(up), upColor)))
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
			wTitle = colorStr(row[2], WHITE)
		case "datetime":
			datetime = colorStr(row[2], WHITE)
		case "disk":
			disk = colorStr(row[2], WHITE)
		case "memory":
			memory = colorStr(row[2], WHITE)
		case "cpu":
			cpu = colorStr(row[2], WHITE)
		case "bw":
			bwStr = colorStr(row[2], WHITE)
		}

		if dzen {
			fmt.Print(dzenOutput())
		} else {
			fmt.Print(lemonbarOutput())
		}
	}
}
