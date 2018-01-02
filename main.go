package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Zac-Garby/reddis/lib"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
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

	r := mux.NewRouter()

	// Pages
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/register", registerHandler)
	r.HandleFunc("/me", meHandler)
	r.HandleFunc("/{name:~[^\\s]+}", userHandler)

	// API
	r.HandleFunc("/get_posts", getPostsHandler)
	r.HandleFunc("/user_exists", userExistsHandler)

	// Resources
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.Handle("/favicon.ico", http.FileServer(http.Dir(".")))

	fmt.Println("listening on https://localhost:3000")
	http.ListenAndServeTLS(":3000", "cert.pem", "key.pem", r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, ".") { // HACK: Change this at some point
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := lib.RenderRegister(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		var (
			name = r.PostFormValue("name")
			hash = r.PostFormValue("hash")
		)

		if len(name) == 0 {
			http.Error(w, "name must be at least 1 character. got "+name, http.StatusBadRequest)
			return
		}

		if len(hash) != 128 {
			http.Error(w, fmt.Sprintf("hash must be exactly 128 characters long. got %d", len(hash)), http.StatusBadRequest)
			return
		}

		user, err := lib.NewUser(name, hash, rdb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}

		sess, err := user.NewSession(rdb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := &http.Cookie{
			Name:   "session",
			Secure: true,
			Value:  sess,
		}

		// Set the session cookie
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		http.Error(w, "unsupported method: "+r.Method, http.StatusMethodNotAllowed)
	}
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	user, err := lib.GetLoggedInUser(r, rdb)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/~%s", user.Name), http.StatusFound)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"][1:]

	if !lib.UserExists(name, rdb) {
		// TODO: Add a better error message
		http.Error(w, "user doesn't exist", http.StatusNotFound)
		return
	}

	user, err := lib.FetchUserName(name, rdb)
	if err != nil {
		// TODO: Add a better error message
		http.Error(w, "could not fetch user", http.StatusInternalServerError)
		return
	}

	if err := lib.RenderUser(w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func userExistsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		name   = r.URL.Query().Get("name")
		exists = lib.UserExists(name, rdb)
	)

	fmt.Fprintf(w, "%t", exists)
}
