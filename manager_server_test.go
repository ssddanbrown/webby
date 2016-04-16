package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func getTestServer() (*managerServer, *httptest.Server) {
	m := new(managerServer)
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
	if err != nil || resp.StatusCode != 200 {
		t.Error("Create Server request failed")
	}

	if len(m.fileServers) != 1 {
		t.Error("A file server was not created")
		t.FailNow()
	}

	if m.fileServers[0].RootPath != tempDir {
		t.Error("New fileserver path is incorrect")
	}
}
