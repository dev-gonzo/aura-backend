package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sistema-editorial/editora/backend/src/pagamentos/entity"
)

type MercadoPagoProvider struct {
	client *http.Client
}

func NewMercadoPagoProvider() *MercadoPagoProvider {
	return &MercadoPagoProvider{
		client: &http.Client{},
	}
}

func (p *MercadoPagoProvider) Code() string {
	return entity.ProviderMercadoPago
}

func (p *MercadoPagoProvider) Label() string {
	return "Mercado Pago"
}

func (p *MercadoPagoProvider) IsEnabled(cfg entity.PaymentConfig) bool {
	return cfg.MercadoPagoEnabled
}

func (p *MercadoPagoProvider) IsConfigured(cfg entity.PaymentConfig) bool {
	return cfg.MercadoPago.IsConfigured()
}

func (p *MercadoPagoProvider) IsSandbox(cfg entity.PaymentConfig) bool {
	return cfg.MercadoPago.Sandbox
}

func (p *MercadoPagoProvider) CreateCheckout(
	ctx context.Context,
	cfg entity.PaymentConfig,
	request entity.CreateCheckoutRequest,
) (entity.CheckoutResponse, error) {
	if !cfg.MercadoPago.IsConfigured() {
		return entity.CheckoutResponse{}, ValidationError{Message: "Mercado Pago nao configurado"}
	}

	payload := mercadoPagoPreferenceRequest{
		ExternalReference:   strings.TrimSpace(request.ExternalReference),
		Items:               make([]mercadoPagoPreferenceItem, 0, len(request.Items)),
		BinaryMode:          cfg.MercadoPago.BinaryMode,
		StatementDescriptor: strings.TrimSpace(cfg.MercadoPago.StatementDescriptor),
	}

	if cfg.MercadoPago.Installments > 0 {
		payload.PaymentMethods = &mercadoPagoPaymentMethods{
			Installments: cfg.MercadoPago.Installments,
		}
	}

	if cfg.MercadoPago.WalletPurchase {
		payload.Purpose = "wallet_purchase"
	}

	successURL := firstNonEmpty(strings.TrimSpace(request.SuccessURL), cfg.MercadoPago.SuccessURL)
	failureURL := firstNonEmpty(strings.TrimSpace(request.FailureURL), cfg.MercadoPago.FailureURL)
	pendingURL := firstNonEmpty(strings.TrimSpace(request.PendingURL), cfg.MercadoPago.PendingURL)
	if successURL != "" || failureURL != "" || pendingURL != "" {
		payload.BackURLs = &mercadoPagoBackURLs{
			Success: successURL,
			Failure: failureURL,
			Pending: pendingURL,
		}
		payload.AutoReturn = "approved"
	}

	payload.NotificationURL = firstNonEmpty(strings.TrimSpace(request.NotificationURL), cfg.MercadoPago.WebhookURL)

	if strings.TrimSpace(request.Payer.Email) != "" {
		payload.Payer = &mercadoPagoPayer{
			Name:    strings.TrimSpace(request.Payer.Name),
			Surname: strings.TrimSpace(request.Payer.Surname),
			Email:   strings.TrimSpace(request.Payer.Email),
		}
		if cpf := digitsOnly(request.Payer.CPF); cpf != "" {
			payload.Payer.Identification = &mercadoPagoIdentification{
				Type:   "CPF",
				Number: cpf,
			}
		}
		if zipCode := digitsOnly(request.Payer.ZipCode); zipCode != "" {
			payload.Payer.Address = &mercadoPagoAddress{
				ZipCode: zipCode,
			}
		}
	}

	for _, item := range request.Items {
		payload.Items = append(payload.Items, mercadoPagoPreferenceItem{
			ID:          strings.TrimSpace(item.ID),
			Title:       strings.TrimSpace(item.Title),
			Description: strings.TrimSpace(item.Description),
			PictureURL:  strings.TrimSpace(item.PictureURL),
			Quantity:    item.Quantity,
			CurrencyID:  "BRL",
			UnitPrice:   item.UnitPrice,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return entity.CheckoutResponse{}, fmt.Errorf("erro ao serializar payload do Mercado Pago: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.MercadoPago.BaseURL+"/checkout/preferences",
		bytes.NewReader(body),
	)
	if err != nil {
		return entity.CheckoutResponse{}, fmt.Errorf("erro ao criar requisicao do Mercado Pago: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+cfg.MercadoPago.AccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(httpRequest)
	if err != nil {
		return entity.CheckoutResponse{}, fmt.Errorf("erro ao conectar com o Mercado Pago: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return entity.CheckoutResponse{}, fmt.Errorf("erro ao ler resposta do Mercado Pago: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return entity.CheckoutResponse{}, fmt.Errorf(
			"erro ao criar checkout no Mercado Pago: %s",
			resolveMercadoPagoError(response.StatusCode, responseBody),
		)
	}

	var parsed mercadoPagoPreferenceResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return entity.CheckoutResponse{}, fmt.Errorf("erro ao interpretar resposta do Mercado Pago: %w", err)
	}

	return entity.CheckoutResponse{
		Provider:           entity.ProviderMercadoPago,
		ExternalReference:  strings.TrimSpace(request.ExternalReference),
		PreferenceID:       strings.TrimSpace(parsed.ID),
		CheckoutURL:        strings.TrimSpace(parsed.InitPoint),
		SandboxCheckoutURL: strings.TrimSpace(parsed.SandboxInitPoint),
		PublicKey:          cfg.MercadoPago.PublicKey,
	}, nil
}

type mercadoPagoPreferenceRequest struct {
	Items               []mercadoPagoPreferenceItem `json:"items"`
	Payer               *mercadoPagoPayer           `json:"payer,omitempty"`
	BackURLs            *mercadoPagoBackURLs        `json:"back_urls,omitempty"`
	PaymentMethods      *mercadoPagoPaymentMethods  `json:"payment_methods,omitempty"`
	NotificationURL     string                      `json:"notification_url,omitempty"`
	StatementDescriptor string                      `json:"statement_descriptor,omitempty"`
	ExternalReference   string                      `json:"external_reference,omitempty"`
	BinaryMode          bool                        `json:"binary_mode,omitempty"`
	AutoReturn          string                      `json:"auto_return,omitempty"`
	Purpose             string                      `json:"purpose,omitempty"`
}

type mercadoPagoPreferenceItem struct {
	ID          string  `json:"id,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	PictureURL  string  `json:"picture_url,omitempty"`
	Quantity    int     `json:"quantity"`
	CurrencyID  string  `json:"currency_id"`
	UnitPrice   float64 `json:"unit_price"`
}

type mercadoPagoPayer struct {
	Name           string                     `json:"name,omitempty"`
	Surname        string                     `json:"surname,omitempty"`
	Email          string                     `json:"email,omitempty"`
	Identification *mercadoPagoIdentification `json:"identification,omitempty"`
	Address        *mercadoPagoAddress        `json:"address,omitempty"`
}

type mercadoPagoIdentification struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type mercadoPagoAddress struct {
	ZipCode string `json:"zip_code,omitempty"`
}

type mercadoPagoBackURLs struct {
	Success string `json:"success,omitempty"`
	Failure string `json:"failure,omitempty"`
	Pending string `json:"pending,omitempty"`
}

type mercadoPagoPaymentMethods struct {
	Installments int `json:"installments,omitempty"`
}

type mercadoPagoPreferenceResponse struct {
	ID               string `json:"id"`
	InitPoint        string `json:"init_point"`
	SandboxInitPoint string `json:"sandbox_init_point"`
}

func resolveMercadoPagoError(statusCode int, body []byte) string {
	type mercadoPagoErrorResponse struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Cause   []struct {
			Description string `json:"description"`
			Code        string `json:"code"`
		} `json:"cause"`
	}

	var parsed mercadoPagoErrorResponse
	if err := json.Unmarshal(body, &parsed); err == nil {
		for _, cause := range parsed.Cause {
			if strings.TrimSpace(cause.Description) != "" {
				return cause.Description
			}
		}
		if strings.TrimSpace(parsed.Message) != "" {
			return parsed.Message
		}
		if strings.TrimSpace(parsed.Error) != "" {
			return parsed.Error
		}
	}

	if trimmed := strings.TrimSpace(string(body)); trimmed != "" {
		return fmt.Sprintf("status %d - %s", statusCode, trimmed)
	}

	return fmt.Sprintf("status %d", statusCode)
}

func digitsOnly(input string) string {
	var builder strings.Builder
	builder.Grow(len(input))
	for _, char := range input {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
