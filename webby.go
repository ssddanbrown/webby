package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
)

var isVerbose bool

func main() {

	flag.Usage = usage
	isVerbosePtr := flag.Bool("v", false, "Show verbose output")
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

	var fServer *fileServer
	var err error

	if portFree {
		// Create a new manager server
		var manager = new(managerServer)
		fServer, err = manager.addFileServer(inputPath)
		checkErr(err)

		if fServer.OpenedFile != "" {
			url := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, fServer.OpenedFile)
			openWebPage(url)
		}

		display(fmt.Sprintf("Webby Manager started at http://localhost:%d", port))
		err = manager.listen(port)
		checkErr(err)
	} else {
		// Send request to add server
		fServer = requestNewFileServer(port, inputPath)
		if fServer.OpenedFile != "" {
			url := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, fServer.OpenedFile)
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
	color.Green("%s", text)
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

func openWebPage(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
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
