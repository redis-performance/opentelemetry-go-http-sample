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

	// Tracers can be accessed with a TracerProvider.
	// In implementations of the API, the TracerProvider is expected to be the stateful
	// object that holds any configuration.
	// Normally, the TracerProvider is expected to be accessed from a central place.
	// Thus, the API SHOULD provide a way to set/register and access a global default TracerProvider.
	//
	// See: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#tracerprovider
	otel.SetTracerProvider(tp)

	// TextMapPropagator performs the injection and extraction of a cross-cutting concern value as string key/values
	// pairs into carriers that travel in-band across process boundaries.
	// The carrier of propagated data on both the client (injector) and server (extractor) side is usually an HTTP request.
	// In order to increase compatibility, the key/value pairs MUST only consist of US-ASCII characters that make up
	// valid HTTP header fields as per RFC 7230.
	//
	// See: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#textmap-propagator
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
