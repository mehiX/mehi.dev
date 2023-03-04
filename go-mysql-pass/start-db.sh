#!/bin/bash
docker run --rm --detach \
  --name founders \
  --env MARIADB_DATABASE=test \
  --env MARIADB_USER=test \
  --env MARIADB_PASSWORD=test \
  --env MARIADB_ROOT_PASSWORD=test \
  -v $(pwd)/db/founders:/docker-entrypoint-initdb.d \
  -v founders:/var/lib/mysql \
  -p "3306:3306" \
mariadb:latest
