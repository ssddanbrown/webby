package main

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	port := getFreePort()
	// Serve static content from current dir
	http.Handle("/", http.FileServer(http.Dir("./")))

	fmt.Println(fmt.Sprintf("Listening on 127.0.0.1:%d", port))

	serverRootPath, err := filepath.Abs("./")
	checkErr(err)
	fmt.Println(fmt.Sprintf("Serving content from %s", serverRootPath))

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func getFreePort() int {
	portMin := 8000
	portMax := 9000
	currentPort := portMin
	for currentPort <= portMax && !checkPortFree(currentPort) {
		currentPort++
	}
	return currentPort
}

func checkPortFree(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
