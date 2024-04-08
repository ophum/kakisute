#!/bin/bash

echo "delete"
rm -f cgi-bin/server.cgi
echo "build"
go build -o ./cgi-bin/server.cgi ./cmd/server-cgi/

echo "run server"
python3 -m http.server --cgi 8080
