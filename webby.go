package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var isVerbose bool

func main() {

	flag.Usage = usage
	isVerbosePtr := flag.Bool("v", false, "Show a verbose output")
	flag.Parse()

	isVerbose = *isVerbosePtr

	commandArgs := flag.Args()
	var inputPath string

	if len(commandArgs) > 0 {
		inputPath, _ = filepath.Abs(commandArgs[0])
	} else {
		inputPath, _ = filepath.Abs("./")
	}

	port := 35729
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
		}
		err = manager.listen(port)
		checkErr(err)
	} else {
		// Send request to add server
		fServer = requestNewFileServer(port, formatRootPath(inputPath))
		if isHtml {
			url := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, filepath.Base(inputPath))
			openWebPage(url)
		}

		display("Server already open")
	}

}

func checkErr(err error) {
	if err != nil && isVerbose {
		color.Red("[ERROR] %s", err.Error())
	}
}

func devlog(text string) {
	if isVerbose {
		color.Blue("[DEVLOG] %s", text)
	}
}

func display(text string) {
	if isVerbose {
		color.Green("[LOG] %s", text)
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
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func usage() {
	color.Blue("Usage of webby:")
	color.Green("  webby [options] [<File or folder path>]")
	fmt.Println("")
	color.Blue("Examples:")
	color.Cyan("  webby ./ 		# Starts a file server in the current directory")
	color.Cyan("  webby test.html 	# As above and opens up test.html in the browser")
	fmt.Println("")
	color.Blue("Options:")
	flag.PrintDefaults()
}
