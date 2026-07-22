package entity

type CoverInput struct {
	Base64       string `json:"base64"`
	Mime         string `json:"mime"`
	Largura      int    `json:"largura"`
	Altura       int    `json:"altura"`
	TamanhoBytes int    `json:"tamanho_bytes"`
	HashSHA256   string `json:"hash_sha256"`
}

type ValidationRequest struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	ISBN         string `json:"isbn"`
	CodigoBarra  string `json:"codigo_barra"`
}

type ValidationResponse struct {
	Valid       bool     `json:"valid"`
	Status      string   `json:"status,omitempty"`
	ISBN        string   `json:"isbn,omitempty"`
	CodigoBarra string   `json:"codigo_barra,omitempty"`
	Disponivel  bool     `json:"disponivel"`
	Erros       []string `json:"erros,omitempty"`
}

type CreateRequest struct {
	AutorID                string      `json:"autor_id"`
	Titulo                 string      `json:"titulo"`
	Subtitulo              string      `json:"subtitulo"`
	Sinopse                string      `json:"sinopse"`
	ISBN                   string      `json:"isbn"`
	CodigoBarra            string      `json:"codigo_barra"`
	Status                 string      `json:"status"`
	Formato                string      `json:"formato"`
	PossuiFormatoFisico    bool        `json:"possui_formato_fisico"`
	PossuiFormatoDigital   bool        `json:"possui_formato_digital"`
	Edicao                 string      `json:"edicao"`
	Idioma                 string      `json:"idioma"`
	NumeroPaginas          *int        `json:"numero_paginas"`
	Genero                 string      `json:"genero"`
	PrecoVenda             *float64    `json:"preco_venda"`
	PrecoVendaFisico       *float64    `json:"preco_venda_fisico"`
	PrecoVendaDigital      *float64    `json:"preco_venda_digital"`
	CanalVendaDigital      string      `json:"canal_venda_digital"`
	URLCompraDigital       string      `json:"url_compra_digital"`
	CustoImpressao         *float64    `json:"custo_impressao"`
	VendaInfinita          bool        `json:"venda_infinita"`
	ControlarEstoque       bool        `json:"controlar_estoque"`
	EstoqueDisponivel      *int        `json:"estoque_disponivel"`
	EstoqueMinimo          *int        `json:"estoque_minimo"`
	PesoGramas             *int        `json:"peso_gramas"`
	LarguraCm              *float64    `json:"largura_cm"`
	AlturaCm               *float64    `json:"altura_cm"`
	ProfundidadeCm         *float64    `json:"profundidade_cm"`
	TipoCapa               string      `json:"tipo_capa"`
	PossuiBox              bool        `json:"possui_box"`
	DetalhesEdicao         string      `json:"detalhes_edicao"`
	DataPublicacaoPrevista string      `json:"data_publicacao_prevista"`
	DataPublicacao         string      `json:"data_publicacao"`
	Capa                   *CoverInput `json:"capa"`
	Ativo                  bool        `json:"ativo"`
}

type UpdateRequest = CreateRequest

type PersistInput struct {
	ID                     string
	AutorID                string
	Titulo                 string
	Subtitulo              *string
	Sinopse                *string
	ISBN                   *string
	CodigoBarra            *string
	Status                 string
	Formato                string
	PossuiFormatoFisico    bool
	PossuiFormatoDigital   bool
	Edicao                 *string
	Idioma                 *string
	NumeroPaginas          *int
	Genero                 *string
	PrecoVenda             float64
	PrecoVendaFisico       float64
	PrecoVendaDigital      float64
	CanalVendaDigital      *string
	URLCompraDigital       *string
	CustoImpressao         float64
	VendaInfinita          bool
	ControlarEstoque       bool
	EstoqueDisponivel      int
	EstoqueMinimo          int
	PesoGramas             *int
	LarguraCm              *float64
	AlturaCm               *float64
	ProfundidadeCm         *float64
	TipoCapa               *string
	PossuiBox              bool
	DetalhesEdicao         *string
	DataPublicacaoPrevista *string
	DataPublicacao         *string
	Capa                   *CoverInput
	Ativo                  bool
}

type ListQuery struct {
	Search string
	Status string
	AutorID string
}

type ListItem struct {
	ID                     string      `json:"id"`
	AutorID                string      `json:"autor_id"`
	AutorNome              string      `json:"autor_nome"`
	Titulo                 string      `json:"titulo"`
	Subtitulo              *string     `json:"subtitulo,omitempty"`
	ISBN                   *string     `json:"isbn,omitempty"`
	CodigoBarra            *string     `json:"codigo_barra,omitempty"`
	Status                 string      `json:"status"`
	Formato                string      `json:"formato"`
	PossuiFormatoFisico    bool        `json:"possui_formato_fisico"`
	PossuiFormatoDigital   bool        `json:"possui_formato_digital"`
	Genero                 *string     `json:"genero,omitempty"`
	PrecoVenda             float64     `json:"preco_venda"`
	PrecoVendaFisico       float64     `json:"preco_venda_fisico"`
	PrecoVendaDigital      float64     `json:"preco_venda_digital"`
	CanalVendaDigital      *string     `json:"canal_venda_digital,omitempty"`
	URLCompraDigital       *string     `json:"url_compra_digital,omitempty"`
	VendaInfinita          bool        `json:"venda_infinita"`
	ControlarEstoque       bool        `json:"controlar_estoque"`
	EstoqueDisponivel      int         `json:"estoque_disponivel"`
	EstoqueMinimo          int         `json:"estoque_minimo"`
	Ativo                  bool        `json:"ativo"`
	PossuiCapa             bool        `json:"possui_capa"`
	Capa                   *CoverInput `json:"capa,omitempty"`
	DataPublicacao         *string     `json:"data_publicacao,omitempty"`
	DataPublicacaoPrevista *string     `json:"data_publicacao_prevista,omitempty"`
	CriadoEm               string      `json:"criado_em"`
	AtualizadoEm           string      `json:"atualizado_em"`
}

type DetailResponse struct {
	ID                     string      `json:"id"`
	AutorID                string      `json:"autor_id"`
	AutorNome              string      `json:"autor_nome"`
	Titulo                 string      `json:"titulo"`
	Subtitulo              *string     `json:"subtitulo,omitempty"`
	Sinopse                *string     `json:"sinopse,omitempty"`
	ISBN                   *string     `json:"isbn,omitempty"`
	CodigoBarra            *string     `json:"codigo_barra,omitempty"`
	Status                 string      `json:"status"`
	Formato                string      `json:"formato"`
	PossuiFormatoFisico    bool        `json:"possui_formato_fisico"`
	PossuiFormatoDigital   bool        `json:"possui_formato_digital"`
	Edicao                 *string     `json:"edicao,omitempty"`
	Idioma                 *string     `json:"idioma,omitempty"`
	NumeroPaginas          *int        `json:"numero_paginas,omitempty"`
	Genero                 *string     `json:"genero,omitempty"`
	PrecoVenda             float64     `json:"preco_venda"`
	PrecoVendaFisico       float64     `json:"preco_venda_fisico"`
	PrecoVendaDigital      float64     `json:"preco_venda_digital"`
	CanalVendaDigital      *string     `json:"canal_venda_digital,omitempty"`
	URLCompraDigital       *string     `json:"url_compra_digital,omitempty"`
	CustoImpressao         float64     `json:"custo_impressao"`
	VendaInfinita          bool        `json:"venda_infinita"`
	ControlarEstoque       bool        `json:"controlar_estoque"`
	EstoqueDisponivel      int         `json:"estoque_disponivel"`
	EstoqueReservado       int         `json:"estoque_reservado"`
	EstoqueMinimo          int         `json:"estoque_minimo"`
	PesoGramas             *int        `json:"peso_gramas,omitempty"`
	LarguraCm              *float64    `json:"largura_cm,omitempty"`
	AlturaCm               *float64    `json:"altura_cm,omitempty"`
	ProfundidadeCm         *float64    `json:"profundidade_cm,omitempty"`
	TipoCapa               *string     `json:"tipo_capa,omitempty"`
	PossuiBox              bool        `json:"possui_box"`
	DetalhesEdicao         *string     `json:"detalhes_edicao,omitempty"`
	DataPublicacao         *string     `json:"data_publicacao,omitempty"`
	DataPublicacaoPrevista *string     `json:"data_publicacao_prevista,omitempty"`
	Capa                   *CoverInput `json:"capa,omitempty"`
	Ativo                  bool        `json:"ativo"`
	CriadoEm               string      `json:"criado_em"`
	AtualizadoEm           string      `json:"atualizado_em"`
}

type StockMovementRequest struct {
	Tipo       string `json:"tipo"`
	Quantidade int    `json:"quantidade"`
	Motivo     string `json:"motivo"`
	Observacao string `json:"observacao"`
}
