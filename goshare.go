package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

var downloads chan FileDownload = make(chan FileDownload, 100)
var download_id chan string = make(chan string)

type FileDownload struct {
	id         string
	name       string
	size       float64
	downloaded float64
	progress   float64
}

func (fd *FileDownload) Progress(bytes int) {
	fd.downloaded += float64(bytes)
	fd.progress = (fd.downloaded / fd.size) * 100
}

type FileMonitor struct {
	f        http.File
	download FileDownload
}

func (fm *FileMonitor) Close() error {
	return fm.f.Close()
}

func (fm FileMonitor) Stat() (os.FileInfo, error) {
	return fm.f.Stat()
}

func (fm FileMonitor) Readdir(count int) ([]os.FileInfo, error) {
	return fm.f.Readdir(count)
}

func (fm *FileMonitor) Read(b []byte) (int, error) {
	bytes, err := fm.f.Read(b)
	if fm.download.size == 0.00 {
		fileInfo, _ := fm.f.Stat()
		fm.download.size = float64(fileInfo.Size())
		fm.download.id = <-download_id
	}
	fm.download.Progress(bytes)
	downloads <- fm.download
	return bytes, err
}

func (fm FileMonitor) Seek(offset int64, whence int) (int64, error) {
	return fm.f.Seek(offset, whence)
}

type FileSystemMonitor string

func (fsm FileSystemMonitor) Open(name string) (http.File, error) {
	fmt.Printf("FileSystemMonitor started: %s\n", name)
	f, err := http.Dir(fsm).Open(name)
	return &FileMonitor{f, FileDownload{id: "", name: name, size: 0.00, downloaded: 0.00, progress: 0.00}}, err
}

func showStats(download_list map[string]FileDownload) {
	fmt.Println("==========================================================")
	for _, download := range download_list {
		fmt.Printf("%s (%s)\n", download.name, strconv.FormatFloat(download.progress, 'f', 2, 64))
	}
	fmt.Println("==========================================================")
}

func serveStats() {
	download_list := make(map[string]FileDownload)
	for {
		download := <-downloads

		download_list[download.id] = download
		showStats(download_list)
	}
}

func serveIds() {
	h := sha1.New()
	c := []byte(time.Now().String())
	for {
		h.Write(c)
		download_id <- fmt.Sprintf("%x", h.Sum(nil))
	}
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Print(err)
	}

	var address = flag.String("a", "127.0.0.1:6060", "IP address and port to listen on")
	var directory = flag.String("d", cwd, "Target folder for sharing. Defaults to current working directory")
	flag.Parse()

	go serveIds()

	http.Handle("/", http.FileServer(FileSystemMonitor(*directory)))

	fmt.Printf("Binded to address: %s\n", *address)
	fmt.Printf("Sharing          : %s\n", *directory)

	go serveStats()

	err = http.ListenAndServe(*address, nil)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
