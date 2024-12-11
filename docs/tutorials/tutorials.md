# Tutorials
To clone and run this application, you'll need [Git](https://git-scm.com) and [Go](https://go.dev/) installed on your computer.

To simulate a cas server on your machine, you'll find a basic implementation inside ./cmd/cas_server/launch_server.go that you can run.

On first use, the program will create the correct config file inside your working directory, default setting are fine and hex-encoded secrets used
to authenticate csrf and session cookies are generated using a cryptographically secure pseudorandom number generator. Feel free to change them:
it must be a correct hex-encoded value.

Regarding the databse, you'll have to setup a MySQL or MariaDB database, copy paste the schema inside the file schema.sql and then fill the database
DSN inside the config file.
