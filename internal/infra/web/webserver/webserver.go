package webserver

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type WebServer struct {
	Router        chi.Router
	Handlers      []Route
	WebServerPort string
}

type Route struct {
	path     string
	httpVerb HttpVerb
	handler  http.HandlerFunc
}

func NewRoute(path string, httpVerb HttpVerb, handler http.HandlerFunc) Route {
	return Route{
		path:     path,
		httpVerb: httpVerb,
		handler:  handler,
	}
}

func NewWebServer(serverPort string) *WebServer {
	return &WebServer{
		Router:        chi.NewRouter(),
		Handlers:      make([]Route, 0),
		WebServerPort: serverPort,
	}
}

func (s *WebServer) AddHandler(route Route) {
	s.Handlers = append(s.Handlers, route)
}

func (s *WebServer) Start() {
	s.Router.Use(middleware.Logger)
	for r := range s.Handlers {
		switch s.Handlers[r].httpVerb {
		case GET:
			s.Router.Get(s.Handlers[r].path, s.Handlers[r].handler)
		case POST:
			s.Router.Post(s.Handlers[r].path, s.Handlers[r].handler)
		default:
			panic("Invalid HTTP verb")
		}
	}
	err := http.ListenAndServe(s.WebServerPort, s.Router)
	if err != nil {
		log.Panic(err)
	}
}
