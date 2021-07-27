package main

import (
	"context"
	"fmt"
	"net/http"
)

func getRedisAuthorKey(authorId int) string {
	return fmt.Sprintf("author:%d", authorId)
}

func (server *server) persistToStorageLayer(w http.ResponseWriter, author Author, err error, ctx context.Context, authorId int) (error, Author, bool) {
	nonNullValues := authorToKeyValueMap(author)
	_, err = server.client.HSet(ctx, getRedisAuthorKey(authorId), nonNullValues...).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, Author{}, true
	}
	redisReply, err := server.client.HGetAll(ctx, getRedisAuthorKey(authorId)).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, Author{}, true
	}
	replyAuthor, _ := UpdateAuthorFromMapState(author, redisReply)
	return err, replyAuthor, false
}

func (server *server) retrieveFromStorageLayer(err error, ctx context.Context, authorId int) (found bool, storageReplyMap map[string]string) {
	storageReplyMap, err = server.client.HGetAll(ctx, getRedisAuthorKey(authorId)).Result()
	found = true
	if len(storageReplyMap) == 0 {
		found = false
	}
	return
}
