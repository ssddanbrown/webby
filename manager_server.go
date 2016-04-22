package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/howeyc/fsnotify"
	"golang.org/x/net/websocket"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type managerServer struct {
	FileServers    []fileServer
	WatchedFolders []string
	fileWatcher    *fsnotify.Watcher
	changedFiles   chan string
	sockets        []*websocket.Conn
	lastFileChange int64
	Port           int
	NetworkIp      string
}

func (m *managerServer) addFileServer(rootPath string) (*fileServer, error) {
	fServer, err := m.findFileServerByPath(rootPath)
	if err == nil {
		display("Server already running")
		return fServer, err
	}

	fServer, err = startFileServer(rootPath)
	if err != nil {
		return nil, err
	}
	m.FileServers = append(m.FileServers, *fServer)
	display(fmt.Sprintf("Serving files from %s at http://localhost:%d", fServer.RootPath, fServer.Port))

	m.watchFolder(fServer.RootPath)
	return fServer, nil
}

func (m *managerServer) findFileServerByPath(rootPath string) (*fileServer, error) {
	searchPath, err := filepath.Abs(rootPath)
	checkErr(err)

	for i := 0; i < len(m.FileServers); i++ {
		if m.FileServers[i].RootPath == searchPath {
			return &m.FileServers[i], nil
		}
	}

	return nil, errors.New("Path not found")
}

func (m *managerServer) listen(port int) error {
	m.startFileWatcher()
	m.Port = port
	m.NetworkIp = getLocalIp()
	handler := m.getManagerRouting()
	http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
	return nil
}

func (m *managerServer) sendReloadSignal(file string) {
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

	devlog("File changed: " + file)
}

func (m *managerServer) handleFileChange(filePath string) {

	// Prevent duplicate changes
	currentTime := time.Now().UnixNano()
	if (currentTime-m.lastFileChange)/1000000 < 10 {
		return
	}

	// Ignore git directories
	if strings.Contains(filePath, ".git") {
		devlog("GITCHANGE")
		return
	}

	m.lastFileChange = currentTime
	m.changedFiles <- filePath
}

func (m *managerServer) startFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	checkErr(err)

	m.fileWatcher = watcher
	m.changedFiles = make(chan string)

	// Process events
	go func() {

		done := make(chan bool)

		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsModify() {
					m.handleFileChange(ev.Name)
				}
			case err := <-watcher.Error:
				devlog("File Watcher Error: " + err.Error())
			}
		}

		// Hang so program doesn't exit
		<-done

		watcher.Close()
	}()

	for i := 0; i < len(m.WatchedFolders); i++ {
		err = watcher.Watch(m.WatchedFolders[i])
		devlog("Adding file watcher to " + m.WatchedFolders[i])
		checkErr(err)
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

func (m *managerServer) watchFolder(folderPath string) error {
	if !stringInSlice(folderPath, m.WatchedFolders) {
		m.WatchedFolders = append(m.WatchedFolders, folderPath)
		if m.fileWatcher == nil {
			return nil
		}
		err := m.fileWatcher.Watch(folderPath)
		devlog("Adding file watcher to " + folderPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (manager *managerServer) findFileServerById(id int) (int, *fileServer) {
	for index, server := range manager.FileServers {
		if server.ID == id {
			return index, &server
		}
	}
	return -1, nil
}

func (manager *managerServer) getManagerRouting() *http.ServeMux {

	handler := http.NewServeMux()

	// Create new file server instance
	handler.HandleFunc("/create-server", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			rootPath := req.FormValue("root_path")
			fileServer, err := manager.addFileServer(rootPath)
			checkErr(err)
			err = json.NewEncoder(w).Encode(fileServer)
			checkErr(err)
		}
	})

	// Delete a file server instance
	handler.HandleFunc("/delete-server", func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		idVal, err := strconv.Atoi(id)
		checkErr(err)
		index, server := manager.findFileServerById(idVal)
		if server == nil {
			err := errors.New(fmt.Sprintf("Fileserver with id of %d not found", idVal))
			checkErr(err)
			return
		}
		// Destroy server
		server.server.Close()
		manager.fileWatcher.RemoveWatch(server.RootPath)
		for i, path := range manager.WatchedFolders {
			if path == server.RootPath {
				manager.WatchedFolders = append(manager.WatchedFolders[:i], manager.WatchedFolders[i+1:]...)
				break
			}
		}
		manager.FileServers = append(manager.FileServers[:index], manager.FileServers[index+1:]...)

		devlog(fmt.Sprintf("Deleted server with id of %d", server.ID))
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
	})

	// Load compiled in static content
	fileBox := rice.MustFindBox("res")

	// Get manager hompage
	handler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		templString := fileBox.MustString("index.html")
		templ, _ := template.New("Home").Parse(templString)
		templ.Execute(w, manager)
	})

	// Static file serving
	staticServer := http.FileServer(fileBox.HTTPBox())
	handler.Handle("/static/", staticServer)

	// Get LiveReload Script
	handler.Handle("/livereload.js", http.FileServer(fileBox.HTTPBox()))

	// Websocket handling
	wsHandler := manager.getLivereloadWsHandler()
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

func (manager *managerServer) getLivereloadWsHandler() func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {

		manager.sockets = append(manager.sockets, ws)

		for {
			// websocket.Message.Send(ws, "Hello, Client!")
			wsData := new(livereloadResponse)
			err := websocket.JSON.Receive(ws, &wsData)
			if err != nil {
				checkErr(err)
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
				devlog("Sending livereload hello")
				websocket.JSON.Send(ws, response)
			}

		}

	}

}

func getLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
