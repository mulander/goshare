package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type FileMonitor struct {
	f          http.File
	downloaded float64
}

func (fm *FileMonitor) Close() error {
	fm.downloaded = 0
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
	fm.progress(bytes)
	return bytes, err
}

func (fm *FileMonitor) progress(downloaded int) {
	fileInfo, _ := fm.f.Stat()
	fm.downloaded = fm.downloaded + float64(downloaded)
	progress := (fm.downloaded / float64(fileInfo.Size())) * 100
	fmt.Printf("Downloading file: %s (%s)\n", fileInfo.Name(), strconv.FormatFloat(progress, 'f', 2, 64))
}

func (fm FileMonitor) Seek(offset int64, whence int) (int64, error) {
	return fm.f.Seek(offset, whence)
}

type FileSystemMonitor string

func (fsm FileSystemMonitor) Open(name string) (http.File, error) {
	fmt.Printf("FileSystemMonitor started: %s\n", name)
	f, err := http.Dir(fsm).Open(name)
	return &FileMonitor{f, 0.00}, err
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Print(err)
	}

	var address = flag.String("a", "127.0.0.1:6060", "IP address and port to listen on")
	var directory = flag.String("d", cwd, "Target folder for sharing. Defaults to current working directory")
	flag.Parse()
	http.Handle("/", http.FileServer(FileSystemMonitor(*directory)))

	fmt.Printf("Binded to address: %s\n", *address)
	fmt.Printf("Sharing          : %s\n", *directory)
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
