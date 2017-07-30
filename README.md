# Webby

A windows-focused tiny static web server app written in go. It features live-reload functionality built in for HTML files which also dynamically reloads CSS files.

You can set webby as your default program to open `.html` files. THen, When a `.html` is opened, webby will start then open the page in your default browser.

Webby sits in the windows taskbar. When clicked it will open the management interface via a browser. You can exit webby by right-clicking this icon.

My experience of writing in Go is limited so there's likely to be inefficiencies & bugs.

## Installing and usage

To install webby simply download the `webby.exe` file from the [latest release here](https://github.com/ssddanbrown/webby/releases/latest) and place somewhere in your system.

There are a few ways to use the `webby.exe` program:

1. Set webby as the default HTML file program:
    * Right click a HTML file.
    * "Open with" > "Choose another app".
    * Scroll to bottom > "More Apps".
    * Check "Always use this app" checkbox.
    * Scroll to bottom > "Look for another app on this PC".
    * Select the `webby.exe` file.
2. Drag HTML files onto the `webby.exe` file.
3. From the command line, execute `webby.exe` followed by a html file you want to open:
    ```shell
    webby.exe hello.html
    ```



## Security Considerations

When ran, Webby makes the entire directory structure below the file/folder location it's used available on a port between `8000` & `9000`. Anyone with access to that port on your pc could sniff around and search for files on your system.

Upon the above, Someone with access to port `35729` on your machine could create new servers at any directory they want then access via the above method. 

It is recommended to only use webby behind a firewall on networks you trust due to the above security concerns.

## Project Goals

* Should be focused on static HTML/CSS/JS development.
* Should be super simple to use.
* Should be lightweight.

## Libs used

These awesome libraries has been used in webby:
* github.com/fatih/color
* github.com/howeyc/fsnotify
* golang.org/x/net/websocket
* github.com/GeertJohan/go.rice
* github.com/akavel/rsrc
* github.com/lxn/walk
* github.com/livereload/livereload-js

## License

Webby is licensed under the [MIT License](https://opensource.org/licenses/MIT).