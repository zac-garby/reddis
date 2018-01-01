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

// NewUser makes a new user and adds it to the database. It returns the user's
// User struct, or an error.
func NewUser(name, hash string, rdb *redis.Client) (*User, error) {
	if UserExists(name, rdb) {
		return nil, fmt.Errorf("new user: user %s already exists", name)
	}

	id, err := rdb.Incr("uid-counter").Result()
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           int(id),
		Name:         name,
		PasswordHash: hash,
	}

	if err := UpdateUser(user, rdb); err != nil {
		return nil, err
	}

	return user, nil
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

// UpdateUser updates a user in the database if it already exists, or creates a
// new one if it doesn't.
func UpdateUser(user *User, rdb *redis.Client) error {
	if UserExists(user.Name, rdb) {
		return fmt.Errorf("update user: user %s already exists", user.Name)
	}

	var (
		key      = fmt.Sprintf("user:%d", user.ID)
		postsKey = fmt.Sprintf("%s:posts", key)
	)

	// TODO: Make use of the bool return values from HSET
	if _, err := rdb.HSet(key, "name", user.Name).Result(); err != nil {
		return err
	}

	if _, err := rdb.HSet(key, "pw-hash", user.PasswordHash).Result(); err != nil {
		return err
	}

	if _, err := rdb.Del(postsKey).Result(); err != nil {
		return err
	}

	// TODO: Refactor into one call to SADD
	for _, post := range user.Posts {
		if _, err := rdb.SAdd(postsKey, post.ID).Result(); err != nil {
			return err
		}
	}

	if _, err := rdb.HSet("usernames", user.Name, strconv.Itoa(user.ID)).Result(); err != nil {
		return err
	}

	return nil
}

// UserExists checks whether or not a user exists by its username.
func UserExists(name string, rdb *redis.Client) bool {
	exists, err := rdb.HExists("usernames", name).Result()
	if err != nil {
		return false
	}

	return exists
}
