package entity

type QuoteItemRequest struct {
	LivroID    string `json:"livro_id"`
	Quantidade int    `json:"quantidade"`
}

type QuotePackageRequest struct {
	PesoKG        float64 `json:"peso_kg"`
	LarguraCM     float64 `json:"largura_cm"`
	AlturaCM      float64 `json:"altura_cm"`
	ComprimentoCM float64 `json:"comprimento_cm"`
}

type QuoteRequest struct {
	Provider           string               `json:"provider"`
	CEPDestino         string               `json:"cep_destino"`
	Itens              []QuoteItemRequest   `json:"itens"`
	Pacote             *QuotePackageRequest `json:"pacote,omitempty"`
	ValorDeclarado     float64              `json:"valor_declarado"`
	Servicos           []int                `json:"servicos"`
	RecebimentoProprio bool                 `json:"recebimento_proprio"`
	MaosProprias       bool                 `json:"maos_proprias"`
	AvisoRecebimento   bool                 `json:"aviso_recebimento"`
}

type CatalogItem struct {
	LivroID             string
	Titulo              string
	PesoGramas          int
	LarguraCM           float64
	AlturaCM            float64
	ProfundidadeCM      float64
	PrecoVenda          float64
	PrecoVendaFisico    float64
	PossuiFormatoFisico bool
	Ativo               bool
}

type QuoteOption struct {
	Provider          string  `json:"provider"`
	ServiceCode       string  `json:"service_code"`
	ServiceName       string  `json:"service_name"`
	CarrierName       string  `json:"carrier_name"`
	CarrierPictureURL string  `json:"carrier_picture_url,omitempty"`
	Price             float64 `json:"price"`
	CustomPrice       float64 `json:"custom_price,omitempty"`
	Currency          string  `json:"currency,omitempty"`
	DeliveryDays      int     `json:"delivery_days,omitempty"`
	Error             string  `json:"error,omitempty"`
}

type QuoteResponse struct {
	Provider       string        `json:"provider"`
	CEPOrigem      string        `json:"cep_origem"`
	CEPDestino     string        `json:"cep_destino"`
	SubtotalFisico float64       `json:"subtotal_fisico"`
	Opcoes         []QuoteOption `json:"opcoes"`
}

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
	OriginCEP       string           `json:"origin_cep"`
	OriginCity      string           `json:"origin_city,omitempty"`
	OriginState     string           `json:"origin_state,omitempty"`
	Providers       []ProviderStatus `json:"providers"`
}
