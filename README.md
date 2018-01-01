# Reddis

A Reddit clone using redis. Just cloning the repository and running it won't work,
since you need to start a redis server first. To do that, install redis (look it up)
then navigate to your cloned directory and run `$ redis-server`. This will start
a database, but it will be empty. At the very least, you need to add a key called
`posts` which you can initialise to an empty set.

As well as setting up the server, you'll need to create an SSL certificate, since
Reddis wants to run on `https`, not `http`. To enable this, run these commands:

```
openssl genrsa -out server.key 2048
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

This will generate two files, server.crt and server.key, in the directory. Now
you can `$ go run main.go` to start up the server. When you connect to it in the
browser, it will probably tell you the site can't be trusted, but will also give
you the option to go to it anyway. Do this once, and it shouldn't ask you again.

![](screenshot.png)
