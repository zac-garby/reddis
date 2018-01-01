package lib

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

// A User is an in-memory user loaded from the database.
type User struct {
	ID           int
	Name         string
	PasswordHash string
	Posts        []*Post
}

// FetchUser attempts to fetch a user of the given id from the database.
func FetchUser(id int, rdb *redis.Client) (*User, error) {
	var (
		key      = fmt.Sprintf("user:%d", id)
		postsKey = fmt.Sprintf("%s:posts", key)
		user     = new(User)
		posts    []*Post
	)

	if !Exists(key, rdb) {
		return nil, errors.New("fetch user: key doesn't exist")
	}

	if !Exists(postsKey, rdb) {
		return nil, errors.New("fetch user: posts key doesn't exist")
	}

	user.ID = id

	name, err := rdb.HGet(key, "name").Result()
	if err != nil {
		return nil, err
	}
	user.Name = name

	hash, err := rdb.HGet(key, "pw-hash").Result()
	if err != nil {
		return nil, err
	}
	user.PasswordHash = hash

	postIDs, err := rdb.SMembers(postsKey).Result()
	if err != nil {
		return nil, err
	}

	for _, pidString := range postIDs {
		pid, err := strconv.Atoi(pidString)
		if err != nil {
			return nil, err
		}

		node, err := FetchPostTree(pid, 1, rdb)
		if err != nil {
			return nil, err
		}

		post, ok := node.(*Post)
		if !ok {
			return nil, errors.New("fetch user: non-post from FetchPostTree")
		}

		posts = append(posts, post)
	}

	user.Posts = posts

	return user, nil
}
