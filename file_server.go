package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type fileServer struct {
	ID         int    `json:"id"`
	Port       int    `json:"port"`
	RootPath   string `json:"path"`
	OpenedFile string `json:"file"`
	manager    *managerServer
	server     net.Listener
}

var usedPorts []int
var idCounter int

func startFileServer(path string, manager *managerServer) (*fileServer, error) {

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

		handler := http.NewServeMux()
		staticHandler := http.FileServer(http.Dir(rootPath))

		handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			rPath := r.URL.Path
			fPath := filepath.Join(serverRootPath, rPath)

			// Prevent caching of served files
			w.Header().Add("Cache-Control", "no-cache")

			// Inject livereload script if serving a HTML file
			if isHTMLFile(fPath) && manager.LiveReload {
				_, err := os.Stat(fPath)
				if err == nil {
					file, err := os.Open(fPath)
					checkErr(err)
					w.Header().Add("Content-Type", "text/html")
					io.Copy(w, file)
					fmt.Fprintf(w, "\n<script src=\"http://localhost:%d/livereload.js\"></script>\n", manager.Port)
					return
				}
			}

			staticHandler.ServeHTTP(w, r)
		})

		http.Serve(listener, handler)
	}()

	idCounter++

	return &fileServer{ID: idCounter, Port: port, RootPath: serverRootPath, OpenedFile: file, manager: manager, server: listener}, nil
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
