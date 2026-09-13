package main

import (
	"context"
	runtimev1 "funtime.local/contracts/runtime/v1"
	"funtime.local/platform/server"
	"funtime.local/platform/store"
	"funtime.local/worlddomination/internal/runtime"
	"google.golang.org/grpc"
	"log"
)

func main() {
	s, err := store.Open(context.Background(), server.Env("DATABASE_URL", ""), server.Env("DATA_DIR", ".cache/data/world-domination"))
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	service, err := runtime.New(context.Background(), s)
	if err != nil {
		log.Fatal(err)
	}
	if err = server.Serve("world-domination", server.Env("GRPC_ADDR", ":9005"), server.Env("HEALTH_ADDR", ":9006"), func(g *grpc.Server) { runtimev1.RegisterGameRuntimeServer(g, service) }, service.Tick); err != nil {
		log.Fatal(err)
	}
}
