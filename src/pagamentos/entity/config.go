package entity

import "time"

const (
	ProviderNone        = ""
	ProviderMercadoPago = "MERCADO_PAGO"
)

type PaymentConfig struct {
	DefaultProvider    string
	Timeout            time.Duration
	ContactEmail       string
	MercadoPagoEnabled bool
	MercadoPago        MercadoPagoConfig
}

type MercadoPagoConfig struct {
	Sandbox             bool   `json:"sandbox"`
	BaseURL             string `json:"base_url"`
	PublicKey           string `json:"public_key"`
	AccessToken         string `json:"access_token"`
	StatementDescriptor string `json:"statement_descriptor"`
	SuccessURL          string `json:"success_url"`
	FailureURL          string `json:"failure_url"`
	PendingURL          string `json:"pending_url"`
	WebhookURL          string `json:"webhook_url"`
	BinaryMode          bool   `json:"binary_mode"`
	WalletPurchase      bool   `json:"wallet_purchase"`
	Installments        int    `json:"installments"`
}

func (cfg MercadoPagoConfig) IsConfigured() bool {
	return cfg.BaseURL != "" && cfg.PublicKey != "" && cfg.AccessToken != ""
}

type SettingsRecord struct {
	DefaultProvider    string
	TimeoutSeconds     int
	ContactEmail       string
	MercadoPagoEnabled bool
	MercadoPago        MercadoPagoConfig
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type UpdateSettingsRequest struct {
	DefaultProvider    string            `json:"default_provider"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ContactEmail       string            `json:"contact_email"`
	MercadoPagoEnabled bool              `json:"mercado_pago_enabled"`
	MercadoPago        MercadoPagoConfig `json:"mercado_pago"`
}

type SettingsResponse struct {
	DefaultProvider    string            `json:"default_provider"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ContactEmail       string            `json:"contact_email"`
	MercadoPagoEnabled bool              `json:"mercado_pago_enabled"`
	MercadoPago        MercadoPagoConfig `json:"mercado_pago"`
	Providers          []ProviderStatus  `json:"providers"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}
