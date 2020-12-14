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
	fmt.Println(vars["id"])
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
		nonNullValues := authorToHsetCmdArgs(author)
		_, err = server.client.HSet(ctx, getRedisAuthorKey(authorId), nonNullValues...).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		redisReply, err := server.client.HGetAll(ctx, getRedisAuthorKey(authorId)).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		replyAuthor, _ := UpdateAuthorFromRedisReply(author, redisReply)
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
		redisReply, err := server.client.HGetAll(ctx, getRedisAuthorKey(authorId)).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		replyAuthor, _ := UpdateAuthorFromRedisReply(Author{}, redisReply)
		encodedAuthor, err := json.Marshal(replyAuthor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, string(encodedAuthor))
	}
}
