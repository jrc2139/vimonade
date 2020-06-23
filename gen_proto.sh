#!/bin/bash

# go
/usr/local/bin/protoc vimonade.proto --proto_path=protos/vimonade \
    --proto_path=. \
    --go_out=plugins=grpc:api
