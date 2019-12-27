package logger

import "github.com/fatih/color"

var isVerbose bool

// ShowVerboseOutput will activate the verbose status of this logger.
func ShowVerboseOutput() {
	isVerbose = true
}

// Error will print out the given error in a suitable attention-seeking format.
func Error(event string, err error) {
	if isVerbose {
		color.Red("[ERROR] on %s; %s", event, err.Error())
	}
}

// Devlog will show the given text to those with verbose output active.
func Devlog(text string) {
	if isVerbose {
		color.Blue("[DEVLOG] %s", text)
	}
}

// Display will show the given text in a user friendly display format.
func Display(text string) {
	color.Green("%s", text)
}
