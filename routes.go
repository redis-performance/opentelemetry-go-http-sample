package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"io"
	"net/http"
	"strconv"
)

func (server *server) routes() {
	server.router.HandleFunc("/author/{id:[0-9]+}", otelhttp.NewHandler(server.handleAuthorGet(), "AuthorGet").ServeHTTP).Methods(http.MethodGet)
	server.router.HandleFunc("/author/{id:[0-9]+}", otelhttp.NewHandler(server.handleAuthorPostOrPut(), "AuthorPost").ServeHTTP).Methods(http.MethodPost)
	server.router.HandleFunc("/author/{id:[0-9]+}", otelhttp.NewHandler(server.handleAuthorPostOrPut(), "AuthorPut").ServeHTTP).Methods(http.MethodPut)
}

func getIdFromRequest(req *http.Request) (id int, err error) {
	vars := mux.Vars(req)
	if vars["id"] != "" {
		id, _ = strconv.Atoi(vars["id"])
	} else {
		err = fmt.Errorf("id property not present in request")
	}
	return
}

func (server *server) handleAuthorPostOrPut() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		authorId, err := getIdFromRequest(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields() // catch unwanted fields
		var author Author
		err = decoder.Decode(&author)
		if err != nil {
			// bad JSON or unrecognized json field
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := req.Context()
		err, replyAuthor, done := server.persistToStorageLayer(w, author, err, ctx, authorId)
		if done {
			return
		}
		encodedAuthor, err := json.Marshal(replyAuthor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, string(encodedAuthor))
	}
}

func (server *server) handleAuthorGet() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		authorId, err := getIdFromRequest(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := req.Context()
		found, storageReplyMap := server.retrieveFromStorageLayer(err, ctx, authorId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !found {
			http.Error(w, fmt.Sprintf("The resource Author with id %d does not exist", authorId), http.StatusNotFound)
			return
		}
		replyAuthor, _ := UpdateAuthorFromMapState(Author{}, storageReplyMap)
		encodedAuthor, err := json.Marshal(replyAuthor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, string(encodedAuthor))
	}
}
