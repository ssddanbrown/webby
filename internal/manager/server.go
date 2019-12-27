package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ssddanbrown/webby/internal/fileserver"
	"github.com/ssddanbrown/webby/internal/logger"
	"github.com/ssddanbrown/webby/internal/util"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/howeyc/fsnotify"
	"golang.org/x/net/websocket"
)

type Server struct {
	FileServers    []fileserver.FileServer
	WatchedFolders []string
	fileWatcher    *fsnotify.Watcher
	changedFiles   chan string
	sockets        []*websocket.Conn
	lastFileChange int64
	NetworkIP      string
	Options        *util.Options
}

// NewServer creates a new server instance using the given Options
func NewServer(options *util.Options) *Server {
	server := new(Server)
	server.Options = options
	return server
}

// AddFileServer adds a file, for the given path, to the manager
func (m *Server) AddFileServer(path string) (*fileserver.FileServer, error) {
	fServer, err := m.findFileServerByPath(util.FormatRootPath(path))
	if err == nil {
		logger.Display("Server already running")
		return fServer, err
	}

	fServer, err = fileserver.StartFileServer(path, m.Options)
	if err != nil {
		return nil, err
	}
	m.FileServers = append(m.FileServers, *fServer)
	logger.Display(fmt.Sprintf("Serving files from %s at http://localhost:%d", fServer.RootPath, fServer.Port))

	m.watchFolder(fServer.RootPath)
	return fServer, nil
}

// Listen starts the manager http server on the given port
func (m *Server) Listen() error {

	m.NetworkIP = getLocalIP()
	m.startFileWatcher()

	handler := m.getManagerRouting()
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", m.Options.ManagerPort), handler)
	m.startUI()
	return nil
}

func (m *Server) findFileServerByPath(rootPath string) (*fileserver.FileServer, error) {
	searchPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	for _, fServer := range m.FileServers {
		if fServer.RootPath == searchPath {
			return &fServer, nil
		}
	}

	return nil, errors.New("path not found")
}

func (m *Server) sendReloadSignal(file string) {
	response := livereloadChange{
		Command: "reload",
		Path:    file,
		LiveCSS: true,
	}
	for i := 0; i < len(m.sockets); i++ {
		if !m.sockets[i].IsServerConn() {
			m.sockets[i].Close()
		}

		if m.sockets[i].IsServerConn() {
			websocket.JSON.Send(m.sockets[i], response)
		}
	}

	logger.Devlog("File changed: " + file)
}

func (m *Server) handleFileChange(filePath string) {

	// Prevent duplicate changes
	currentTime := time.Now().UnixNano()
	if (currentTime-m.lastFileChange)/1000000 < 10 {
		return
	}

	// Ignore git directories
	if strings.Contains(filePath, ".git") {
		logger.Devlog("GITCHANGE")
		return
	}

	m.lastFileChange = currentTime
	m.changedFiles <- filePath
}

func (m *Server) startFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	m.fileWatcher = watcher
	m.changedFiles = make(chan string)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsModify() {
					m.handleFileChange(ev.Name)
				}
			case err := <-watcher.Error:
				logger.Devlog("File Watcher Error: " + err.Error())
			}
		}
	}()

	for i := 0; i < len(m.WatchedFolders); i++ {
		err = watcher.Watch(m.WatchedFolders[i])
		logger.Devlog("Adding file watcher to " + m.WatchedFolders[i])
		if err != nil {
			return err
		}
	}

	go func() {
		for f := range m.changedFiles {
			if len(m.sockets) > 0 {
				m.sendReloadSignal(f)
			}
		}
	}()

	return nil
}

func (m *Server) watchFolder(folderPath string) error {
	if !util.StringInSlice(folderPath, m.WatchedFolders) {
		m.WatchedFolders = append(m.WatchedFolders, folderPath)
		if m.fileWatcher == nil {
			return nil
		}
		err := m.fileWatcher.Watch(folderPath)
		logger.Devlog("Adding file watcher to " + folderPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Server) findFileServerById(id int) (int, *fileserver.FileServer) {
	for index, server := range m.FileServers {
		if server.ID == id {
			return index, &server
		}
	}
	return -1, nil
}

func (m *Server) getManagerRouting() *http.ServeMux {

	handler := http.NewServeMux()

	// Create new file server instance
	handler.HandleFunc("/create-server", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			rootPath := req.FormValue("root_path")
			fileServer, err := m.AddFileServer(rootPath)

			if err != nil {
				logger.Error(err)
				return
			}

			err = json.NewEncoder(w).Encode(fileServer)
			if err != nil {
				logger.Error(err)
			}
		}
	})

	// Delete a file server instance
	handler.HandleFunc("/delete-server", func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		idVal, err := strconv.Atoi(id)

		if err != nil {
			logger.Error(err)
			return
		}

		index, server := m.findFileServerById(idVal)
		if server == nil {
			err := fmt.Errorf("Fileserver with ID of %d not found", idVal)
			logger.Error(err)
			return
		}

		// Destroy server
		server.Destroy()
		m.fileWatcher.RemoveWatch(server.RootPath)
		for i, path := range m.WatchedFolders {
			if path == server.RootPath {
				m.WatchedFolders = append(m.WatchedFolders[:i], m.WatchedFolders[i+1:]...)
				break
			}
		}
		m.FileServers = append(m.FileServers[:index], m.FileServers[index+1:]...)

		logger.Devlog(fmt.Sprintf("Deleted server with id of %d", server.ID))
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
	})

	// Toggle livereload on/off
	handler.HandleFunc("/toggle-livereload", func(w http.ResponseWriter, req *http.Request) {
		m.Options.LiveReloadEnabled = !m.Options.LiveReloadEnabled
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
	})

	// Load compiled in static content
	fileBox := rice.MustFindBox("../../res")

	// Get manager homepage
	handler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		templString := fileBox.MustString("index.html")
		templ, _ := template.New("Home").Parse(templString)
		templ.Execute(w, m)
	})

	// Static file serving
	staticServer := http.FileServer(fileBox.HTTPBox())
	handler.Handle("/static/", staticServer)

	// Get LiveReload Script
	handler.Handle("/livereload.js", http.FileServer(fileBox.HTTPBox()))

	// Websocket handling
	wsHandler := m.getLivereloadWsHandler()
	handler.Handle("/livereload", websocket.Handler(wsHandler))

	return handler
}

type livereloadResponse struct {
	Command string `json:"command"`
}

type livereloadHello struct {
	Command    string   `json:"command"`
	Protocols  []string `json:"protocols"`
	ServerName string   `json:"serverName"`
}

type livereloadChange struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	LiveCSS bool   `json:"liveCSS"`
}

func (m *Server) getLivereloadWsHandler() func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {

		m.sockets = append(m.sockets, ws)

		for {
			// websocket.Message.Send(ws, "Hello, Client!")
			wsData := new(livereloadResponse)
			err := websocket.JSON.Receive(ws, &wsData)
			if err != nil {
				logger.Error(err)
				return
			}

			if wsData.Command == "hello" {
				response := livereloadHello{
					Command: "hello",
					Protocols: []string{
						"http://livereload.com/protocols/connection-check-1",
						"http://livereload.com/protocols/official-7",
						"http://livereload.com/protocols/official-8",
						"http://livereload.com/protocols/official-9",
						"http://livereload.com/protocols/2.x-origin-version-negotiation",
					},
					ServerName: "Webby",
				}
				logger.Devlog("Sending livereload hello")
				websocket.JSON.Send(ws, response)
			}

		}

	}

}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipString := ipnet.IP.String()
				if strings.Index(ipString, "192.168") == 0 || strings.Index(ipString, "10.") == 0 {
					return ipString
				}
			}
		}
	}
	return ""
}
