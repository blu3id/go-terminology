package server

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/soheilhy/cmux"
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func serveGRPC(l net.Listener, sct *terminology.Svc) {
	// Register gRPC SNOMED Service
	var gRPCOpts []grpc.ServerOption
	gRPCServer := grpc.NewServer(gRPCOpts...)
	snomed.RegisterSnomedCTServer(gRPCServer, &snomedCTSrv{svc: sct})

	if err := gRPCServer.Serve(l); err != nil {
		log.Fatalf("Error while serving gRPC: %v", err)
	}
	return
}

func serveGRPCGateway(l net.Listener, host string) {
	// Register gRPC Gateway HTTP Reverse Proxy
	ctx := context.Background()
	gRPCGateway := runtime.NewServeMux()
	gRPCGatewayOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := snomed.RegisterSnomedCTHandlerFromEndpoint(ctx, gRPCGateway, host, gRPCGatewayOpts)
	if err != nil {
		log.Fatalf("Error Registering gRPC Gateway Handler: %v", err)
		return
	}

	if err := http.Serve(l, gRPCGateway); err != nil {
		log.Fatalf("Error while serving HTTP: %v", err)
	}
	return
}

// Serve starts listening on host (interface:port) for either gRPC requests or
// HTTP request and passess connections onto appropriate handler. HTTP requests
// are passed through gRPC Gateway HTTP Reverse Proxy and back to gRPC server.
func Serve(sct *terminology.Svc, host string) {
	// Create a listener at the desired port.
	l, err := net.Listen("tcp", host)
	defer l.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Create a cmux object.
	tcpm := cmux.New(l)

	// Declare the match for different services required.
	grpc := tcpm.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	http := tcpm.Match(cmux.HTTP1Fast())

	// Initialize the servers by passing in the custom listeners (sub-listeners).
	go serveGRPC(grpc, sct)
	go serveGRPCGateway(http, host)
	log.Println("gRPC and HTTP server listening on", host)

	// Start cmux serving.
	if err := tcpm.Serve(); !strings.Contains(err.Error(),
		"use of closed network connection") {
		log.Fatal(err)
	}
}

/*
// Only works if using TLS with gRPC

// Serve starts listening on host (interface:port) for either gRPC requests or
// HTTP request and passess connections onto appropriate handler. HTTP requests
// are passed through gRPC Gateway HTTP Reverse Proxy and back to gRPC server.
func Serve(sct *terminology.Svc, host string) error {
	// Register gRPC SNOMED Service
	var gRPCOpts []grpc.ServerOption
	gRPCServer := grpc.NewServer(gRPCOpts...)
	snomed.RegisterSnomedCTServer(gRPCServer, &myServer{svc: sct})

	// Register gRPC Gateway HTTP Reverse Proxy
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	gRPCGateway := runtime.NewServeMux()
	gRPCGatewayOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := snomed.RegisterSnomedCTHandlerFromEndpoint(ctx, gRPCGateway, host, gRPCGatewayOpts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(host, grpcHandlerFunc(gRPCServer, gRPCGateway))
}
*/
