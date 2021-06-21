# Go API Demo

This repo contains code for an API written in Go.

## Overview

To run the API you'll need to have Go installed.

### gRPC
Install a _Protocol Buffer Compiler_ and the
_Go Plugins_ for the compiler (see the
[gRPC Quickstart](https://grpc.io/docs/languages/go/quickstart/) for details) if you 
want to:
* compile the _.proto_ files by running `make proto` and/or
* manually issue gRPC requests using [gRPCurl](https://github.com/fullstorydev/grpcurl).

### Make

The _Makefile_ contains commands for building, running and testing the API.

* `make run` builds and runs the binary.
* `make build` just builds the binary.
* `make fmt` formats the code, updates import lines and runs clean.
* `make lint` runs golangci-lint.
* `make proto` generates protobuf files from proto files.
* `make test` runs the linter then the tests (see [Tests](#tests)).

### <a name="tests"></a>Tests

Install [testify](https://github.com/stretchr/testify#installation) then run `make test`.

### Manual Testing

You can test the API manually using a client. For instance 
[Insomnia](https://insomnia.rest/download)
supports both HTTP and [gRPC](https://support.insomnia.rest/article/188-grpc#overview). 

Alternatively, requests can be issued using cURL and
[gRPCurl](https://github.com/fullstorydev/grpcurl).

## v0.1.0

Basic HTTP and gRPC server.

### HTTP
#### Request

    curl -i localhost:3000
 
##### Response

    HTTP/1.1 200 OK
    Date: Tue, 22 Jun 2021 11:19:48 GMT
    Content-Length: 0

### gRPC

You'll need to generate a protoset and have 
[gRPCurl](https://github.com/fullstorydev/grpcurl) installed.

#### Generate protoset

    protoc \
    -I=proto \
    --descriptor_set_out=generated/hello.protoset \
    hello.proto

#### Request

    grpcurl \
    -plaintext \
    -protoset generated/hello.protoset \
    -d '{"name": "world"}' \
    localhost:1234 Hello/Hello

#### Response

    {
      "message": "Hello world"
    }
