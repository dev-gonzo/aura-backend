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

type LoggiProvider struct{}

func NewLoggiProvider() *LoggiProvider {
	return &LoggiProvider{}
}

func (p *LoggiProvider) Code() string {
	return entity.ProviderLoggi
}

func (p *LoggiProvider) Label() string {
	return "Loggi"
}

func (p *LoggiProvider) IsEnabled(cfg entity.LogisticsConfig) bool {
	return cfg.LoggiEnabled && p.IsConfigured(cfg)
}

func (p *LoggiProvider) IsConfigured(cfg entity.LogisticsConfig) bool {
	return cfg.Loggi.IsConfigured()
}

func (p *LoggiProvider) IsSandbox(cfg entity.LogisticsConfig) bool {
	return cfg.Loggi.Sandbox
}

func (p *LoggiProvider) Quote(
	ctx context.Context,
	cfg entity.LogisticsConfig,
	input ProviderQuoteInput,
) ([]entity.QuoteOption, error) {
	token, err := p.authenticate(ctx, cfg)
	if err != nil {
		return nil, err
	}

	payload := loggiQuotePayload{
		ShipFrom: loggiAddressPayload{
			Correios: loggiCorreiosAddress{
				PostalCode: input.Origin.CEP,
				Street:     input.Origin.Address,
				Number:     input.Origin.Number,
				Complement: input.Origin.District,
				District:   input.Origin.District,
				City:       input.Origin.City,
				State:      input.Origin.State,
			},
		},
		ShipTo: loggiAddressPayload{
			Correios: loggiCorreiosAddress{
				PostalCode: input.DestinationCEP,
			},
		},
		Packages: make([]loggiPackagePayload, 0, len(input.Items)),
	}

	if strings.TrimSpace(cfg.Loggi.ExternalServiceID) != "" {
		payload.ExternalServiceIDs = []string{strings.TrimSpace(cfg.Loggi.ExternalServiceID)}
	} else {
		payload.PickupTypes = []string{strings.TrimSpace(cfg.Loggi.PickupType)}
	}

	for _, item := range input.Items {
		payload.Packages = append(payload.Packages, loggiPackagePayload{
			Weight:         normalizeDecimal(item.WeightKG),
			Width:          normalizeDecimal(item.WidthCM),
			Height:         normalizeDecimal(item.HeightCM),
			Length:         normalizeDecimal(item.LengthCM),
			InsuranceValue: normalizeDecimal(item.InsuranceValue),
			Quantity:       item.Quantity,
			Reference:      item.ID,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload da Loggi: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(cfg.Loggi.BaseURL, "/")+"/v1/companies/"+cfg.Loggi.CompanyID+"/quotations",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao montar requisicao da Loggi: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: cfg.Timeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("erro de comunicacao com a Loggi: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		var apiErr loggiAPIErrorResponse
		if err := json.NewDecoder(response.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("erro da Loggi com status %d", response.StatusCode)
		}

		message := firstNonEmpty(
			apiErr.Title,
			apiErr.Message,
			apiErr.Detail,
			apiErr.Error,
		)
		if message == "" {
			message = fmt.Sprintf("status %d", response.StatusCode)
		}
		return nil, ValidationError{Message: "erro ao cotar frete na Loggi: " + message}
	}

	var decoded loggiQuoteResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("erro ao interpretar resposta da Loggi: %w", err)
	}

	options := decoded.toQuoteOptions(p.Code())
	if len(options) == 0 {
		return nil, ValidationError{Message: "a Loggi nao retornou opcoes de frete para esta consulta"}
	}

	return options, nil
}

func (p *LoggiProvider) authenticate(ctx context.Context, cfg entity.LogisticsConfig) (string, error) {
	payload := map[string]string{
		"client_id":     cfg.Loggi.ClientID,
		"client_secret": cfg.Loggi.ClientSecret,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar autenticacao da Loggi: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(cfg.Loggi.BaseURL, "/")+"/v2/oauth2/token",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("erro ao montar autenticacao da Loggi: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: cfg.Timeout}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("erro de comunicacao com a autenticacao da Loggi: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		var apiErr loggiAPIErrorResponse
		if err := json.NewDecoder(response.Body).Decode(&apiErr); err != nil {
			return "", fmt.Errorf("erro de autenticacao da Loggi com status %d", response.StatusCode)
		}

		message := firstNonEmpty(
			apiErr.Title,
			apiErr.Message,
			apiErr.Detail,
			apiErr.Error,
		)
		if message == "" {
			message = fmt.Sprintf("status %d", response.StatusCode)
		}

		return "", ValidationError{Message: "erro ao autenticar na Loggi: " + message}
	}

	var authResponse loggiAuthResponse
	if err := json.NewDecoder(response.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("erro ao interpretar autenticacao da Loggi: %w", err)
	}
	if strings.TrimSpace(authResponse.AccessToken) == "" {
		return "", ValidationError{Message: "a autenticacao da Loggi nao retornou access token"}
	}

	return strings.TrimSpace(authResponse.AccessToken), nil
}

type loggiQuotePayload struct {
	ExternalServiceIDs []string              `json:"externalServiceIds,omitempty"`
	PickupTypes        []string              `json:"pickupTypes,omitempty"`
	ShipFrom           loggiAddressPayload   `json:"shipFrom"`
	ShipTo             loggiAddressPayload   `json:"shipTo"`
	Packages           []loggiPackagePayload `json:"packages"`
}

type loggiAddressPayload struct {
	Correios loggiCorreiosAddress `json:"correios"`
}

type loggiCorreiosAddress struct {
	PostalCode string `json:"postalCode"`
	Street     string `json:"street,omitempty"`
	Number     string `json:"number,omitempty"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
}

type loggiPackagePayload struct {
	Weight         float64 `json:"weight"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
	Length         float64 `json:"length"`
	InsuranceValue float64 `json:"insuranceValue,omitempty"`
	Quantity       int     `json:"quantity,omitempty"`
	Reference      string  `json:"reference,omitempty"`
}

type loggiAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type loggiAPIErrorResponse struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Error   string `json:"error"`
}

type loggiQuoteResponse struct {
	Quotations []loggiQuotationEnvelope `json:"quotations"`
	Packages   []struct {
		Quotations []loggiQuotationEnvelope `json:"quotations"`
		Error      string                   `json:"error"`
	} `json:"packages"`
	Message string `json:"message"`
}

type loggiQuotationEnvelope struct {
	ID                string          `json:"id"`
	ServiceCode       string          `json:"serviceCode"`
	ServiceName       string          `json:"serviceName"`
	ServiceType       string          `json:"serviceType"`
	ExternalServiceID string          `json:"externalServiceId"`
	PickupType        string          `json:"pickupType"`
	Carrier           loggiCarrier    `json:"carrier"`
	Price             json.RawMessage `json:"price"`
	FinalPrice        json.RawMessage `json:"finalPrice"`
	TotalPrice        json.RawMessage `json:"totalPrice"`
	DeliveryDays      int             `json:"deliveryDays"`
	DeliveryRange     struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"deliveryRangeInDays"`
	Error string `json:"error"`
}

type loggiCarrier struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Logo  string `json:"logo"`
}

func (response loggiQuoteResponse) toQuoteOptions(providerCode string) []entity.QuoteOption {
	candidates := make([]loggiQuotationEnvelope, 0, len(response.Quotations))
	candidates = append(candidates, response.Quotations...)
	for _, currentPackage := range response.Packages {
		candidates = append(candidates, currentPackage.Quotations...)
	}

	options := make([]entity.QuoteOption, 0, len(candidates))
	for _, current := range candidates {
		deliveryDays := current.DeliveryDays
		if deliveryDays <= 0 {
			deliveryDays = current.DeliveryRange.Max
		}

		options = append(options, entity.QuoteOption{
			Provider:          providerCode,
			ServiceCode:       firstNonEmpty(current.ServiceCode, current.ExternalServiceID, current.PickupType, current.ID),
			ServiceName:       firstNonEmpty(current.ServiceName, current.ServiceType, "Loggi"),
			CarrierName:       firstNonEmpty(current.Carrier.Name, "Loggi"),
			CarrierPictureURL: firstNonEmpty(current.Carrier.Image, current.Carrier.Logo),
			Price:             parseLoggiPrice(current.Price, current.FinalPrice, current.TotalPrice),
			Currency:          "BRL",
			DeliveryDays:      deliveryDays,
			Error:             strings.TrimSpace(current.Error),
		})
	}

	return options
}

func parseLoggiPrice(values ...json.RawMessage) float64 {
	for _, current := range values {
		if len(current) == 0 {
			continue
		}

		var number float64
		if err := json.Unmarshal(current, &number); err == nil {
			return number
		}

		var text string
		if err := json.Unmarshal(current, &text); err == nil {
			return parseFloat(text)
		}

		var payload map[string]any
		if err := json.Unmarshal(current, &payload); err == nil {
			for _, key := range []string{"amount", "value", "total", "final"} {
				if parsed := parseAnyFloat(payload[key]); parsed > 0 {
					return parsed
				}
			}
		}
	}

	return 0
}

func parseAnyFloat(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		parsed, err := typed.Float64()
		if err == nil {
			return parsed
		}
	case string:
		return parseFloat(typed)
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, current := range values {
		if strings.TrimSpace(current) != "" {
			return strings.TrimSpace(current)
		}
	}
	return ""
}
