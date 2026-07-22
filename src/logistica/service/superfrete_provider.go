package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"sistema-editorial/editora/backend/src/logistica/entity"
)

type SuperFreteProvider struct{}

func NewSuperFreteProvider() *SuperFreteProvider {
	return &SuperFreteProvider{}
}

func (p *SuperFreteProvider) Code() string {
	return entity.ProviderSuperFrete
}

func (p *SuperFreteProvider) Label() string {
	return "SuperFrete"
}

func (p *SuperFreteProvider) IsEnabled(cfg entity.LogisticsConfig) bool {
	return cfg.SuperFreteEnabled && p.IsConfigured(cfg)
}

func (p *SuperFreteProvider) IsConfigured(cfg entity.LogisticsConfig) bool {
	return cfg.SuperFrete.IsConfigured()
}

func (p *SuperFreteProvider) IsSandbox(cfg entity.LogisticsConfig) bool {
	return cfg.SuperFrete.Sandbox
}

func (p *SuperFreteProvider) Quote(
	ctx context.Context,
	cfg entity.LogisticsConfig,
	input ProviderQuoteInput,
) ([]entity.QuoteOption, error) {
	payload := superFreteQuotePayload{
		From: superFretePostalCodePayload{
			PostalCode: digitsOnly(input.Origin.CEP),
		},
		To: superFretePostalCodePayload{
			PostalCode: digitsOnly(input.DestinationCEP),
		},
		Services: resolveSuperFreteServices(cfg.SuperFrete.Services, input.Services),
		Options: superFreteOptionsPayload{
			OwnHand:           input.OwnHand,
			Receipt:           input.Receipt,
			InsuranceValue:    normalizeDecimal(input.DeclaredValue),
			UseInsuranceValue: input.DeclaredValue > 0,
		},
		Products: make([]superFreteProductPayload, 0, len(input.Items)),
	}

	for _, item := range input.Items {
		payload.Products = append(payload.Products, superFreteProductPayload{
			Quantity: item.Quantity,
			Weight:   normalizeDecimal(item.WeightKG),
			Height:   normalizeDimension(item.HeightCM),
			Width:    normalizeDimension(item.WidthCM),
			Length:   normalizeDimension(item.LengthCM),
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload do SuperFrete: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(cfg.SuperFrete.BaseURL, "/")+"/api/v0/calculator",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao montar requisicao do SuperFrete: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cfg.SuperFrete.Token))
	request.Header.Set("User-Agent", strings.TrimSpace(cfg.SuperFrete.UserAgent))

	client := &http.Client{Timeout: cfg.Timeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("erro de comunicacao com o SuperFrete: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		var apiErr superFreteErrorResponse
		if err := json.NewDecoder(response.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("erro do SuperFrete com status %d", response.StatusCode)
		}

		message := firstNonEmpty(
			apiErr.Error,
			apiErr.Message,
			apiErr.Details,
			apiErr.Code,
		)
		if message == "" {
			message = fmt.Sprintf("status %d", response.StatusCode)
		}
		return nil, ValidationError{Message: "erro ao cotar frete no SuperFrete: " + message}
	}

	var decoded superFreteQuoteResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("erro ao interpretar resposta do SuperFrete: %w", err)
	}

	options := decoded.toQuoteOptions(p.Code())
	if len(options) == 0 {
		return nil, ValidationError{Message: "o SuperFrete nao retornou opcoes de frete para esta consulta"}
	}

	return options, nil
}

type superFreteQuotePayload struct {
	From     superFretePostalCodePayload `json:"from"`
	To       superFretePostalCodePayload `json:"to"`
	Services string                      `json:"services"`
	Options  superFreteOptionsPayload    `json:"options"`
	Products []superFreteProductPayload  `json:"products,omitempty"`
}

type superFretePostalCodePayload struct {
	PostalCode string `json:"postal_code"`
}

type superFreteOptionsPayload struct {
	OwnHand           bool    `json:"own_hand"`
	Receipt           bool    `json:"receipt"`
	InsuranceValue    float64 `json:"insurance_value"`
	UseInsuranceValue bool    `json:"use_insurance_value"`
}

type superFreteProductPayload struct {
	Quantity int     `json:"quantity"`
	Weight   float64 `json:"weight"`
	Height   float64 `json:"height"`
	Width    float64 `json:"width"`
	Length   float64 `json:"length"`
}

type superFreteErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details"`
	Code    string `json:"code"`
}

type superFreteQuoteResponse []superFreteQuoteOption

type superFreteQuoteOption struct {
	ID           json.RawMessage `json:"id"`
	Name         string          `json:"name"`
	Service      string          `json:"service"`
	Company      superFreteBrand `json:"company"`
	Price        json.RawMessage `json:"price"`
	CustomPrice  json.RawMessage `json:"custom_price"`
	Currency     string          `json:"currency"`
	DeliveryTime int             `json:"delivery_time"`
	DeliveryDays int             `json:"delivery_days"`
	Error        string          `json:"error"`
}

type superFreteBrand struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Logo    string `json:"logo"`
}

func (response superFreteQuoteResponse) toQuoteOptions(providerCode string) []entity.QuoteOption {
	options := make([]entity.QuoteOption, 0, len(response))
	for _, current := range response {
		deliveryDays := current.DeliveryDays
		if deliveryDays <= 0 {
			deliveryDays = current.DeliveryTime
		}

		serviceID := rawMessageToString(current.ID)

		options = append(options, entity.QuoteOption{
			Provider:          providerCode,
			ServiceCode:       firstNonEmpty(serviceID, current.Service),
			ServiceName:       firstNonEmpty(current.Name, current.Service, "SuperFrete"),
			CarrierName:       firstNonEmpty(current.Company.Name, "SuperFrete"),
			CarrierPictureURL: firstNonEmpty(current.Company.Picture, current.Company.Logo),
			Price:             parseLoggiPrice(current.CustomPrice, current.Price),
			CustomPrice:       parseLoggiPrice(current.CustomPrice),
			Currency:          firstNonEmpty(current.Currency, "BRL"),
			DeliveryDays:      deliveryDays,
			Error:             strings.TrimSpace(current.Error),
		})
	}

	return options
}

func resolveSuperFreteServices(defaultServices string, requested []int) string {
	if len(requested) > 0 {
		parts := make([]string, 0, len(requested))
		for _, current := range requested {
			if current > 0 {
				parts = append(parts, fmt.Sprintf("%d", current))
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ",")
		}
	}

	services := strings.TrimSpace(defaultServices)
	if services == "" {
		return "1,2,17"
	}
	return services
}

func rawMessageToString(value json.RawMessage) string {
	if len(value) == 0 {
		return ""
	}

	var text string
	if err := json.Unmarshal(value, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var number json.Number
	if err := json.Unmarshal(value, &number); err == nil {
		return number.String()
	}

	var integer int64
	if err := json.Unmarshal(value, &integer); err == nil {
		return fmt.Sprintf("%d", integer)
	}

	var decimal float64
	if err := json.Unmarshal(value, &decimal); err == nil {
		return fmt.Sprintf("%.0f", decimal)
	}

	return strings.TrimSpace(string(value))
}
