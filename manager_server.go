package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

type managerServer struct {
	fileServers []fileServer
}

func (m *managerServer) addFileServer(rootPath string) error {
	fServer, err := startFileServer(rootPath)
	if err != nil {
		return err
	}
	m.fileServers = append(m.fileServers, *fServer)
	fmt.Println(fmt.Sprintf("Listening on port %d", fServer.port))
	return nil
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
		manager.addFileServer(rootPath)
		return nil
	})

}
