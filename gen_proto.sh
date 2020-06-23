#!/bin/bash

# go
/usr/local/bin/protoc vimonade.proto --proto_path=protos/vimonade \
    --go_out=plugins=grpc:api
