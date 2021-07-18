# Go API Demo

This repo contains code for an API written in Go.

## Overview

To run the API you'll need to have Go installed.

### gRPC
Install a _Protocol Buffer Compiler_, and the
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

## v0.3.0

Stores created users either in-memory or in MySQL.

To use MySQL storage you'll need to install 
[golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) and
run the following commands:

    make docker-up
    make migrate-up

Running `make docker-up` will
* copy `.env.dist` => `.env`
  * `USER_STORAGE` (either _memory_ or _sql_) determines whether users are stored 
    in-memory or in MySQL.
* start a docker-based instance of MySQL.

Running `make migrate-up` creates the table in MySQL for storing users.

The same cURL and gRPCurl requests as described for [v0.2.0](#v0.2.0) can be used.

## <a name="v0.2.0"></a>v0.2.0

Adding HTTP and gRPC endpoints for user creation.

Users are stored in-memory.

### HTTP
#### Request

    curl -i --request POST \
    --url http://localhost:3000/user \
    --header 'Content-Type: application/json' \
    --data '{
        "first_name": "john",
        "last_name": "smith"
    }'

##### Response

    HTTP/1.1 201 Created
    Content-Type: application/json
    Date: Tue, 06 Jul 2021 12:03:25 GMT
    Content-Length: 127

    {
        "id":"afaa2920-77e4-49d0-a29f-5f2d9b6bf2d1",
        "first_name":"john",
        "last_name":"smith",
        "created_at":"2021-07-06T13:03:25+01:00"
    }

### gRPC

You'll need to generate a protoset and have 
[gRPCurl](https://github.com/fullstorydev/grpcurl) installed.

#### Generate protoset

    protoc \
    -I=proto \
    --descriptor_set_out=generated/user.protoset \
    user.proto

#### Request

    grpcurl \
    -plaintext \
    -protoset generated/user.protoset \
    -d '{"first_name": "john", "last_name": "smith"}' \
    localhost:1234 User/Create

#### Response

    {
        "id": "ca3d9549-eb8d-4b5a-b45f-3551fb4fbdc9",
        "firstName": "john",
        "lastName": "smith",
        "createdAt": "2021-07-06T13:08:45+01:00"
    }

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
