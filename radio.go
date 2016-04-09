package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

/**
 * Everything for talking between webby servers
 */

func requestNewFileServer(masterPort int, path string) *fileServer {
	localServer := fmt.Sprintf("http://127.0.0.1:%d/create-server", masterPort)

	form := url.Values{}
	form.Add("root_path", path)
	resp, err := http.PostForm(localServer, form)
	checkErr(err)

	defer resp.Body.Close()
	var serverData fileServer
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&serverData)
	checkErr(err)

	return &serverData
}
