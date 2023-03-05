#!/bin/bash

#go run ./... -mysql test:test@/test -w 5 -dict small.txt -out out.txt
#go run ./... -mysql test:test@/test -w 10 -dict dutch.txt -out out.txt
go run ./... -mysql "test:test@tcp(127.0.0.1:3307)/test" -w 10 -dict rockyou.txt -out out.txt
