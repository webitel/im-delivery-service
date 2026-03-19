#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PORT=9090
FILE="$SCRIPT_DIR/hook_server.go" 

if [ ! -f "$FILE" ]; then
    echo "Error: $FILE not found!"
    exit 1
fi

echo "Starting Push Debug Server..."
echo "Target URL: http://localhost:$PORT"
echo "---------------------------------"

go run "$FILE"