package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	pb "github.com/injae/go_grpc_study/proto/proto"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("port", 8080, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	//log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func allHandler(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request match: %+v", r)
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func main() {
	flag.Parse()
	addr := fmt.Sprintf(":%d", *port)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor),
		),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor),
		),
	)

	pb.RegisterGreeterServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	gwmux := runtime.NewServeMux()
	err := pb.RegisterGreeterHandlerFromEndpoint(
		context.Background(),
		gwmux,
		addr,
		[]grpc.DialOption{grpc.WithInsecure()},
	)
	if err != nil {
		log.Fatal("Failed to register server:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "proto/openapiv2/proto/hello.swagger.json")
	})
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("swagger-ui/dist"))))
	mux.Handle("/", gwmux)
	mux.Handle("/metrics", promhttp.Handler())
	grpc_prometheus.Register(grpcServer)

	log.Printf("grpc server listen: %v", addr)

	err = http.ListenAndServe(addr, allHandler(grpcServer, mux))
	if err != nil {
		log.Fatal("Unable to start a http server.", err.Error())
	}
}
