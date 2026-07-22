package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"sistema-editorial/editora/backend/src/pagamentos/entity"
)

type settingsRepository interface {
	Get(ctx context.Context) (entity.SettingsRecord, error)
	Upsert(ctx context.Context, input entity.SettingsRecord) error
}

type provider interface {
	Code() string
	Label() string
	IsEnabled(cfg entity.PaymentConfig) bool
	IsConfigured(cfg entity.PaymentConfig) bool
	IsSandbox(cfg entity.PaymentConfig) bool
	CreateCheckout(ctx context.Context, cfg entity.PaymentConfig, request entity.CreateCheckoutRequest) (entity.CheckoutResponse, error)
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func IsValidationError(err error) bool {
	var target ValidationError
	return errors.As(err, &target)
}

type Service struct {
	settings  settingsRepository
	providers map[string]provider
}

func NewService(settings settingsRepository, providers ...provider) *Service {
	registry := make(map[string]provider, len(providers))
	for _, current := range providers {
		if current == nil {
			continue
		}
		registry[current.Code()] = current
	}

	return &Service{
		settings:  settings,
		providers: registry,
	}
}

func (s *Service) GetConfig(ctx context.Context) (entity.SettingsResponse, error) {
	record, err := s.settings.Get(ctx)
	if err != nil {
		return entity.SettingsResponse{}, fmt.Errorf("erro ao carregar configuracao de pagamentos: %w", err)
	}

	status, err := s.GetProviderStatus(ctx)
	if err != nil {
		return entity.SettingsResponse{}, err
	}

	return entity.SettingsResponse{
		DefaultProvider:    record.DefaultProvider,
		TimeoutSeconds:     record.TimeoutSeconds,
		ContactEmail:       record.ContactEmail,
		MercadoPagoEnabled: record.MercadoPagoEnabled,
		MercadoPago:        record.MercadoPago,
		Providers:          status.Providers,
		CreatedAt:          record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          record.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) GetProviderStatus(ctx context.Context) (entity.ConfigResponse, error) {
	settings, err := s.loadRuntimeConfig(ctx)
	if err != nil {
		return entity.ConfigResponse{}, err
	}

	response := entity.ConfigResponse{
		DefaultProvider: s.resolveProviderCode(settings, ""),
		Providers:       make([]entity.ProviderStatus, 0, len(s.providers)),
	}

	providerCodes := make([]string, 0, len(s.providers))
	for code := range s.providers {
		providerCodes = append(providerCodes, code)
	}
	slices.Sort(providerCodes)

	for _, code := range providerCodes {
		current := s.providers[code]
		response.Providers = append(response.Providers, entity.ProviderStatus{
			Code:       current.Code(),
			Label:      current.Label(),
			Enabled:    current.IsEnabled(settings),
			IsDefault:  code == response.DefaultProvider,
			Configured: current.IsConfigured(settings),
			Sandbox:    current.IsSandbox(settings),
		})
	}

	return response, nil
}

func (s *Service) UpdateConfig(ctx context.Context, request entity.UpdateSettingsRequest) error {
	record, err := s.normalizeSettingsRequest(request)
	if err != nil {
		return err
	}

	if err := s.settings.Upsert(ctx, record); err != nil {
		return fmt.Errorf("erro ao salvar configuracao de pagamentos: %w", err)
	}

	return nil
}

func (s *Service) CreateCheckout(ctx context.Context, request entity.CreateCheckoutRequest) (entity.CheckoutResponse, error) {
	settings, err := s.loadRuntimeConfig(ctx)
	if err != nil {
		return entity.CheckoutResponse{}, err
	}

	providerCode := s.resolveProviderCode(settings, request.Provider)
	current, ok := s.providers[providerCode]
	if !ok {
		return entity.CheckoutResponse{}, ValidationError{Message: "provider de pagamento nao suportado"}
	}
	if !current.IsEnabled(settings) {
		return entity.CheckoutResponse{}, ValidationError{Message: "provider de pagamento nao esta habilitado"}
	}

	if len(request.Items) == 0 {
		return entity.CheckoutResponse{}, ValidationError{Message: "informe ao menos um item para gerar o checkout"}
	}

	for _, item := range request.Items {
		if strings.TrimSpace(item.Title) == "" {
			return entity.CheckoutResponse{}, ValidationError{Message: "titulo do item de pagamento e obrigatorio"}
		}
		if item.Quantity <= 0 {
			return entity.CheckoutResponse{}, ValidationError{Message: "quantidade do item deve ser maior que zero"}
		}
		if item.UnitPrice <= 0 {
			return entity.CheckoutResponse{}, ValidationError{Message: "valor do item deve ser maior que zero"}
		}
	}

	response, err := current.CreateCheckout(ctx, settings, request)
	if err != nil {
		return entity.CheckoutResponse{}, err
	}

	return response, nil
}

func (s *Service) loadRuntimeConfig(ctx context.Context) (entity.PaymentConfig, error) {
	record, err := s.settings.Get(ctx)
	if err != nil {
		return entity.PaymentConfig{}, fmt.Errorf("erro ao carregar configuracao de pagamentos: %w", err)
	}

	timeoutSeconds := record.TimeoutSeconds
	if timeoutSeconds < 5 {
		timeoutSeconds = 15
	}

	return entity.PaymentConfig{
		DefaultProvider:    strings.ToUpper(strings.TrimSpace(record.DefaultProvider)),
		Timeout:            time.Duration(timeoutSeconds) * time.Second,
		ContactEmail:       strings.TrimSpace(record.ContactEmail),
		MercadoPagoEnabled: record.MercadoPagoEnabled,
		MercadoPago:        normalizeMercadoPagoConfig(record),
	}, nil
}

func (s *Service) normalizeSettingsRequest(request entity.UpdateSettingsRequest) (entity.SettingsRecord, error) {
	defaultProvider := strings.ToUpper(strings.TrimSpace(request.DefaultProvider))
	if defaultProvider == "" {
		defaultProvider = entity.ProviderMercadoPago
	}
	if defaultProvider != entity.ProviderMercadoPago {
		return entity.SettingsRecord{}, ValidationError{Message: "provider padrao invalido"}
	}

	timeoutSeconds := request.TimeoutSeconds
	if timeoutSeconds < 5 {
		timeoutSeconds = 15
	}

	record := entity.SettingsRecord{
		DefaultProvider:    defaultProvider,
		TimeoutSeconds:     timeoutSeconds,
		ContactEmail:       strings.TrimSpace(request.ContactEmail),
		MercadoPagoEnabled: request.MercadoPagoEnabled,
		MercadoPago: entity.MercadoPagoConfig{
			Sandbox:             request.MercadoPago.Sandbox,
			BaseURL:             strings.TrimSpace(request.MercadoPago.BaseURL),
			PublicKey:           strings.TrimSpace(request.MercadoPago.PublicKey),
			AccessToken:         strings.TrimSpace(request.MercadoPago.AccessToken),
			StatementDescriptor: strings.TrimSpace(request.MercadoPago.StatementDescriptor),
			SuccessURL:          strings.TrimSpace(request.MercadoPago.SuccessURL),
			FailureURL:          strings.TrimSpace(request.MercadoPago.FailureURL),
			PendingURL:          strings.TrimSpace(request.MercadoPago.PendingURL),
			WebhookURL:          strings.TrimSpace(request.MercadoPago.WebhookURL),
			BinaryMode:          request.MercadoPago.BinaryMode,
			WalletPurchase:      request.MercadoPago.WalletPurchase,
			Installments:        request.MercadoPago.Installments,
		},
	}

	if record.ContactEmail == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "email de contato do pagamento e obrigatorio"}
	}
	if record.MercadoPago.Installments <= 0 {
		record.MercadoPago.Installments = 12
	}
	if record.MercadoPagoEnabled {
		mercadoPago := normalizeMercadoPagoConfig(record)
		if mercadoPago.BaseURL == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "base url do Mercado Pago e obrigatoria"}
		}
		if mercadoPago.PublicKey == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "public key do Mercado Pago e obrigatoria"}
		}
		if mercadoPago.AccessToken == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "access token do Mercado Pago e obrigatorio"}
		}
	}

	return record, nil
}

func normalizeMercadoPagoConfig(record entity.SettingsRecord) entity.MercadoPagoConfig {
	baseURL := strings.TrimSpace(record.MercadoPago.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.mercadopago.com"
	}

	installments := record.MercadoPago.Installments
	if installments <= 0 {
		installments = 12
	}

	return entity.MercadoPagoConfig{
		Sandbox:             record.MercadoPago.Sandbox,
		BaseURL:             strings.TrimRight(baseURL, "/"),
		PublicKey:           strings.TrimSpace(record.MercadoPago.PublicKey),
		AccessToken:         strings.TrimSpace(record.MercadoPago.AccessToken),
		StatementDescriptor: strings.TrimSpace(record.MercadoPago.StatementDescriptor),
		SuccessURL:          strings.TrimSpace(record.MercadoPago.SuccessURL),
		FailureURL:          strings.TrimSpace(record.MercadoPago.FailureURL),
		PendingURL:          strings.TrimSpace(record.MercadoPago.PendingURL),
		WebhookURL:          strings.TrimSpace(record.MercadoPago.WebhookURL),
		BinaryMode:          record.MercadoPago.BinaryMode,
		WalletPurchase:      record.MercadoPago.WalletPurchase,
		Installments:        installments,
	}
}

func (s *Service) resolveProviderCode(settings entity.PaymentConfig, requested string) string {
	code := strings.ToUpper(strings.TrimSpace(requested))
	if code != "" {
		return code
	}
	if strings.TrimSpace(settings.DefaultProvider) != "" {
		return strings.ToUpper(strings.TrimSpace(settings.DefaultProvider))
	}
	return entity.ProviderMercadoPago
}
