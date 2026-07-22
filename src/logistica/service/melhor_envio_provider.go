package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sistema-editorial/editora/backend/src/logistica/entity"
)

type MelhorEnvioProvider struct{}

func NewMelhorEnvioProvider() *MelhorEnvioProvider {
	return &MelhorEnvioProvider{}
}

func (p *MelhorEnvioProvider) Code() string {
	return entity.ProviderMelhorEnvio
}

func (p *MelhorEnvioProvider) Label() string {
	return "Melhor Envio"
}

func (p *MelhorEnvioProvider) IsEnabled(cfg entity.LogisticsConfig) bool {
	return cfg.MelhorEnvioEnabled && p.IsConfigured(cfg)
}

func (p *MelhorEnvioProvider) IsConfigured(cfg entity.LogisticsConfig) bool {
	return cfg.MelhorEnvio.IsConfigured()
}

func (p *MelhorEnvioProvider) IsSandbox(cfg entity.LogisticsConfig) bool {
	return cfg.MelhorEnvio.Sandbox
}

func (p *MelhorEnvioProvider) Quote(
	ctx context.Context,
	cfg entity.LogisticsConfig,
	input ProviderQuoteInput,
) ([]entity.QuoteOption, error) {
	client := &http.Client{Timeout: cfg.Timeout}
	payload := melhorEnvioQuotePayload{
		From: melhorEnvioAddress{
			PostalCode: input.Origin.CEP,
			Address:    input.Origin.Address,
			Number:     input.Origin.Number,
			District:   input.Origin.District,
			City:       input.Origin.City,
			StateAbbr:  input.Origin.State,
		},
		To: melhorEnvioAddress{
			PostalCode: input.DestinationCEP,
		},
		Products: make([]melhorEnvioProduct, 0, len(input.Items)),
		Options: melhorEnvioQuoteOptions{
			Receipt:  input.Receipt,
			OwnHand:  input.OwnHand,
			Platform: "Aura Editora",
		},
		Services: input.Services,
	}

	for _, item := range input.Items {
		payload.Products = append(payload.Products, melhorEnvioProduct{
			ID:             item.ID,
			Width:          normalizeDecimal(item.WidthCM),
			Height:         normalizeDecimal(item.HeightCM),
			Length:         normalizeDecimal(item.LengthCM),
			Weight:         normalizeDecimal(item.WeightKG),
			InsuranceValue: normalizeDecimal(item.InsuranceValue),
			Quantity:       item.Quantity,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload do Melhor Envio: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.MelhorEnvio.BaseURL+"/me/shipment/calculate",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao montar requisicao do Melhor Envio: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+cfg.MelhorEnvio.AccessToken)
	request.Header.Set("User-Agent", cfg.MelhorEnvio.UserAgent)

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("erro de comunicacao com o Melhor Envio: %w", err)
	}
	defer response.Body.Close()

	var result melhorEnvioErrorResponse
	if response.StatusCode >= http.StatusBadRequest {
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("erro do Melhor Envio com status %d", response.StatusCode)
		}
		if result.Message != "" {
			return nil, ValidationError{Message: "erro ao cotar frete no Melhor Envio: " + result.Message}
		}
		return nil, ValidationError{Message: fmt.Sprintf("erro ao cotar frete no Melhor Envio: status %d", response.StatusCode)}
	}

	var quoteResults []melhorEnvioQuoteResult
	if err := json.NewDecoder(response.Body).Decode(&quoteResults); err != nil {
		return nil, fmt.Errorf("erro ao interpretar resposta do Melhor Envio: %w", err)
	}

	options := make([]entity.QuoteOption, 0, len(quoteResults))
	for _, current := range quoteResults {
		options = append(options, entity.QuoteOption{
			Provider:          p.Code(),
			ServiceCode:       strconv.Itoa(current.ID),
			ServiceName:       strings.TrimSpace(current.Name),
			CarrierName:       strings.TrimSpace(current.Company.Name),
			CarrierPictureURL: strings.TrimSpace(current.Company.Picture),
			Price:             parseFloat(current.Price),
			CustomPrice:       parseFloat(current.CustomPrice),
			Currency:          strings.TrimSpace(current.Currency),
			DeliveryDays:      current.DeliveryTime,
			Error:             strings.TrimSpace(current.Error),
		})
	}

	return options, nil
}

type melhorEnvioQuotePayload struct {
	From     melhorEnvioAddress      `json:"from"`
	To       melhorEnvioAddress      `json:"to"`
	Products []melhorEnvioProduct    `json:"products"`
	Options  melhorEnvioQuoteOptions `json:"options,omitempty"`
	Services []int                   `json:"services,omitempty"`
}

type melhorEnvioAddress struct {
	PostalCode string `json:"postal_code"`
	Address    string `json:"address,omitempty"`
	Number     string `json:"number,omitempty"`
	District   string `json:"district,omitempty"`
	City       string `json:"city,omitempty"`
	StateAbbr  string `json:"state_abbr,omitempty"`
}

type melhorEnvioProduct struct {
	ID             string  `json:"id,omitempty"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
	Length         float64 `json:"length"`
	Weight         float64 `json:"weight"`
	InsuranceValue float64 `json:"insurance_value"`
	Quantity       int     `json:"quantity"`
}

type melhorEnvioQuoteOptions struct {
	Receipt  bool   `json:"receipt"`
	OwnHand  bool   `json:"own_hand"`
	Platform string `json:"platform,omitempty"`
}

type melhorEnvioQuoteResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Price        string `json:"price"`
	CustomPrice  string `json:"custom_price"`
	Currency     string `json:"currency"`
	DeliveryTime int    `json:"delivery_time"`
	Error        string `json:"error"`
	Company      struct {
		Name    string `json:"name"`
		Picture string `json:"picture"`
	} `json:"company"`
}

type melhorEnvioErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func normalizeDecimal(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func parseFloat(value string) float64 {
	normalized := strings.TrimSpace(strings.ReplaceAll(value, ",", "."))
	if normalized == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0
	}
	return parsed
}
