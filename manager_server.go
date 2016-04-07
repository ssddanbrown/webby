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
	return nil
}

func (*managerServer) listen(port int) error {

	eServer := echo.New()

	// listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	// if err != nil {
	// 	return err
	// }

	// sMux := http.NewServeMux()
	// sMux.Handle("/add", handler)

	// http.Serve(listener, http.DefaultServeMux)

	eServer.Run(standard.New(fmt.Sprintf("127.0.0.1:%d", port)))

	return nil
}
