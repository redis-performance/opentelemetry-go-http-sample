package main

import (
	"fmt"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
	"os"
)

var (
	jaegerEndpoint string
	redisEndpoint  string
	address        string
)

type server struct {
	client *redis.Client
	router *mux.Router
}

func initTracer(serviceName string, jaegerAgentEndpoint string) {

	// Create and install Jaeger export pipeline
	tp, _, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(fmt.Sprintf("%s/api/traces", jaegerAgentEndpoint)),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func main() {
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	flag.StringVar(&jaegerEndpoint, "jaegerAgentEndpoint", "http://localhost:14268", "Jaeger agent endpoint")
	flag.StringVar(&redisEndpoint, "redisEndpoint", "localhost:6379", "Redis endpoint")
	flag.StringVar(&address, "address", ":7777", "Serving address")
	flag.Parse()
	logger := log.New(os.Stdout, "sampleHTTPServer: ", log.LstdFlags)
	logger.Println("Starting tracer...")
	initTracer("sampleHTTPServer", jaegerEndpoint)
	logger.Println("Connecting to redis...")
	redisClient := initRedis(redisEndpoint)

	logger.Println("Starting server...")
	sampleServer := initService(redisClient)

	err := http.ListenAndServe(address, sampleServer.router)
	if err != nil {
		panic(err)
	}
}

func initRedis(address string) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: address,
	})
	redisClient.AddHook(redisotel.TracingHook{})
	return redisClient
}

func initService(redisClient *redis.Client) server {
	sampleServer := server{
		client: redisClient,
		router: mux.NewRouter(),
	}
	sampleServer.routes()
	return sampleServer
}
