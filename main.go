package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Progress struct {
	Length, Downloaded int64
	isDone             bool
}

func (p Progress) String() string {
	if p.Length < 0 {
		return fmt.Sprintf("Downloaded %s", PrettifySize(p.Downloaded))
	}

	var percentage int = int(p.Downloaded * 10 / p.Length)
	perStr := []string{strings.Repeat("=", percentage), strings.Repeat("-", 10-percentage)}

	return fmt.Sprintf("%s Downloaded %s of %s", perStr, PrettifySize(p.Downloaded), PrettifySize(p.Length))
}

func (p *Progress) Write(buf []byte) (n int, err error) {
	p.Downloaded += int64(len(buf))
	return len(buf), nil
}

func PrettifySize(size int64) string {
	pow := 0
	for i := 0; i < 5; i++ {
		if size > 1024 {
			size /= 1024
			pow++
		}
	}

	switch pow {
	case 0:
		return fmt.Sprintf("%.2fB", float64(size))
	case 1:
		return fmt.Sprintf("%.2fK", float64(size))
	case 2:
		return fmt.Sprintf("%.2fM", float64(size))
	case 3:
		return fmt.Sprintf("%.2fG", float64(size))
	case 4:
		return fmt.Sprintf("%.2fT", float64(size))
	default:
		return fmt.Sprintf("%.2fB", float64(size))
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "Error: no URL specified\n")
		return
	}

	URL := os.Args[1]
	resp, err := http.Get(URL)
	defer resp.Body.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s \n", err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: status: %s \n", resp.Status)
		return
	}

	progress := Progress{
		Length:     resp.ContentLength,
		Downloaded: 0,
	}

	fileName := URL[strings.LastIndex(URL, "/")+1 : len(URL)]
	file, _ := os.Create(fileName)
	defer file.Close()

	if resp.ContentLength > 0 {
		if err := file.Truncate(resp.ContentLength); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s \n", err.Error())
			return
		}
		fmt.Println("Disk space allocated")
	}

	go func() {
		for {
			time.Sleep(time.Second)
			if progress.isDone {
				return
			}
			fmt.Printf("\rOn %s", progress)
		}
	}()

	tee := io.TeeReader(resp.Body, &progress)
	io.Copy(file, tee)

	progress.isDone = true
	fmt.Println()
	fmt.Println("Finished")
}
