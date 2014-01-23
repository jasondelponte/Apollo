[![baby-gopher](https://raw2.github.com/drnic/babygopher-site/gh-pages/images/babygopher-badge.png)](http://www.babygopher.org)

## Installing

To install the package follow the normal go package installtion.  The go.net websocket package may need to be installed manually, and mercurial is require to pull from that repo.

```bash
go get code.google.com/p/go.net/websocket
go get github.com/jasondelponte/Apollo
```

## Command line args
* -p PortNum - The port the app will listen on, default is blank which should mean port 80
* -a Address - The IP address the app will listen on, default is blank, which whould mean "localhost"
* -r RootURLPath - The root URL path that you'll use to access the app at. eg. "jasondelponte.com/goapps/apollo/" would be "-r /goapps/apollo".  Blank is the default which translates into "/"
* -s true|false - Sets if the webapp should serve up the resources in the "assets" directory its self. Default is false, and expects some other service to serve the files in the assets directory.
* -w gb|gn - Sets which websocket library to use. **gb** (go.net/websocket) which supports version 13 and 8. **gb** (gauryburd/go-websocket) which supports version 13

example:
```bash
Apollo -r="/goapps/apollo" -a="192.168.1.128" -s=true -p=8080
```

## compatibility ##
I've verified the canvas and websockets work with the below platforms. IE 9 doesn't support websockets, and i dont have IE 10 installed.
```
iOS 4.3 + (works extremely well with its hardware acceleration)
Android 4.0 w/ Chrome browser (stock browser doesn't support websockets, but canvas performance is really bad)
Chrome 17 +
FireFox 11 +
```

## Why?
I'm very interested in Go Lang and decided to try building simple applications with it.  This app exercises some go's paralle event driven capabilities. While at the same time I'm playing around with the json encode/deocder, and websockets.

I haven't added the input controls from the client to the backend yet.  To be honest I'm not even really sure where I want to take this yet.  But I think I've found a good starting point.

## Dependencies
The only packages Apollo depends on at the moment are: go.net's websocket and guryburd/go-websocket.  I currently am including both websocket packages until I can evaluate them better.  By default go.net's websocket is being used, but if you want to switch to guryburd's websocket use the command line arg, "-w gb".
