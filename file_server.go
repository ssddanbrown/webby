package main

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
)

type fileServer struct {
	ID         int    `json:"id"`
	Port       int    `json:"port"`
	RootPath   string `json:"path"`
	OpenedFile string `json:"file"`
	server     net.Listener
}

var usedPorts []int
var idCounter int

func startFileServer(path string) (*fileServer, error) {

	rootPath := formatRootPath(path)
	port := getFreePort()
	file := ""

	if isHTMLFile(path) {
		file = filepath.Base(path)
	}

	serverRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		http.Serve(listener, http.FileServer(http.Dir(rootPath)))
	}()

	idCounter++

	return &fileServer{ID: idCounter, Port: port, RootPath: serverRootPath, OpenedFile: file, server: listener}, nil
}

func (fs *fileServer) Url() string {
	return fmt.Sprintf("http://localhost:%d", fs.Port)
}

func getFreePort() int {
	portMin := 8000
	portMax := 9000
	currentPort := portMin

	for currentPort <= portMax && !checkPortFree(currentPort) {
		currentPort++
	}

	usedPorts = append(usedPorts, currentPort)
	return currentPort
}

func checkPortFree(port int) bool {

	if intInSlice(port, usedPorts) {
		return false
	}

	conn, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

func formatRootPath(path string) string {
	basePath := filepath.Base(path)
	if strings.Contains(basePath, ".") {
		return filepath.Dir(path)
	}
	return path
}

func isHTMLFile(path string) bool {
	exts := strings.Split(path, ".")
	ext := strings.ToLower(exts[len(exts)-1])
	htmlExts := []string{"html", "htm"}
	return stringInSlice(ext, htmlExts)
}
