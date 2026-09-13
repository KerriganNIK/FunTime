package main

import (
	"context"
	roomv1 "funtime.local/contracts/room/v1"
	runtimev1 "funtime.local/contracts/runtime/v1"
	"funtime.local/platform/server"
	"funtime.local/platform/store"
	"funtime.local/room/internal/room"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	s, err := store.Open(context.Background(), server.Env("DATABASE_URL", ""), server.Env("DATA_DIR", ".cache/data/room"))
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	conn, err := grpc.NewClient(server.Env("WORLD_DOMINATION_ADDR", "127.0.0.1:9005"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	service := room.New(s, map[string]runtimev1.GameRuntimeClient{"world-domination": runtimev1.NewGameRuntimeClient(conn)})
	if err = server.Serve("room", server.Env("GRPC_ADDR", ":9003"), server.Env("HEALTH_ADDR", ":9004"), func(g *grpc.Server) { roomv1.RegisterRoomServiceServer(g, service) }, nil); err != nil {
		log.Fatal(err)
	}
}
