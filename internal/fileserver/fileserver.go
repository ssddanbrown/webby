package fileserver

import (
	"fmt"
	"github.com/ssddanbrown/webby/internal/logger"
	"github.com/ssddanbrown/webby/internal/util"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

type FileServer struct {
	ID         int    `json:"id"`
	Port       int    `json:"port"`
	RootPath   string `json:"path"`
	OpenedFile string `json:"file"`
	server     net.Listener
	options    *util.Options
}

var usedPorts = make(map[int]bool)
var idCounter int

// StartFileServer starts a new file server and returns the instance
func StartFileServer(path string, options *util.Options) (*FileServer, error) {

	rootPath := util.FormatRootPath(path)
	port := getFreePort()
	file := ""

	if util.IsHTMLFile(path) {
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

	go listenAndServe(options, listener, rootPath, serverRootPath)

	idCounter++

	return &FileServer{
		ID:         idCounter,
		Port:       port,
		RootPath:   serverRootPath,
		OpenedFile: file,
		server:     listener,
		options:    options,
	}, nil
}

// Url provides the direct URL for the root of the started server
func (fs *FileServer) Url() string {
	return fmt.Sprintf("http://localhost:%d", fs.Port)
}

// Destroy the file server and take it offline
func (fs *FileServer) Destroy() {
	err := fs.server.Close()
	if err != nil {
		logger.Error(err)
	}
}

func listenAndServe(options *util.Options, listener net.Listener, rootPath string, serverRootPath string) {
	handler := http.NewServeMux()
	staticHandler := http.FileServer(http.Dir(rootPath))

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rPath := r.URL.Path
		fPath := filepath.Join(serverRootPath, rPath)
		logger.Devlog(fPath)

		// Check if an index html file is being served and update request path if so
		if rPath == "/" {
			indexFilePath := filepath.Join(serverRootPath, rPath, "index.html")
			_, err := os.Stat(indexFilePath)
			if err == nil {
				fPath = indexFilePath
			}
		}

		// Prevent caching of served files
		w.Header().Add("Cache-Control", "no-cache")

		// Inject livereload script if serving a HTML file
		if util.IsHTMLFile(fPath) && options.LiveReloadEnabled {
			_, err := os.Stat(fPath)
			if err == nil {
				file, err := os.Open(fPath)
				if err == nil {
					w.Header().Add("Content-Type", "text/html")
					io.Copy(w, file)
					fmt.Fprintf(w, "\n<script src=\"http://localhost:%d/livereload.js\"></script>\n", options.ManagerPort)
					return
				}
			}
		}

		// Otherwise serve a static file
		staticHandler.ServeHTTP(w, r)
	})

	http.Serve(listener, handler)
}

func getFreePort() int {
	portMin := 8000
	portMax := 9000
	currentPort := portMin

	for currentPort <= portMax && !isPortFree(currentPort) {
		currentPort++
	}

	usedPorts[currentPort] = true
	return currentPort
}

func isPortFree(port int) bool {

	if usedPorts[port] {
		return false
	}

	return util.IsPortFree(port)
}
