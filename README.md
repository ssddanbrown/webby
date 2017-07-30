# Webby

A windows-focused tiny static webserver app written in go. It features live-reload functionality built in for HTML files which also dynamically reloads CSS files.

You can set webby as your default program to open `.html` files. THen, When a `.html` is opened, webby will start then open the page in your default browser.

Webby sits in the windows taskbar. When clicked it will open the managment interface via a browser. You can exit webby by right-clicking this icon.

Just to be clear - I currently have very little experience in writing Go so please allow a little slack in the code quality.
If you spot areas to be improved please create a pull request.


### Security Considerations

When ran, Webby makes the entire directory structure below the file/folder location it's used available on a port between `8000` & `9000`. Anyone with access to that port on your pc could sniff around and search for files on your system.

Upon the above, Someone with access to port `35729` on your machine could create new servers at any directory they want then access via the above method. 

It is recommended to only use webby behind a firewall on networks you trust due to the above security concerns.

### Libs used

These awesome libraries has been used in webby:
* github.com/fatih/color
* github.com/howeyc/fsnotify
* golang.org/x/net/websocket
* github.com/GeertJohan/go.rice
* github.com/akavel/rsrc
* github.com/lxn/walk
* github.com/livereload/livereload-js

### License

Webby is licensed under the [MIT License](https://opensource.org/licenses/MIT).