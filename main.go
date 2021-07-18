package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/model"
	reporterhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"log"
	"net/http"
	"time"
)

const endpointURL = "http://localhost:9411/api/v2/spans"

/*
https://github.com/openzipkin/zipkin-go/blob/master/examples/httpserver_test.go
https://medium.com/devthoughts/instrumenting-a-go-application-with-zipkin-b79cc858ac3e
*/
func main() {
	tracer, err := newTracer()
	if err != nil {
		log.Fatal(err)
	}
	// We add the instrumented transport to the defaultClient
	// that comes with the zipkin-go library
	client, err := zipkinhttp.NewClient(tracer, zipkinhttp.ClientTrace(true))
	if err != nil {
		log.Fatalf("unable to create client: %+v\n", err)
	}
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandlerFactory(client))
	r.HandleFunc("/foo", FooHandler)
	r.Use(zipkinhttp.NewServerMiddleware(
		tracer,
		zipkinhttp.SpanName("request")), // name for request span
	)
	log.Fatal(http.ListenAndServe(":8081", r))
}

func FooHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newTracer() (*zipkin.Tracer, error) {
	// The reporter sends traces to zipkin server
	reporter := reporterhttp.NewReporter(endpointURL)

	// Local endpoint represent the local service information
	localEndpoint := &model.Endpoint{ServiceName: "ping pong service", Port: 8081}

	// Sampler tells you which traces are going to be sampled or not. In this case we will record 100% (1.00) of traces.
	sampler, err := zipkin.NewCountingSampler(1)
	if err != nil {
		return nil, err
	}

	t, err := zipkin.NewTracer(
		reporter,
		zipkin.WithSampler(sampler),
		zipkin.WithLocalEndpoint(localEndpoint),
	)
	if err != nil {
		return nil, err
	}

	return t, err
}

func HomeHandlerFactory(client *zipkinhttp.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		span := zipkin.SpanFromContext(r.Context())
		span.Tag("custom_key", "some value")
		span.Annotate(time.Now(), "expensive_calc_done")
		newRequest, err := http.NewRequest("POST", "http://example.com", nil)
		if err != nil {
			log.Printf("unable to create client: %+v\n", err)
			http.Error(w, err.Error(), 500)
			return
		}
		ctx := zipkin.NewContext(newRequest.Context(), span)
		newRequest = newRequest.WithContext(ctx)
		res, err := client.DoWithAppSpan(newRequest, "other_function")
		if err != nil {
			log.Printf("call to other_function returned error: %+v\n", err)
			http.Error(w, err.Error(), 500)

			return
		}
		res.Body.Close()
		w.WriteHeader(http.StatusOK)
	}
}

// ZipkinHTTPRoute sets http.route if a span and gorilla mux
// template path are found.
func ZipkinHTTPRoute(ctx context.Context, r *http.Request) context.Context {
	if span := zipkin.SpanFromContext(ctx); span != nil {
		if route := mux.CurrentRoute(r); route != nil {
			if routePath, err := route.GetPathTemplate(); err == nil {
				zipkin.TagHTTPRoute.Set(span, routePath)
				span.SetName(r.Method + " " + routePath)
			}
		}
	}
	return ctx
}
