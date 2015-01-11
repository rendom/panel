package utils

import (
	"bufio"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func getActiveWindowName() string {
	out, err := exec.Command("bash", "-c", "xprop -id $(xprop -root _NET_ACTIVE_WINDOW | cut -d' ' -f5) _NET_WM_NAME | grep -oP '(?<=\")(.+)(?=\")'").Output()
	if err != nil {
		glog.Fatalf("stderr: %s", err)
		return "No title"
	}
	return string(out)
}

func GetHlwmtags(monitor string) {
	//	out, err := exec.Command("herbstclient", "tag_status", monitor).Output()
	//	if err != nil {
	//		glog.Fatalf("hlc stderr %s", err)
	//	}

	//	tags := strings.Split(string(out), "\t")
}

func getPacmanUpdatesCount() int {
	out, err := exec.Command("bash", "-c", "checkupdates | wc -l").Output()
	if err != nil {
		glog.Fatalf("checkupdates stderr:%s", err)
		return -1
	}
	count, err := strconv.Atoi(string(out[:len(out)-1]))
	return count
}

func GetDiskSpace() (output string) {
	cmd := exec.Command("df", "-h")
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
			output = output + " " + cols[5] + " " + cols[4] + " " + cols[3]
		}
	}

	return
}

func GetCpuLoad(previdle *int, prevtotal *int) int {
	f, err := os.Open("/proc/stat")
	if err != nil {
		glog.Fatalf("Initialization failed: %s", err)
		return -1
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	idle, total := 0, 0
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

	diffidle := idle - *previdle
	difftotal := total - *prevtotal
	*previdle = idle
	*prevtotal = total

	diffusage := (1000 * (difftotal - diffidle) / (difftotal + 5)) / 10
	return diffusage
}

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
