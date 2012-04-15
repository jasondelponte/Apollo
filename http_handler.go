package main

import (
	gnws "code.google.com/p/go.net/websocket"
	"fmt"
	gbws "github.com/garyburd/go-websocket/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// HTTP Error Enumerables
type HttpError struct {
	ErrorString string
	CodeNum     int
}

func (h HttpError) Error() string                { return h.ErrorString }
func (h HttpError) Code() int                    { return h.CodeNum }
func (h HttpError) Report(w http.ResponseWriter) { http.Error(w, h.ErrorString, h.CodeNum) }

var (
	ErrHttpResourceNotFound = &HttpError{ErrorString: "Not found", CodeNum: 404}
	ErrHttpMethodNotAllowed = &HttpError{ErrorString: "Method not allowed", CodeNum: 405}
	ErrHttpBadRequeset      = &HttpError{ErrorString: "Bad request", CodeNum: 400}
	ErrHttpInternalError    = &HttpError{ErrorString: "Internal failure", CodeNum: 500}
)

type HttpHandler struct {
	RootURLPath    string
	Addr           string
	Port           uint
	ServeStatic    bool
	templates      *template.Template
	nextWsConnId   uint64
	rootURLPathLen int
	WsConnType     string
}

// Configures the http connection and starts the listender
func (h *HttpHandler) HandleHttpConnection(world *World) {
	h.rootURLPathLen = len(h.RootURLPath + "/")

	h.loadTemplates()

	h.initServeHomeHndlr(h.RootURLPath+"/", world)

	// If the goapp is serving the static files
	if h.ServeStatic {
		h.initServeStaticHndlr(h.RootURLPath + "/assets/")
	}

	// Switch between the different go websocket libraries
	if h.WsConnType == "gb" {
		h.initServeGbWsHndlr(h.RootURLPath+"/ws", world)
	} else {
		h.initServeGnWsHndlr(h.RootURLPath+"/ws", world)
	}

	// Build the address with port if it's provided
	address := h.Addr
	if h.Port != 0 {
		address = fmt.Sprintf("%s:%d", h.Addr, h.Port)
	}

	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Load all the temmplates into memeory
func (h *HttpHandler) loadTemplates() {
	h.templates = template.Must(template.ParseFiles("templates/home.html"))
}

// Network event handler for HTTP trafic. Serves up the 
// home.html file which will allow connection to the websocket
func (h *HttpHandler) initServeHomeHndlr(path string, world *World) {
	tmplData := map[string]string{
		"Host":     "",
		"WsHost":   "",
		"RootPath": h.RootURLPath,
	}

	hostPortRep := regexp.MustCompile(":\\d+$")

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != h.RootURLPath+"/" {
			ErrHttpResourceNotFound.Report(w)
			return
		}
		if r.Method != "GET" {
			ErrHttpMethodNotAllowed.Report(w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Normal http host
		if len(tmplData["Host"]) == 0 {
			tmplData["Host"] = r.Host
			if h.ServeStatic && h.Port != 0 {
				if strings.Contains(r.Host, ":") {
					tmplData["Host"] = hostPortRep.ReplaceAllString(r.Host, fmt.Sprintf(":%d", h.Port))
				} else {
					tmplData["Host"] = fmt.Sprintf("%s:%d", r.Host, h.Port)
				}
			}
		}
		// Host for websockets
		if len(tmplData["WsHost"]) == 0 {
			tmplData["WsHost"] = r.Host
			if h.Port != 0 {
				if strings.Contains(r.Host, ":") {
					tmplData["WsHost"] = hostPortRep.ReplaceAllString(r.Host, fmt.Sprintf(":%d", h.Port))
				} else {
					tmplData["WsHost"] = fmt.Sprintf("%s:%d", r.Host, h.Port)
				}
			}
		}

		h.templates.Execute(w, tmplData)
	})
}

// Simple handler for serving static files
func (h *HttpHandler) initServeStaticHndlr(path string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		asset := r.URL.Path[h.rootURLPathLen:]
		file, err := os.Open(asset)
		if err != nil {
			ErrHttpResourceNotFound.Report(w)
			return
		}
		stat, err := file.Stat()
		if err != nil {
			ErrHttpInternalError.Report(w)
			return
		}
		http.ServeContent(w, r, asset, stat.ModTime(), file)
	})
}

// creats the webocket http upgrade handler requests from the client.
func (h *HttpHandler) initServeGbWsHndlr(path string, world *World) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			ErrHttpMethodNotAllowed.Report(w)
			return
		}
		ws, err := gbws.Upgrade(w, r.Header, "", 1024, 1024)
		if err != nil {
			log.Println("Unable to upgrade connection,", err)
			ErrHttpBadRequeset.Report(w)
			return
		}

		h.kickOffPlayer(NewGbWsConn(h.nextWsConnId, ws), world)
	})
}

// Creates the websocket http upgrade using the go.net websocket version
func (h *HttpHandler) initServeGnWsHndlr(path string, world *World) {
	http.Handle(path, gnws.Handler(func(ws *gnws.Conn) {
		h.kickOffPlayer(NewGnWsConn(h.nextWsConnId, ws), world)
	}))
}

func (h *HttpHandler) kickOffPlayer(conn Connection, world *World) {
	player := NewPlayer(h.nextWsConnId, conn)
	h.nextWsConnId += 1

	defer func() {
		log.Println("Player", player.GetId(), " connection closing, unregistering")
		world.unregister <- player
	}()

	go conn.WritePump()
	world.register <- player

	// Read pump will hold the connection open until we are finished with it.
	conn.ReadPump()

}
