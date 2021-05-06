package main

import "context"

func main() {
	grpcServer := setup()
	grpcServer.RunGRPCServer(context.Background())
}
