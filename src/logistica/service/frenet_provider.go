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

type FrenetProvider struct{}

func NewFrenetProvider() *FrenetProvider {
	return &FrenetProvider{}
}

func (p *FrenetProvider) Code() string {
	return entity.ProviderFrenet
}

func (p *FrenetProvider) Label() string {
	return "Frenet"
}

func (p *FrenetProvider) IsEnabled(cfg entity.LogisticsConfig) bool {
	return cfg.FrenetEnabled && p.IsConfigured(cfg)
}

func (p *FrenetProvider) IsConfigured(cfg entity.LogisticsConfig) bool {
	return cfg.Frenet.IsConfigured()
}

func (p *FrenetProvider) IsSandbox(cfg entity.LogisticsConfig) bool {
	return cfg.Frenet.Sandbox
}

func (p *FrenetProvider) Quote(
	ctx context.Context,
	cfg entity.LogisticsConfig,
	input ProviderQuoteInput,
) ([]entity.QuoteOption, error) {
	payload := frenetQuotePayload{
		Token:                strings.TrimSpace(cfg.Frenet.Token),
		PlatformName:         strings.TrimSpace(cfg.Frenet.Platform),
		PlatformVersion:      strings.TrimSpace(cfg.Frenet.PlatformVer),
		SellerCEP:            digitsOnly(input.Origin.CEP),
		RecipientCEP:         digitsOnly(input.DestinationCEP),
		ShipmentInvoiceValue: normalizeDecimal(input.DeclaredValue),
		RecipientCountry:     "BR",
		ShippingItemArray:    make([]frenetShippingItem, 0, len(input.Items)),
	}

	for _, item := range input.Items {
		payload.ShippingItemArray = append(payload.ShippingItemArray, frenetShippingItem{
			Weight:    normalizeDecimal(item.WeightKG),
			Length:    normalizeDimension(item.LengthCM),
			Height:    normalizeDimension(item.HeightCM),
			Width:     normalizeDimension(item.WidthCM),
			Diameter:  0,
			SKU:       item.ID,
			Category:  item.Title,
			IsFragile: false,
			Quantity:  item.Quantity,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload da Frenet: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(cfg.Frenet.BaseURL, "/")+"/shipping/quote",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao montar requisicao da Frenet: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: cfg.Timeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("erro de comunicacao com a Frenet: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		var apiErr frenetErrorResponse
		if err := json.NewDecoder(response.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("erro da Frenet com status %d", response.StatusCode)
		}

		message := firstNonEmpty(
			apiErr.Message,
			apiErr.Error,
			apiErr.Detail,
			apiErr.Status,
		)
		if message == "" {
			message = fmt.Sprintf("status %d", response.StatusCode)
		}
		return nil, ValidationError{Message: "erro ao cotar frete na Frenet: " + message}
	}

	var decoded frenetQuoteResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("erro ao interpretar resposta da Frenet: %w", err)
	}

	if !decoded.Success && decoded.Message != "" {
		return nil, ValidationError{Message: "erro ao cotar frete na Frenet: " + strings.TrimSpace(decoded.Message)}
	}

	options := decoded.toQuoteOptions(p.Code())
	if len(options) == 0 {
		return nil, ValidationError{Message: "a Frenet nao retornou opcoes de frete para esta consulta"}
	}

	return options, nil
}

type frenetQuotePayload struct {
	Token                string               `json:"Token"`
	Coupom               string               `json:"Coupom,omitempty"`
	PlatformName         string               `json:"PlatformName,omitempty"`
	PlatformVersion      string               `json:"PlatformVersion,omitempty"`
	SellerCEP            string               `json:"SellerCEP"`
	RecipientCEP         string               `json:"RecipientCEP"`
	RecipientDocument    string               `json:"RecipientDocument,omitempty"`
	ShipmentInvoiceValue float64              `json:"ShipmentInvoiceValue"`
	ShippingItemArray    []frenetShippingItem `json:"ShippingItemArray"`
	RecipientCountry     string               `json:"RecipientCountry"`
	Timeout              int                  `json:"Timeout,omitempty"`
}

type frenetShippingItem struct {
	Weight    float64 `json:"Weight"`
	Length    float64 `json:"Length"`
	Height    float64 `json:"Height"`
	Width     float64 `json:"Width"`
	Diameter  float64 `json:"Diameter"`
	SKU       string  `json:"SKU,omitempty"`
	Category  string  `json:"Category,omitempty"`
	IsFragile bool    `json:"isFragile"`
	Quantity  int     `json:"Quantity,omitempty"`
}

type frenetErrorResponse struct {
	Message string `json:"Message"`
	Error   string `json:"Error"`
	Detail  string `json:"Detail"`
	Status  string `json:"Status"`
}

type frenetQuoteResponse struct {
	Success      bool                 `json:"Success"`
	Message      string               `json:"Message"`
	ShippingSev  []frenetQuoteService `json:"ShippingSevicesArray"`
	ShippingServ []frenetQuoteService `json:"ShippingServicesArray"`
	Services     []frenetQuoteService `json:"Services"`
}

type frenetQuoteService struct {
	ShippingServiceCode        string          `json:"ShippingServiceCode"`
	ServiceCode                string          `json:"ServiceCode"`
	ShippingServiceDescription string          `json:"ShippingServiceDescription"`
	ServiceDescription         string          `json:"ServiceDescription"`
	ShippingServiceName        string          `json:"ShippingServiceName"`
	Carrier                    string          `json:"Carrier"`
	CarrierCode                string          `json:"CarrierCode"`
	CarrierLogo                string          `json:"CarrierLogo"`
	CarrierPicture             string          `json:"CarrierPicture"`
	ShippingPrice              json.RawMessage `json:"ShippingPrice"`
	PlatformShippingPrice      json.RawMessage `json:"PlatformShippingPrice"`
	Price                      json.RawMessage `json:"Price"`
	DeliveryTime               int             `json:"DeliveryTime"`
	OriginalDeliveryTime       int             `json:"OriginalDeliveryTime"`
	Error                      string          `json:"Error"`
	ErrorMessage               string          `json:"ErrorMessage"`
}

func (response frenetQuoteResponse) toQuoteOptions(providerCode string) []entity.QuoteOption {
	candidates := make([]frenetQuoteService, 0, len(response.ShippingSev)+len(response.ShippingServ)+len(response.Services))
	candidates = append(candidates, response.ShippingSev...)
	candidates = append(candidates, response.ShippingServ...)
	candidates = append(candidates, response.Services...)

	options := make([]entity.QuoteOption, 0, len(candidates))
	for _, current := range candidates {
		deliveryDays := current.DeliveryTime
		if deliveryDays <= 0 {
			deliveryDays = current.OriginalDeliveryTime
		}

		options = append(options, entity.QuoteOption{
			Provider:          providerCode,
			ServiceCode:       firstNonEmpty(current.ShippingServiceCode, current.ServiceCode, current.CarrierCode),
			ServiceName:       firstNonEmpty(current.ShippingServiceDescription, current.ServiceDescription, current.ShippingServiceName, "Frenet"),
			CarrierName:       firstNonEmpty(current.Carrier, "Frenet"),
			CarrierPictureURL: firstNonEmpty(current.CarrierPicture, current.CarrierLogo),
			Price:             parseLoggiPrice(current.PlatformShippingPrice, current.ShippingPrice, current.Price),
			Currency:          "BRL",
			DeliveryDays:      deliveryDays,
			Error:             firstNonEmpty(current.Error, current.ErrorMessage),
		})
	}

	return options
}

func normalizeDimension(value float64) float64 {
	if value <= 0 {
		return 1
	}
	return normalizeDecimal(value)
}
