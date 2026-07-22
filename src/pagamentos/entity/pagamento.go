package entity

type ProviderStatus struct {
	Code       string `json:"code"`
	Label      string `json:"label"`
	Enabled    bool   `json:"enabled"`
	IsDefault  bool   `json:"is_default"`
	Configured bool   `json:"configured"`
	Sandbox    bool   `json:"sandbox,omitempty"`
}

type ConfigResponse struct {
	DefaultProvider string           `json:"default_provider"`
	Providers       []ProviderStatus `json:"providers"`
}

type CheckoutItemRequest struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	PictureURL  string  `json:"picture_url"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type CheckoutPayerRequest struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
	CPF     string `json:"cpf"`
	ZipCode string `json:"zip_code"`
}

type CreateCheckoutRequest struct {
	Provider          string                `json:"provider"`
	ExternalReference string                `json:"external_reference"`
	Items             []CheckoutItemRequest `json:"items"`
	Payer             CheckoutPayerRequest  `json:"payer"`
	SuccessURL        string                `json:"success_url"`
	FailureURL        string                `json:"failure_url"`
	PendingURL        string                `json:"pending_url"`
	NotificationURL   string                `json:"notification_url"`
}

type CheckoutResponse struct {
	Provider           string `json:"provider"`
	ExternalReference  string `json:"external_reference"`
	PreferenceID       string `json:"preference_id"`
	CheckoutURL        string `json:"checkout_url"`
	SandboxCheckoutURL string `json:"sandbox_checkout_url"`
	PublicKey          string `json:"public_key"`
}
