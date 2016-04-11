package main

import (
	"errors"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/howeyc/fsnotify"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"golang.org/x/net/websocket"
	"net/http"
	"path/filepath"
	"strings"
)

type managerServer struct {
	fileServers    []fileServer
	watchedFolders []string
	fileWatcher    *fsnotify.Watcher
	changedFiles   chan string
	sockets        []*websocket.Conn
}

func (m *managerServer) addFileServer(rootPath string) (*fileServer, error) {
	fServer, err := m.findFileServer(rootPath)
	if err == nil {
		display("Server already running")
		return fServer, err
	}

	fServer, err = startFileServer(rootPath)
	if err != nil {
		return nil, err
	}
	m.fileServers = append(m.fileServers, *fServer)
	display(fmt.Sprintf("Serving files from %s at http://localhost:%d", fServer.RootPath, fServer.Port))

	m.watchFolder(fServer.RootPath)
	return fServer, nil
}

func (m *managerServer) findFileServer(rootPath string) (*fileServer, error) {
	searchPath, err := filepath.Abs(rootPath)
	checkErr(err)

	for i := 0; i < len(m.fileServers); i++ {
		if m.fileServers[i].RootPath == searchPath {
			return &m.fileServers[i], nil
		}
	}

	return nil, errors.New("Path not found")
}

func (m *managerServer) listen(port int) error {

	eServer := echo.New()
	setupManagerRouting(m, eServer)

	m.startFileWatcher()

	go func() {
		for f := range m.changedFiles {
			if len(m.sockets) > 0 {
				m.sendReloadSignal(f)
			}
		}
	}()

	eServer.Run(standard.New(fmt.Sprintf("127.0.0.1:%d", port)))
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

	if strings.Contains(filePath, ".git") {
		devlog("GITCHANGE")
		return
	}

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

	for i := 0; i < len(m.watchedFolders); i++ {
		err = watcher.Watch(m.watchedFolders[i])
		devlog("Adding file watcher to " + m.watchedFolders[i])
		checkErr(err)
	}

	return nil
}

func (m *managerServer) watchFolder(folderPath string) error {
	if !stringInSlice(folderPath, m.watchedFolders) {
		m.watchedFolders = append(m.watchedFolders, folderPath)
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

func setupManagerRouting(manager *managerServer, eServer *echo.Echo) {

	// Create new file server instance
	eServer.Post("/create-server", func(c echo.Context) error {
		rootPath := c.FormValue("root_path")
		fileServer, err := manager.addFileServer(rootPath)
		checkErr(err)
		return c.JSON(http.StatusCreated, fileServer)
	})

	assetHandler := http.FileServer(rice.MustFindBox("res").HTTPBox())

	// eServer.Get("/test.html", standard.WrapHandler(assetHandler))
	eServer.Get("/*", standard.WrapHandler(http.StripPrefix("/", assetHandler)))

	wsHandler := getLivereloadWsHandler(manager)
	eServer.Get("/livereload", standard.WrapHandler(websocket.Handler(wsHandler)))
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

func getLivereloadWsHandler(manager *managerServer) func(ws *websocket.Conn) {
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
