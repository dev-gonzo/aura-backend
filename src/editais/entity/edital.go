package entity

import "io"

const (
	StatusRascunho  = "RASCUNHO"
	StatusAgendado  = "AGENDADO"
	StatusPublicado = "PUBLICADO"
	StatusEncerrado = "ENCERRADO"
	StatusCancelado = "CANCELADO"
)

type UploadRequest struct {
	FileName    string
	ContentType string
	Size        int64
	Body        io.Reader
}

type UploadResponse struct {
	FileName    string `json:"nome_arquivo"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"tamanho_bytes"`
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	URL         string `json:"url"`
}

type CapaInput struct {
	Base64       string `json:"base64"`
	Mime         string `json:"mime"`
	Largura      int    `json:"largura"`
	Altura       int    `json:"altura"`
	TamanhoBytes int    `json:"tamanho_bytes"`
	HashSHA256   string `json:"hash_sha256"`
}

type AnexoInput struct {
	NomeArquivo  string `json:"nome_arquivo"`
	ContentType  string `json:"content_type"`
	TamanhoBytes int64  `json:"tamanho_bytes"`
	Bucket       string `json:"bucket"`
	Key          string `json:"key"`
	URL          string `json:"url"`
}

type CreateRequest struct {
	Capa                   CapaInput   `json:"capa"`
	Titulo                 string      `json:"titulo"`
	Descricao              string      `json:"descricao"`
	Anexo                  *AnexoInput `json:"anexo"`
	TaxaInscricao          *float64    `json:"taxa_inscricao"`
	TaxaPublicacao         *float64    `json:"taxa_publicacao"`
	Status                 string      `json:"status"`
	DataInicio             string      `json:"data_inicio"`
	DataFim                string      `json:"data_fim"`
	TotalVagas             *int        `json:"total_vagas"`
	DataPrevistaPublicacao string      `json:"data_prevista_publicacao"`
}

type UpdateRequest = CreateRequest

type PersistInput struct {
	ID                     string
	Capa                   CapaInput
	Titulo                 string
	Descricao              string
	Anexo                  *AnexoInput
	TaxaInscricao          *float64
	TaxaPublicacao         *float64
	Status                 string
	DataInicio             *string
	DataFim                *string
	TotalVagas             *int
	DataPrevistaPublicacao *string
}

type ListQuery struct {
	Search string
	Status string
}

type ListItem struct {
	ID                     string `json:"id"`
	Titulo                 string `json:"titulo"`
	Descricao              string `json:"descricao"`
	Status                 string `json:"status"`
	DataInicio             string `json:"data_inicio,omitempty"`
	DataFim                string `json:"data_fim,omitempty"`
	TotalVagas             *int   `json:"total_vagas,omitempty"`
	DataPrevistaPublicacao string `json:"data_prevista_publicacao,omitempty"`
	TemCapa                bool   `json:"tem_capa"`
	TemAnexo               bool   `json:"tem_anexo"`
	AnexoNomeArquivo       string `json:"anexo_nome_arquivo,omitempty"`
	CriadoEm               string `json:"criado_em"`
	AtualizadoEm           string `json:"atualizado_em"`
}

type DetailResponse struct {
	ID                     string      `json:"id"`
	Capa                   CapaInput   `json:"capa"`
	Titulo                 string      `json:"titulo"`
	Descricao              string      `json:"descricao"`
	Anexo                  *AnexoInput `json:"anexo,omitempty"`
	TaxaInscricao          *float64    `json:"taxa_inscricao,omitempty"`
	TaxaPublicacao         *float64    `json:"taxa_publicacao,omitempty"`
	Status                 string      `json:"status"`
	DataInicio             string      `json:"data_inicio,omitempty"`
	DataFim                string      `json:"data_fim,omitempty"`
	TotalVagas             *int        `json:"total_vagas,omitempty"`
	DataPrevistaPublicacao string      `json:"data_prevista_publicacao,omitempty"`
	CriadoEm               string      `json:"criado_em"`
	AtualizadoEm           string      `json:"atualizado_em"`
}
