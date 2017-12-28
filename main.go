package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Zac-Garby/social/lib"
	"github.com/go-redis/redis"
)

const (
	normalDepth = 8
)

var (
	rdb   *redis.Client
	chttp = http.NewServeMux()
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer rdb.Close()

	// Check for a connection
	if _, err := rdb.Ping().Result(); err != nil {
		panic(err)
	}

	http.HandleFunc("/", indexHandler)
	chttp.Handle("/", http.FileServer(http.Dir("./static/")))

	fmt.Println("listening...")
	http.ListenAndServe(":3000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, ".") {
		chttp.ServeHTTP(w, r)
	} else {
		tree, err := lib.FetchPostTree(0, normalDepth, rdb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := lib.RenderIndex(w, tree); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
