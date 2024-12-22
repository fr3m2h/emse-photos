<h1 align="center">
  <br>
  EMSE-Photos
  <br>
</h1>

<h4 align="center">An easy-to-deploy, full-stack website for hosting your photos.</h4>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#prerequisites">Prerequisites</a> •
  <a href="#getting-started">Getting started</a> •
  <a href="#license">License</a>
</p>

## Features

* Authentication
	- [CAS login](https://en.wikipedia.org/wiki/Central_Authentication_Service)
	- Admin/student login distinction.

* Admin
	- Upload photos.
	- Create/modify events.
	- Add date and description to events.
	- Set photo name (using a format like IMG_date_id.png).
	- Option to include/exclude metadata.
	- Password protection for photo folders.
	- Tag photos by event or type.
	- Delete or hide photos.
	- Manage admin accounts.

* Student
	- Search photos by event or tag.
	- View compressed versions of photos.
	- Download full-size photos.
	- Create/manage personal photo folders.
	- Report a photo.

## Prerequisites

* [Download Git](https://git-scm.com/downloads) to clone the repository.

* [Download Go](https://go.dev/dl/) (version 1.23.1 or newer recommended) from the official Go website to compile, run and test this project.
* Download and install either [*MariaDB*](https://mariadb.org/download) or [*MySQL*](https://dev.mysql.com/downloads/mysql/) for the photos database.

## Getting started

You can find a quick guide inside ./docs/tutorials/tutorials.md to help you setup the environnement

```bash
# Clone this repository
$ git clone https://github.com/fr3m2h/photos

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

## License

GPLv3
