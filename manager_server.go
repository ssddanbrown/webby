package main

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"net/http"
	"path/filepath"
)

type managerServer struct {
	fileServers []fileServer
}

func (m *managerServer) addFileServer(rootPath string) (*fileServer, error) {

	fServer, err := m.findFileServer(rootPath)
	if err == nil {
		fmt.Println("Server already running")
		return fServer, err
	}

	fServer, err = startFileServer(rootPath)
	if err != nil {
		return nil, err
	}
	m.fileServers = append(m.fileServers, *fServer)
	fmt.Println(fmt.Sprintf("Serving files from %s at http://localhost:%d", fServer.RootPath, fServer.Port))
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

	eServer.Run(standard.New(fmt.Sprintf("127.0.0.1:%d", port)))
	return nil
}

func setupManagerRouting(manager *managerServer, eServer *echo.Echo) {

	// Create new file server instance
	eServer.Post("/create-server", func(c echo.Context) error {
		rootPath := c.FormValue("root_path")
		fmt.Println(rootPath)
		fileServer, err := manager.addFileServer(rootPath)
		checkErr(err)
		return c.JSON(http.StatusCreated, fileServer)
	})

}
