package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const api_endpoint = "https://api.pokemontcg.io/v2/cards"

type Card struct {
	ID     string   `json:"ID"`
	Name   string   `json:"Name"`
	Type   []string `json:"Types"`
	HP     string   `json:"HP"`
	Rarity string   `json:"Rarity"`
}

type CardList struct {
	Cards []Card `json:"data"`
}

func main() {
	var (
		limitFlag = flag.Int("limit", 10, "the number of cards to retrieve")
		logFile   = flag.String("log", "./logs/pokecards.log", "the file to write logs to")
	)
	flag.Parse()
	// Create log directory
	err := os.MkdirAll("./logs", os.ModePerm)
	errHandler(err, "Failed to create log directory")

	// Open file handler for logging
	logWriter, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	errHandler(err, "Failed to open log file")
	defer logWriter.Close()

	// Define log writers
	infoLogger := log.New(io.MultiWriter(logWriter), "INFO ", log.LstdFlags)
	errorLogger := log.New(io.MultiWriter(logWriter), "ERROR ", log.LstdFlags)

	// Set up the OTel exporter
	ctx, tp := initTraceProvider()

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			errorLogger.Printf("Failed to shut down tracer provider: %v", err)
		}
	}()

	// Create an HTTP client with OpenTelemetry tracing
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	// Start the main span
	tracer := otel.Tracer("pokecards")
	ctx, span := tracer.Start(ctx, "drawCards")
	defer span.End()

	// Draw cards
	cardData, httpStatusCode, err := drawCards(ctx, client, *limitFlag, tracer)
	logError(errorLogger, err, "Failed to make cards appear")

	// Print the cards and log the result or log the error.
	jsonData, err := json.Marshal(struct{ Cards []Card }{cardData.Cards})
	logError(errorLogger, err, "Failed to marshal JSON: %v")
	logInfo(infoLogger, fmt.Sprintf("HTTP/%v A wild %d cards appeared!\n", httpStatusCode, len(cardData.Cards)))
	fmt.Println(string(jsonData))
}

func drawCards(ctx context.Context, client *http.Client, limit int, tracer trace.Tracer) (*CardList, string, error) {
	// Create new request
	req, err := http.NewRequestWithContext(ctx, "GET", api_endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	query := req.URL.Query()
	query.Add("pageSize", strconv.Itoa(limit))
	query.Add("orderBy", "id")
	query.Add("q", "(types:fire OR types:grass) -types:metal rarity:rare -rarity:holo -rarity:promo")
	req.URL.RawQuery = query.Encode()

	// Create a child span to trace the HTTP request
	ctx, span := tracer.Start(ctx, "drawCardsHTTP")
	defer span.End()

	// Perform the request
	resp, err := client.Do(req)
	httpStatusCode := resp.StatusCode
	if err != nil {
		return nil, strconv.Itoa(httpStatusCode), fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Deserialize the response
	startTime := time.Now()
	var cardList CardList
	err = json.NewDecoder(resp.Body).Decode(&cardList)
	elapsed := time.Since(startTime).Milliseconds()
	if err != nil {
		return nil, strconv.Itoa(httpStatusCode), fmt.Errorf("failed to decode JSON: %w", err)
	}

	span.SetAttributes(
		attribute.Int("http.status_code", httpStatusCode),
		attribute.Int64("json.unmarshal_time_ms", elapsed),
	)
	if httpStatusCode != http.StatusOK {
		span.SetAttributes(attribute.String("error", "non-OK HTTP status code"))
		return &cardList, strconv.Itoa(httpStatusCode), fmt.Errorf("HTTP/%v", httpStatusCode)
	}

	return &cardList, strconv.Itoa(httpStatusCode), nil
}

// Error handler
func errHandler(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

// Error logging
func logError(logger *log.Logger, err error, message string) {
	if err != nil {
		logger.Fatalf("%s: %v", message, err)
	}
}

// Info logging
func logInfo(logger *log.Logger, message string) {
	logger.Printf("%s", message)
}

// Initialize the Trace Provider
func initTraceProvider() (context.Context, *sdktrace.TracerProvider) {
	ctx := context.Background()
	httpClient := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("localhost:4317"),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithTimeout(5*time.Millisecond),
	)
	exporter, err := otlptrace.New(ctx, httpClient)
	errHandler(err, "Failed to create OTLP exporter")

	return ctx, sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	)
}
