package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ssddanbrown/webby/internal/fileserver"
	"github.com/ssddanbrown/webby/internal/logger"
	"github.com/ssddanbrown/webby/internal/manager"
	"github.com/ssddanbrown/webby/internal/util"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
)

func main() {
	flag.Usage = usage
	isVerbosePtr := flag.Bool("v", false, "Show verbose output")
	flag.Parse()

	if *isVerbosePtr {
		logger.ShowVerboseOutput()
	}

	commandArgs := flag.Args()
	var inputPath string

	if len(commandArgs) > 0 {
		inputPath, _ = filepath.Abs(commandArgs[0])
	} else {
		inputPath, _ = filepath.Abs("./")
	}

	opts := &util.Options{
		LiveReloadEnabled: true,
		ManagerPort:       35729,
	}
	portFree := util.IsPortFree(opts.ManagerPort)

	var fServer *fileserver.FileServer
	var err error

	if portFree {
		// Create a new manager server
		var mgr = manager.NewServer(opts)
		fServer, err = mgr.AddFileServer(inputPath)
		if err != nil {
			logger.Error("Adding initial file server", err)
			return
		}

		if fServer.OpenedFile != "" {
			urlToOpen := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, fServer.OpenedFile)
			_ = openWebPage(urlToOpen)
		}

		logger.Display(fmt.Sprintf("Webby Manager started at http://localhost:%d", opts.ManagerPort))
		err = mgr.Listen()
	} else {
		// Send request to add server
		err, fServer = requestNewFileServer(opts.ManagerPort, inputPath)
		if err != nil {
			logger.Error("Requesting new file server on existing manager", err)
			return
		}

		if util.IsHTMLFile(inputPath) {
			urlToOpen := fmt.Sprintf("http://localhost:%d/%s", fServer.Port, filepath.Base(inputPath))
			_ = openWebPage(urlToOpen)
		}
		logger.Display("Server already open")
	}

	if err != nil {
		logger.Error("Startup error", err)
	}
}

func requestNewFileServer(masterPort int, path string) (error, *fileserver.FileServer) {
	localServer := fmt.Sprintf("http://127.0.0.1:%d/create-server", masterPort)

	form := url.Values{}
	form.Add("root_path", path)
	resp, err := http.PostForm(localServer, form)
	if err != nil {
		return err, nil
	}

	defer resp.Body.Close()
	var serverData fileserver.FileServer
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&serverData)

	return err, &serverData
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
	color.Cyan("  -v 		# Show verbose output")
	flag.PrintDefaults()
}
