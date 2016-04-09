package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {

	commandArgs := os.Args
	var inputPath string

	if len(commandArgs) > 1 {
		inputPath, _ = filepath.Abs(commandArgs[1])
	} else {
		inputPath, _ = filepath.Abs("./")
	}

	port := 8080
	portFree := checkPortFree(port)
	isHtml := isHtmlFile(inputPath)

	var fServer *fileServer
	var err error

	if portFree {
		// Create a new manage
		var manager *managerServer = new(managerServer)

		var serverPath string
		if isHtml {
			serverPath = filepath.Dir(inputPath)
		} else {
			serverPath = inputPath
		}

		fServer, err = manager.addFileServer(serverPath)
		checkErr(err)
		if isHtml {
			url := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, filepath.Base(inputPath))
			openWebPage(url)
			fmt.Println("ishtml")
		}
		err = manager.listen(8080)
		checkErr(err)
	} else {
		// Send request to add server
		fServer = requestNewFileServer(port, formatRootPath(inputPath))
		if isHtml {
			url := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, filepath.Base(inputPath))
			openWebPage(url)
		}
	}

}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func intInSlice(integer int, list []int) bool {
	for _, v := range list {
		if v == integer {
			return true
		}
	}
	return false
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func isHtmlFile(path string) bool {
	exts := strings.Split(path, ".")
	ext := strings.ToLower(exts[len(exts)-1])
	htmlExts := []string{"html", "htm"}
	return stringInSlice(ext, htmlExts)
}

func formatRootPath(path string) string {
	basePath := filepath.Base(path)
	if strings.Contains(basePath, ".") {
		return filepath.Dir(path)
	}
	return path
}

func openWebPage(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		fmt.Println(url)
		err = exec.Command("xdg-open", url).Run()
	case "windows", "darwin":
		err = exec.Command("open", url).Run()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
