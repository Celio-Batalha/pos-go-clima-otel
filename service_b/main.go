package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type WeatherResponse struct {
	TempC  float64 `json:"temp_c"`
	TempF  float64 `json:"temp_f"`
	TempK  float64 `json:"temp_k"`
	Cidade string
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro"` // Indica se houve erro na consulta do CEP.
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
	} `json:"current"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

const viaCEPURL = "https://viacep.com.br/ws/%s/json/"             // URL da API ViaCEP para buscar dados do CEP.
const weatherAPIURL = "http://api.weatherapi.com/v1/current.json" // URL da API WeatherAPI para buscar dados climáticos.
var weatherKey string

func main() {

	weatherKey = os.Getenv("WEATHER_KEY")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	setTracing()

	http.HandleFunc("/weather", Handle)
	fmt.Println("Servidor rodando na porta 8081...")
	http.ListenAndServe(":"+port, nil)

}

func Handle(w http.ResponseWriter, r *http.Request) {

	cepParam := r.URL.Query().Get("cep")
	if cepParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx, span := otel.Tracer("service-b").Start(r.Context(), "2 - service-b-start")
	defer span.End()

	if !validarCEP(cepParam) {
		// json.NewEncoder(w).Encode(ErrorResponse{Message: "CEP inválido"})
		http.Error(w, "CEP inválido", http.StatusUnprocessableEntity)
		return
	}

	localizacao, err := buscarLocalizacao(ctx, cepParam)
	if err != nil {
		http.Error(w, "Erro ao buscar localização", http.StatusNotFound)
		// json.NewEncoder(w).Encode(ErrorResponse{Message: "Erro ao buscar localização"})
		return
	}
	if localizacao.Erro {
		http.Error(w, "CEP Nao encontrado!", http.StatusNotFound)
		return
	}
	clima, err := buscarClimaAtual(ctx, localizacao.Localidade)
	if err != nil {
		http.Error(w, "Erro ao buscar clima atuall", http.StatusInternalServerError)
		return
	}

	tempC := clima.Current.TempC
	tempF := clima.Current.TempF
	tempK := tempC + 273.15

	// t.Execute(w, WeatherResponse{
	// 	TempC:  tempC,
	// 	TempF:  tempF,
	// 	TempK:  tempK,
	// 	Cidade: localizacao.Localidade,
	// })

	response := WeatherResponse{
		TempC:  tempC,
		TempF:  tempF,
		TempK:  tempK,
		Cidade: localizacao.Localidade,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Garante o código 200
	json.NewEncoder(w).Encode(response)
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
			semconv.ServiceNameKey.String("service-b"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func buscarLocalizacao(ctx context.Context, cep string) (ViaCEPResponse, error) {
	_, span := otel.Tracer("service-b").Start(ctx, "3 - service-b-get-location")
	defer span.End()

	url := fmt.Sprintf(viaCEPURL, cep)
	resp, err := http.Get(url)
	if err != nil {
		return ViaCEPResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ViaCEPResponse{}, fmt.Errorf("erro ao buscar localização: %v", resp.StatusCode)
	}
	var viaCEP ViaCEPResponse
	err = json.NewDecoder(resp.Body).Decode(&viaCEP)
	return viaCEP, err
}

func buscarClimaAtual(ctx context.Context, localidade string) (WeatherAPIResponse, error) {
	_, span := otel.Tracer("service-b").Start(ctx, "4 - service-b-get-weather")
	defer span.End()

	var apiKey = "1dc5342330fc472795c31126251706"
	fmt.Printf("%s", apiKey)
	if apiKey == "" {
		return WeatherAPIResponse{}, fmt.Errorf("chave da API não configurada")
	}

	cidade := url.QueryEscape(localidade)
	url := fmt.Sprintf("%s?key=%s&q=%s", weatherAPIURL, apiKey, cidade)
	// imprime a URL para depuração
	resp, err := http.Get(url)
	json.NewEncoder(os.Stdout).Encode(url)
	if err != nil {
		return WeatherAPIResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherAPIResponse{}, fmt.Errorf("erro ao buscar clima: %s", resp.Status)
	}

	var climaAPIResp WeatherAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&climaAPIResp)
	return climaAPIResp, err
}

func validarCEP(cep string) bool {
	re := regexp.MustCompile(`^[0-9]{8}$`)
	if !re.MatchString(cep) {
		return false
	}
	// Implementar a validação do CEP (ex: verificar se tem 8 dígitos)
	if len(cep) != 8 {
		return false
	}
	return true
}
