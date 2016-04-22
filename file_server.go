package main

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
)

type fileServer struct {
	ID       int          `json:"id"`
	Port     int          `json:"port"`
	RootPath string       `json:"path"`
	server   net.Listener `json:"-"`
}

var usedPorts []int
var idCounter int

func startFileServer(rootPath string) (*fileServer, error) {

	port := getFreePort()

	serverRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		http.Serve(listener, http.FileServer(http.Dir(rootPath)))
	}()

	idCounter++

	return &fileServer{ID: idCounter, Port: port, RootPath: serverRootPath, server: listener}, nil
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

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}

	conn.Close()
	return true
}
