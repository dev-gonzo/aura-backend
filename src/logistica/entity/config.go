package entity

import "time"

const (
	ProviderNone        = ""
	ProviderSuperFrete  = "SUPERFRETE"
	ProviderMelhorEnvio = "MELHOR_ENVIO"
	ProviderLoggi       = "LOGGI"
	ProviderFrenet      = "FRENET"
)

type LogisticsConfig struct {
	DefaultProvider    string
	Timeout            time.Duration
	ContactEmail       string
	Origin             OriginConfig
	MelhorEnvioEnabled bool
	MelhorEnvio        MelhorEnvioConfig
	SuperFreteEnabled  bool
	SuperFrete         SuperFreteConfig
	LoggiEnabled       bool
	Loggi              LoggiConfig
	FrenetEnabled      bool
	Frenet             FrenetConfig
}

type OriginConfig struct {
	Name     string
	CEP      string
	Address  string
	Number   string
	District string
	City     string
	State    string
}

type MelhorEnvioConfig struct {
	Sandbox      bool
	BaseURL      string
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	UserAgent    string
}

type SuperFreteConfig struct {
	Sandbox   bool
	BaseURL   string
	Token     string
	UserAgent string
	Services  string
}

type LoggiConfig struct {
	Sandbox           bool
	BaseURL           string
	CompanyID         string
	ClientID          string
	ClientSecret      string
	PickupType        string
	ExternalServiceID string
}

type FrenetConfig struct {
	Sandbox     bool
	BaseURL     string
	Token       string
	Platform    string
	PlatformVer string
}

func (cfg LogisticsConfig) HasOriginAddress() bool {
	return cfg.Origin.CEP != ""
}

func (cfg MelhorEnvioConfig) IsConfigured() bool {
	return cfg.BaseURL != "" && cfg.AccessToken != "" && cfg.UserAgent != ""
}

func (cfg SuperFreteConfig) IsConfigured() bool {
	return cfg.BaseURL != "" && cfg.Token != "" && cfg.UserAgent != ""
}

func (cfg LoggiConfig) IsConfigured() bool {
	return cfg.BaseURL != "" &&
		cfg.CompanyID != "" &&
		cfg.ClientID != "" &&
		cfg.ClientSecret != "" &&
		(cfg.PickupType != "" || cfg.ExternalServiceID != "")
}

func (cfg FrenetConfig) IsConfigured() bool {
	return cfg.BaseURL != "" && cfg.Token != ""
}

type SettingsRecord struct {
	DefaultProvider    string
	TimeoutSeconds     int
	ContactEmail       string
	Origin             OriginConfig
	MelhorEnvioEnabled bool
	MelhorEnvio        MelhorEnvioConfig
	SuperFreteEnabled  bool
	SuperFrete         SuperFreteConfig
	LoggiEnabled       bool
	Loggi              LoggiConfig
	FrenetEnabled      bool
	Frenet             FrenetConfig
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type UpdateSettingsRequest struct {
	DefaultProvider    string            `json:"default_provider"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ContactEmail       string            `json:"contact_email"`
	Origin             OriginConfig      `json:"origin"`
	MelhorEnvioEnabled bool              `json:"melhor_envio_enabled"`
	MelhorEnvio        MelhorEnvioConfig `json:"melhor_envio"`
	SuperFreteEnabled  bool              `json:"superfrete_enabled"`
	SuperFrete         SuperFreteConfig  `json:"superfrete"`
	LoggiEnabled       bool              `json:"loggi_enabled"`
	Loggi              LoggiConfig       `json:"loggi"`
	FrenetEnabled      bool              `json:"frenet_enabled"`
	Frenet             FrenetConfig      `json:"frenet"`
}

type SettingsResponse struct {
	DefaultProvider    string            `json:"default_provider"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ContactEmail       string            `json:"contact_email"`
	Origin             OriginConfig      `json:"origin"`
	MelhorEnvioEnabled bool              `json:"melhor_envio_enabled"`
	MelhorEnvio        MelhorEnvioConfig `json:"melhor_envio"`
	SuperFreteEnabled  bool              `json:"superfrete_enabled"`
	SuperFrete         SuperFreteConfig  `json:"superfrete"`
	LoggiEnabled       bool              `json:"loggi_enabled"`
	Loggi              LoggiConfig       `json:"loggi"`
	FrenetEnabled      bool              `json:"frenet_enabled"`
	Frenet             FrenetConfig      `json:"frenet"`
	Providers          []ProviderStatus  `json:"providers"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}
