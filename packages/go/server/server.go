package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func Env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Serve(name, grpcAddr, healthAddr string, register func(*grpc.Server), tick func(context.Context)) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	srv := grpc.NewServer(grpc.MaxRecvMsgSize(1024 * 1024))
	register(srv)
	h := health.NewServer()
	h.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)
	healthv1.RegisterHealthServer(srv, h)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	probe := &http.Server{Addr: healthAddr, Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	failures := make(chan error, 2)
	go func() { failures <- srv.Serve(listener) }()
	go func() { failures <- probe.ListenAndServe() }()
	if tick != nil {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					tick(ctx)
				}
			}
		}()
	}
	slog.Info(name+" listening", "grpc", grpcAddr)
	select {
	case <-ctx.Done():
	case err = <-failures:
	}
	cancel()
	h.Shutdown()
	done := make(chan struct{})
	go func() { srv.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		srv.Stop()
	}
	shutdown, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	_ = probe.Shutdown(shutdown)
	if errors.Is(err, http.ErrServerClosed) || errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}
