# Tutorials
To clone and run this application, you'll need [Git](https://git-scm.com) and [Go](https://go.dev/) installed on your computer.

To simulate a cas server on your machine, you'll find a basic implementation inside ./cmd/cas_server/launch_server.go that you can run.

On first use, the program will create the correct config file inside your working directory, default setting are fine and hex-encoded secrets used
to authenticate csrf and session cookies are generated using a cryptographically secure pseudorandom number generator. Feel free to change them:
it must be a correct hex-encoded value.

Regarding the database, you will have to setup a MySQL or MariaDB database, copy paste the schema inside the file schema.sql and then fill the database
DSN inside the config file.


```bash
# Clone this repository
$ git clone https://github.com/fr3m2h/emse-photos

# Go into the repository
$ cd photos

# Install dependencies
$ go mod tidy

# [Run|Build] the photos or the mock cas server
$ go [run|build] -o bin/launch_photo_server ./cmd/photos_server/launch_server.go
$ go [run|build] -o bin/launch_mock_cas_server ./cmd/cas_server/launch_server.go

# Test the app
go test -cover ./...
```

