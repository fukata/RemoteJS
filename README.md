# RemoteJS

## What's this
WEB page displayed on a remote server and return the results after running the JS.

## JS Executor
Firefox on the request to the virtual frame buffer and a request is received the JS virtual frame buffer management.

### Compile
	make

### Run
	make run

### Configuration
	vi appconfig.go

### Dependencies
	goinstall github.com/garyburd/go-mongo

## Firefox UserScript
Run the JS received and the results registered in the server. Install Firefox to run on a virtual frame buffer. (Otherwise, you need to install the Greasemonkey)

### Configuration
	vi remotejs.user.js

### window.close
Firefox openurl "about:config" and change dom.allow_scripts_to_close_windows is true.
	dom.allow_scripts_to_close_windows true

## API Server
To register the run JS API server.

### Run
	node app.js 3030

### Test
	nodeunit test

### Configuration
	vi appconfig.js

