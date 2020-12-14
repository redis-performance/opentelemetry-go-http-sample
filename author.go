package main

import "fmt"

// The Author struct represents the data in the JSON/JSONB column.
// We can use struct tags to control how each field is encoded.
type Author struct {
	Name     string `json:"Name"`
	Username string `json:"Username"`
	About    string `json:"About"`
}

func authorToHsetCmdArgs(author Author) []interface{} {
	nonNullValues := make([]interface{}, 0)
	if author.Name != "" {
		nonNullValues = append(nonNullValues, "Name", author.Name)
	}
	if author.Username != "" {
		nonNullValues = append(nonNullValues, "Username", author.Username)
	}
	if author.About != "" {
		nonNullValues = append(nonNullValues, "About", author.About)
	}
	return nonNullValues
}

func getRedisAuthorKey(authorId int) string {
	return fmt.Sprintf("author:%d", authorId)
}

func UpdateAuthorFromRedisReply(author Author, redisReply map[string]string) (updatedAuthor Author, err error) {
	updatedAuthor = author
	if name, ok := redisReply["Name"]; ok {
		updatedAuthor.Name = name
	}
	if username, ok := redisReply["Username"]; ok {
		updatedAuthor.Username = username
	}
	if about, ok := redisReply["About"]; ok {
		updatedAuthor.About = about
	}
	return
}
