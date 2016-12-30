package service

import (
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

// NewServer configures and returns a server.
func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(AuthMiddleware))
	mx := mux.NewRouter()
	db := &dataHandler{}
	initRoutes(mx, formatter, db)
	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render, database Database) {
	mx.HandleFunc("/friends/request", postAddFriendHandler(formatter, database)).Methods("POST")
	mx.HandleFunc("/friends/{request_id}/reject", rejectRequestHandler(formatter, database)).Methods("PUT")
	mx.HandleFunc("/friends/{request_id}/accept", acceptRequestHandler(formatter, database)).Methods("PUT")
	mx.HandleFunc("/friends", getFriendsHandler(formatter, database)).Methods("GET")
}
