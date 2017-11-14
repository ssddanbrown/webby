package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"

	rice "github.com/GeertJohan/go.rice"
)

func (m *managerServer) startUI() {
	m.setupTrayIcon()
}

func (m *managerServer) setupTrayIcon() {
	// We need either a walk.MainWindow or a walk.Dialog for their message loop.
	// We will not make it visible in this example, though.
	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	// We load our icon from a file.
	fileBox := rice.MustFindBox("res")
	iconBytes := fileBox.MustBytes("webby_icon16.png")

	pngData, err := png.Decode(bytes.NewReader(iconBytes))
	checkErr(err)

	icon, err := walk.NewIconFromImage(pngData)
	if err != nil {
		log.Fatal(err)
	}

	// Create the notify icon and make sure we clean it up on exit.
	ni, err := walk.NewNotifyIcon()
	if err != nil {
		log.Fatal(err)
	}
	defer ni.Dispose()

	// Set the icon and a tool tip text.
	if err := ni.SetIcon(icon); err != nil {
		log.Fatal(err)
	}
	if err := ni.SetToolTip("Click for info or use the context menu to exit."); err != nil {
		log.Fatal(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}

		openWebPage(fmt.Sprintf("http://localhost:%d", m.Port))
	})

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}

	// Now that the icon is visible, we can bring up an info balloon.
	if err := ni.ShowInfo("Webby Started", "Click the tray icon to open the management interface."); err != nil {
		log.Fatal(err)
	}

	// Run the message loop.
	mw.Run()
}
