package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Zac-Garby/reddis/lib"
	"github.com/go-redis/redis"
)

const (
	normalDepth = 4
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
	http.HandleFunc("/get_posts", getPostsHandler)

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

func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")

	if len(idParam) == 0 {
		fmt.Fprintf(w, "<pre>Could not load posts</pre>")
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		fmt.Fprintf(w, "<pre>Could not load posts - invalid post ID</pre>")
		return
	}

	tree, err := lib.FetchPostTree(id, -1, rdb)
	if err != nil {
		fmt.Fprintf(w, "<pre>Could not load posts</pre>")
		return
	}

	if err := lib.RenderPosts(w, tree); err != nil {
		fmt.Fprintf(w, "<pre>Could not render posts</pre>")
	}
}
