#!/usr/bin/env bash

docker build --output type=local,dest=. --file Dockerfile.generate .

ls -l testproto/testproto.pb.go
