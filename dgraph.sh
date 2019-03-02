#!/bin/bash

# Run Dgraph Zero
docker run -d -p 5080:5080 -p 6080:6080 -p 8080:8080 -p 9080:9080 -p 8000:8000 -v ~/go/src/github.com/xn3cr0nx/bitgodine_code/dgraph:/dgraph --name diggy dgraph/dgraph dgraph zero

# Run Dgraph Alpha
docker exec -d diggy dgraph alpha --lru_mb 2048 --zero localhost:5080

# Run Dgraph Ratel
docker exec -d diggy dgraph-ratel