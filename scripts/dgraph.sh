#!/bin/bash

running="$(docker ps -a | grep diggy)"

if [[ running != "" ]]; then
    echo "Restarting diggy"
    docker start diggy
else
    # Run Dgraph Zero
    echo "Starting dgraph named diggy"
    docker run -d -p 5080:5080 -p 6080:6080 -p 8080:8080 -p 9080:9080 -p 8000:8000 -v ~/go/src/github.com/xn3cr0nx/bitgodine_code/dgraph:/dgraph --name diggy dgraph/dgraph dgraph zero
fi

# Run Dgraph Alpha
docker exec -d diggy dgraph alpha --lru_mb 4096 --zero localhost:5080

# Run Dgraph Ratel
docker exec -d diggy dgraph-ratel