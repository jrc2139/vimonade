#!/bin/bash

# go
/usr/local/bin/protoc message.proto --proto_path=protos/message \
    --proto_path=. \
    --go_out=plugins=grpc:pkg/api/v1
