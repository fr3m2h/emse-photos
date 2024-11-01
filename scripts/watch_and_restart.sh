#!/bin/bash

# Path to the binary and the directory to monitor
BINARY_PATH="/home/martin/documents/workspace/github.com/martinfemt/go/photos/bin/launch_server"
DIRECTORY_TO_WATCH="/home/martin/documents/workspace/github.com/martinfemt/go/photos/"

# Function to start the server
start_server() {
    echo "Starting the server..."
    # Start the binary server in the background and save its PID
    "$BINARY_PATH" &
    SERVER_PID=$!
    echo "Server started with PID: $SERVER_PID"
}

# Function to stop the server
stop_server() {
    if [ -n "$SERVER_PID" ]; then
        echo "Stopping the server with PID: $SERVER_PID..."
        kill "$SERVER_PID"
        wait "$SERVER_PID" 2>/dev/null
        echo "Server stopped."
    fi
}

# Trap EXIT signal to ensure the server is stopped when the script exits
trap stop_server EXIT

# Start the server
start_server

# Monitor the directory for changes and restart the server as needed
inotifywait -m -r -e modify,move,create,delete "$DIRECTORY_TO_WATCH" |
while read -r path event file; do
    echo "Change detected in $path$file: $event"
    # Restart the server
    stop_server
    start_server
done
