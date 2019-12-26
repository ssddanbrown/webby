package main

import (
	"flag"
	"fmt"
	"github.com/ssddanbrown/webby/internal/logger"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
)


func main() {
	flag.Usage = usage
	isVerbosePtr := flag.Bool("v", false, "Show verbose output")
	flag.Parse()

	if *isVerbosePtr {
		logger.ShowVerboseOutput();
	}

	commandArgs := flag.Args()
	var inputPath string

	if len(commandArgs) > 0 {
		inputPath, _ = filepath.Abs(commandArgs[0])
	} else {
		inputPath, _ = filepath.Abs("./")
	}

	port := 35729
	portFree := isPortFree(port)

	var fServer *fileServer
	var err error

	if portFree {
		// Create a new manager server
		var manager = new(managerServer)
		fServer, err = manager.addFileServer(inputPath)
		checkErr(err)

		if fServer.OpenedFile != "" {
			url := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, fServer.OpenedFile)
			_ = openWebPage(url)
		}

		logger.Display(fmt.Sprintf("Webby Manager started at http://localhost:%d", port))
		err = manager.listen(port)
		checkErr(err)
	} else {
		// Send request to add server
		fServer = requestNewFileServer(port, inputPath)
		if isHTMLFile(inputPath) {
			url := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, filepath.Base(inputPath))
			_ = openWebPage(url)
		}
		logger.Display("Server already open")
	}

}

func checkErr(err error) {
	if err != nil {
		logger.Error(err)
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
