#!/bin/bash
docker run --rm --detach \
  --name kdb \
  --env MARIADB_DATABASE=test \
  --env MARIADB_USER=test \
  --env MARIADB_PASSWORD=test \
  --env MARIADB_ROOT_PASSWORD=test \
  -v $(pwd)/db/k:/docker-entrypoint-initdb.d \
  -v kdb:/var/lib/mysql \
  -p "3307:3306" \
mariadb:latest
