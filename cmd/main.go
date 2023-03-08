package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	api "dedb"
	"dedb/internal"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	//ctx := context.Background()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	ctx := context.Background()

	// Setup signal interuption for graceful shutdowns
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	// Setup the service and grpc server
	log := log.With().Str("logger", "dedb").Logger()

	g, ctx := errgroup.WithContext(ctx)
	grpcServer := createGrpcServer()

	service := &internal.Service{}
	g.Go(func() error {
		config := internal.Config{}
		err := envconfig.Process("", &config)
		if err != nil {
			log.Error().Err(err).Msg("failed to process config")
		}

		log.Info().Msg("registering services with GRPC server")
		api.RegisterDeDBServer(grpcServer, service)

		log.Info().Msgf("starting service at: %s", config.ServiceGrpcPort)
		err = service.Start(config)
		if err != nil {
			log.Error().Err(err).Msg("could not start service")
			return err
		}

		lis, err := net.Listen("tcp", config.ServiceGrpcPort)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("failed to listen: %v", err))
			return err
		}

		err = grpcServer.Serve(lis)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("failed to serve: %v", err))
		}

		return err
	})

	// Handle shutdown from a signals
	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	log.Warn().Msg("received shutdown signal")
	service.Shutdown()
	_, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	/*
		if httpServer != nil {
			_ = httpServer.Shutdown(shutdownCtx)
		}
	*/
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	// Wait on the goroutines to run
	if err := g.Wait(); err != nil {
		log.Error().Err(err).Msg("Could not start servers")
		os.Exit(2)
	}
}

func createGrpcServer(interceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				interceptors...,
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(),
		),
	)
	return s
}
