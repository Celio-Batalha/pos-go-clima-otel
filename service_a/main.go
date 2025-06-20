package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type (
	DTOInput struct {
		Cep string `json:"cep"`
	}
	DTOOutput struct {
		TempC  float64 `json:"temp_C"`
		TempF  float64 `json:"temp_F"`
		TempK  float64 `json:"temp_K"`
		Cidade string  `json:"cidade"`
	}
)

func main() {
	setTracing()
	http.HandleFunc("/", Handle)
	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var input DTOInput
	var output DTOOutput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !validCep(input.Cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx, span := otel.Tracer("service-a").Start(r.Context(), "1 - req-to-service-b")
	defer span.End()

	output, status, err := getInfo(input.Cep, ctx)
	if err != nil {
		log.Printf("Error fetching info: %v", err)
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

func setTracing() {
	exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
	if err != nil {
		log.Fatalf("Fail to create Zipkin exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-a"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func getInfo(cep string, ctx context.Context) (DTOOutput, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://goappb:8081/weather?cep="+strings.ReplaceAll(cep, "-", ""), nil)
	if err != nil {
		return DTOOutput{}, http.StatusInternalServerError, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return DTOOutput{}, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return DTOOutput{}, resp.StatusCode, errors.New("can not find zipcode")
	}

	var output DTOOutput
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return output, resp.StatusCode, err
	}

	return output, resp.StatusCode, nil
}

func validCep(cep string) bool {
	cep = strings.ReplaceAll(cep, "-", "") // Remove o hífen
	if len(cep) != 8 {
		return false
	}
	for _, c := range cep {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
