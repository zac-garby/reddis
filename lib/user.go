package lib

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	sessIDLength   = 32
	sessIDChars    = "0123456789abcdef"
	sessExpiryTime = time.Hour * 168 // 1 week
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

	if Exists(postsKey, rdb) {
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
	}

	return user, nil
}

// FetchUserName fetches a user by its username.
func FetchUserName(name string, rdb *redis.Client) (*User, error) {
	if !UserExists(name, rdb) {
		return nil, errors.New("fetch user: user doesn't exist")
	}

	uidString, err := rdb.HGet("usernames", name).Result()
	if err != nil {
		return nil, err
	}

	uid, err := strconv.Atoi(uidString)
	if err != nil {
		return nil, err
	}

	return FetchUser(uid, rdb)
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

// NewSession makes a new session for a user with a random 32-character hexadecimal
// string as the session id. It returns the session id and any errors.
func (u *User) NewSession(rdb *redis.Client) (string, error) {
	var (
		key    = fmt.Sprintf("user:%d:sessions", u.ID)
		id     = GenerateSessionID()
		expiry = time.Now().Add(sessExpiryTime).Unix()

		member = redis.Z{
			Score:  float64(expiry),
			Member: id,
		}
	)

	if _, err := rdb.ZAdd(key, member).Result(); err != nil {
		return "", err
	}

	if _, err := rdb.HSet("sessions", id, u.ID).Result(); err != nil {
		return "", err
	}

	return id, nil
}

// GetSessions gets all current valid sessions of the user. It also purges the
// out-of-date ones.
func (u *User) GetSessions(rdb *redis.Client) ([]string, error) {
	var (
		key = fmt.Sprintf("user:%d:sessions", u.ID)
		now = time.Now().Unix()
	)

	sessions, err := rdb.ZRangeByScore(key, redis.ZRangeBy{
		Min: fmt.Sprintf("%v", now),
		Max: fmt.Sprintf("%v", math.Inf(1)),
	}).Result()

	if err != nil {
		return []string{}, err
	}

	return sessions, nil
}

// IsValidSession checks if the given session id is currently assigned to the user.
// Since it calls GetSessions, it purges out-of-date sessions from the database.
func (u *User) IsValidSession(id string, rdb *redis.Client) (bool, error) {
	if len(id) != 32 {
		return false, nil
	}

	sessions, err := u.GetSessions(rdb)
	if err != nil {
		return false, err
	}

	for _, sess := range sessions {
		if sess == id {
			return true, nil
		}
	}

	return false, nil
}

// GenerateSessionID generates a new session ID.
// TODO: Use crypto/rand instead of math/rand
func GenerateSessionID() string {
	var id bytes.Buffer

	for i := 0; i < sessIDLength; i++ {
		char := sessIDChars[rand.Intn(len(sessIDChars)-1)]
		id.WriteByte(char)
	}

	return id.String()
}

// GetUserFromSession gets the user identified by the session id.
func GetUserFromSession(id string, rdb *redis.Client) (*User, error) {
	uidString, err := rdb.HGet("sessions", id).Result()
	if err != nil {
		return nil, err
	}

	uid, err := strconv.Atoi(uidString)
	if err != nil {
		return nil, err
	}

	return FetchUser(uid, rdb)
}

// GetLoggedInUser gets the logged in user using the 'session' cookie.
func GetLoggedInUser(r *http.Request, rdb *redis.Client) (*User, error) {
	sess, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	return GetUserFromSession(sess.Value, rdb)
}
