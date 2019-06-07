#!/bin/bash

running="$(docker ps -a | grep diggy)"

if [[ "$running" ]]; then
    echo "Restarting diggy"
    docker start diggy
else
    # Run Dgraph Zero
    echo "Starting dgraph named diggy"
    docker run -d -p 5080:5080 -p 6080:6080 -p 8080:8080 -p 9080:9080 -p 8000:8000 -v ~/.bitgodine/dgraph:/dgraph --name diggy dgraph/dgraph dgraph zero
    # Run Dgraph Alpha
    docker exec -d diggy dgraph alpha --lru_mb 12288 --zero localhost:5080
    # Run Dgraph Ratel
    docker exec -d diggy dgraph-ratel
fi


