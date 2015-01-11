package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Bandwidth struct {
	Interface string
	Download  uint64
	Upload    uint64

	// Prev
	pDownload uint64
	pUpload   uint64
}

func (b *Bandwidth) update() {
	ul := b.getVal("/sys/class/net/" + b.Interface + "/statistics/tx_bytes")
	dl := b.getVal("/sys/class/net/" + b.Interface + "/statistics/rx_bytes")

	b.Upload = ul - b.pUpload
	b.Download = dl - b.pDownload

	b.pDownload = dl
	b.pUpload = ul
}

func (b *Bandwidth) getVal(file string) uint64 {
	f, err := os.Open(file)
	if err != nil {
		// error
		return 0
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)
	scanner.Scan()

	n, err := strconv.ParseUint(scanner.Text(), 0, 64)
	if err != nil {
		fmt.Println("error int")
		return 0
	}
	return n
}

func (b *Bandwidth) New(i string) {
	b.Interface = i
	go func() {
		for {
			b.update()
			time.Sleep(time.Second)
		}
	}()
}
