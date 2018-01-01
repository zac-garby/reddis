package lib

import "github.com/go-redis/redis"

// Exists determines whether or not a key exists in the databse. It ignores
// any errors to make it easier to use, and since any error probably means the
// key doesn't exist.
func Exists(key string, rdb *redis.Client) bool {
	exists, err := rdb.Exists(key).Result()
	if err != nil {
		return false
	}

	return exists == 1
}
