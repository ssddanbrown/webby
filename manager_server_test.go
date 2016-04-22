package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func getTestServer() (*managerServer, *httptest.Server) {
	m := new(managerServer)
	m.startFileWatcher()
	handler := m.getManagerRouting()
	server := httptest.NewServer(handler)
	return m, server
}

func TestStaticServing(t *testing.T) {
	_, server := getTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/livereload.js")
	if err != nil {
		t.Error("Request to local manager server failed")
	}

	if resp.StatusCode != 200 || resp.ContentLength < 500 {
		t.Error("Local request to livereload script failed")
	}
}

func TestCreateServerRequest(t *testing.T) {
	m, server := getTestServer()
	defer server.Close()

	tempDir, _ := ioutil.TempDir("", "webby-test")
	defer os.RemoveAll(tempDir)

	resp, err := http.PostForm(server.URL+"/create-server",
		url.Values{"root_path": {tempDir}})
	// body, _ := ioutil.ReadAll(resp.Body)
	// t.Log(string(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Error("Create Server request failed")
	}

	if len(m.FileServers) != 1 {
		t.Error("A file server was not created")
		t.FailNow()
	}

	if m.FileServers[0].RootPath != tempDir {
		t.Error("New fileserver path is incorrect")
	}
}

func TestDeleteServerRequest(t *testing.T) {
	m, server := getTestServer()
	defer server.Close()

	// Create a file server
	tempDir, _ := ioutil.TempDir("", "webby-test")
	defer os.RemoveAll(tempDir)

	resp, err := http.PostForm(server.URL+"/create-server",
		url.Values{"root_path": {tempDir}})

	if len(m.FileServers) != 1 {
		t.Error("A file server was not created")
		t.FailNow()
	}

	if len(m.WatchedFolders) != 1 {
		t.Error("A file watched was not created")
		t.FailNow()
	}

	fileServerId := m.FileServers[0].ID

	resp, err = http.Get(server.URL + fmt.Sprintf("/delete-server?id=%d", fileServerId))
	if err != nil {
		t.Fatal(err.Error())
	} else if resp.StatusCode != http.StatusOK {
		t.Errorf("Delete request returned reponse code %d", resp.StatusCode)
	}

	if resp.Request.URL.String() != server.URL+"/" {
		t.Error("Response did not redirect")
	}

	if len(m.FileServers) > 0 {
		t.Error("The manager fileserver was not deleted")
	}

	if len(m.WatchedFolders) > 0 {
		t.Error("The manager folder watcher was not deleted")
	}

}
