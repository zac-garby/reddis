# Reddis

A Reddit clone using redis. Just cloning the repository and running it won't work,
since you need to start a redis server first. To do that, install redis (look it up)
then navigate to your cloned directory and run `$ redis-server`. This will start
a database, but it will be empty. At the very least, you need to add a key called
`posts` which you can initialise to an empty set. I think then you can just run
`$ go run main.go` and open `localhost:3000` in your web-browser and it _should_
start. Obviously, there won't be any posts displayed because you haven't created
any yet. There will be a mechanism for writing posts via the actual app sometime
in the future, but before that, I'll have to implement users and logging in.

![](screenshot.png)
