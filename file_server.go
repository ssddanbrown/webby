package main

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
)

type fileServer struct {
	port     int
	rootPath string
	server   net.Listener
}

var usedPorts []int

func startFileServer(rootPath string) (*fileServer, error) {

	port := getFreePort()

	serverRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		http.Serve(listener, http.FileServer(http.Dir(rootPath)))
	}()

	return &fileServer{port: port, rootPath: serverRootPath, server: listener}, nil
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

	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}

	conn.Close()
	return true
}
