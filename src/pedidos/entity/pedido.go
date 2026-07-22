package entity

type ItemRequest struct {
	LivroID       string  `json:"livro_id"`
	Quantidade    int     `json:"quantidade"`
	PrecoUnitario float64 `json:"preco_unitario"`
}

type EntregaRequest struct {
	TipoEntrega           string `json:"tipo_entrega"`
	StatusEntrega         string `json:"status_entrega"`
	Transportadora        string `json:"transportadora"`
	CodigoRastreio        string `json:"codigo_rastreio"`
	DestinatarioNome      string `json:"destinatario_nome"`
	DestinatarioDocumento string `json:"destinatario_documento"`
	CEP                   string `json:"cep"`
	Logradouro            string `json:"logradouro"`
	Numero                string `json:"numero"`
	Complemento           string `json:"complemento"`
	Bairro                string `json:"bairro"`
	Cidade                string `json:"cidade"`
	UF                    string `json:"uf"`
	PrazoPrevistoEm       string `json:"prazo_previsto_em"`
	PostadoEm             string `json:"postado_em"`
	EntregueEm            string `json:"entregue_em"`
	Observacao            string `json:"observacao"`
}

type CreateRequest struct {
	CanalVenda      string          `json:"canal_venda"`
	Status          string          `json:"status"`
	ClienteNome     string          `json:"cliente_nome"`
	ClienteEmail    string          `json:"cliente_email"`
	ClienteWhatsapp string          `json:"cliente_whatsapp"`
	Desconto        float64         `json:"desconto"`
	Frete           float64         `json:"frete"`
	Observacao      string          `json:"observacao"`
	Itens           []ItemRequest   `json:"itens"`
	Entrega         *EntregaRequest `json:"entrega"`
}

type UpdateRequest = CreateRequest

type PersistItem struct {
	LivroID       string
	TituloLivro   string
	AutorNome     string
	Quantidade    int
	PrecoUnitario float64
	Subtotal      float64
}

type PersistEntrega struct {
	TipoEntrega           string
	StatusEntrega         string
	Transportadora        *string
	CodigoRastreio        *string
	DestinatarioNome      string
	DestinatarioDocumento *string
	CEP                   *string
	Logradouro            *string
	Numero                *string
	Complemento           *string
	Bairro                *string
	Cidade                *string
	UF                    *string
	PrazoPrevistoEm       *string
	PostadoEm             *string
	EntregueEm            *string
	Observacao            *string
}

type PersistInput struct {
	ID              string
	Codigo          string
	CanalVenda      string
	Status          string
	ClienteNome     string
	ClienteEmail    *string
	ClienteWhatsapp *string
	Subtotal        float64
	Desconto        float64
	Frete           float64
	Total           float64
	Observacao      *string
	Itens           []PersistItem
	Entrega         *PersistEntrega
}

type ListItem struct {
	ID              string  `json:"id"`
	Codigo          string  `json:"codigo"`
	CanalVenda      string  `json:"canal_venda"`
	Status          string  `json:"status"`
	ClienteNome     string  `json:"cliente_nome"`
	ClienteEmail    *string `json:"cliente_email,omitempty"`
	Subtotal        float64 `json:"subtotal"`
	Desconto        float64 `json:"desconto"`
	Frete           float64 `json:"frete"`
	Total           float64 `json:"total"`
	ItensQuantidade int     `json:"itens_quantidade"`
	CriadoEm        string  `json:"criado_em"`
	AtualizadoEm    string  `json:"atualizado_em"`
}

type DetailResponse struct {
	ID              string           `json:"id"`
	Codigo          string           `json:"codigo"`
	CanalVenda      string           `json:"canal_venda"`
	Status          string           `json:"status"`
	ClienteNome     string           `json:"cliente_nome"`
	ClienteEmail    *string          `json:"cliente_email,omitempty"`
	ClienteWhatsapp *string          `json:"cliente_whatsapp,omitempty"`
	Subtotal        float64          `json:"subtotal"`
	Desconto        float64          `json:"desconto"`
	Frete           float64          `json:"frete"`
	Total           float64          `json:"total"`
	Observacao      *string          `json:"observacao,omitempty"`
	Itens           []PersistItem    `json:"itens"`
	Entrega         *PersistEntrega  `json:"entrega,omitempty"`
	CriadoEm        string           `json:"criado_em"`
	AtualizadoEm    string           `json:"atualizado_em"`
}

type ListQuery struct {
	Search string
	Status string
}
