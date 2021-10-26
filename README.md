# porker-backend

This is a backend application implemented in Go language.  
It works by communicating with frontend application implemented in the Flutter.

## Related repositories

- https://github.com/swallowarc/porker-proto
- https://github.com/swallowarc/porker-front

## Getting Started

Follow the steps below to get started.
(Run on localhost)

### 1. Install Go language

Please set up Golang version 1.16 or higher in advance.

### 2. Setup tools

Run the following command to set up the required tools.

```shell
make setup/tools
```

### 3. Docker compose

Please execute the following command after installing Docker.  
The docker container needed to run the program will start.

```shell
make setup/service
```

### 5. Backend application launch

Use the following command.

```shell
go run ./cmd/porker-rpc/
```

Or use an IDE (Intellij IDEA, Visual Studio Code, etc.) to start debugging.  
We strongly recommend using the IDE from the perspective of development efficiency.

## Other steps

### Create mocks

You can generate a mock with the following command.

```shell
make mock/gen
```
