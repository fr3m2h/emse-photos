golangci-lint run &&
go build -o bin/launch_photo_server ./cmd/photos_server/launch_server.go
go build -o bin/launch_mock_cas_server ./cmd/cas_server/launch_server.go
