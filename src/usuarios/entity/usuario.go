package entity

const (
	StatusAcessoAtivo             = "ATIVO"
	StatusAcessoPendenteAprovacao = "PENDENTE_APROVACAO"
	StatusAcessoBloqueado         = "BLOQUEADO"
)

type FotoInput struct {
	Base64       string `json:"base64"`
	Mime         string `json:"mime"`
	Largura      int    `json:"largura"`
	Altura       int    `json:"altura"`
	TamanhoBytes int    `json:"tamanho_bytes"`
	HashSHA256   string `json:"hash_sha256"`
}

type EnderecoInput struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Cidade      string `json:"cidade"`
	UF          string `json:"uf"`
	Pais        string `json:"pais"`
}

type CreateRequest struct {
	CPF               string         `json:"cpf"`
	Email             string         `json:"email"`
	NomeCompleto      string         `json:"nome_completo"`
	Foto              FotoInput      `json:"foto"`
	Descricao         string         `json:"descricao"`
	Pseudonimo        string         `json:"pseudonimo"`
	EnderecoPrincipal *EnderecoInput `json:"endereco_principal"`
	WhatsApp          string         `json:"whatsapp"`
	DataNascimento    string         `json:"data_nascimento"`
	Nacionalidade     string         `json:"nacionalidade"`
	Senha             string         `json:"senha"`
	Papeis            []string       `json:"papeis"`
	OrigemCadastro    string         `json:"origem_cadastro"`
}

type UpdateRequest struct {
	Email             string         `json:"email"`
	NomeCompleto      string         `json:"nome_completo"`
	Foto              FotoInput      `json:"foto"`
	Descricao         string         `json:"descricao"`
	Pseudonimo        string         `json:"pseudonimo"`
	EnderecoPrincipal *EnderecoInput `json:"endereco_principal"`
	WhatsApp          string         `json:"whatsapp"`
	DataNascimento    string         `json:"data_nascimento"`
	Nacionalidade     string         `json:"nacionalidade"`
	Senha             string         `json:"senha"`
	Papeis            []string       `json:"papeis"`
}

type CreateResponse struct {
	ID                 string   `json:"id"`
	CPF                string   `json:"cpf"`
	Email              string   `json:"email"`
	NomeCompleto       string   `json:"nome_completo"`
	Pseudonimo         string   `json:"pseudonimo,omitempty"`
	Papeis             []string `json:"papeis"`
	PrecisaTrocarSenha bool     `json:"precisa_trocar_senha"`
}

type ListQuery struct {
	Search   string
	Role     string
	Page     int
	PageSize int
}

type ListItem struct {
	ID                string         `json:"id"`
	CPF               string         `json:"cpf"`
	Email             string         `json:"email"`
	NomeCompleto      string         `json:"nome_completo"`
	Pseudonimo        string         `json:"pseudonimo,omitempty"`
	WhatsApp          string         `json:"whatsapp"`
	Papeis            []string       `json:"papeis"`
	OrigemCadastro    string         `json:"origem_cadastro"`
	Status            string         `json:"status"`
	StatusCodigo      string         `json:"status_codigo"`
	ClienteAtivo      bool           `json:"cliente_ativo"`
	Foto              *FotoInput     `json:"foto,omitempty"`
	Descricao         string         `json:"descricao,omitempty"`
	EnderecoPrincipal *EnderecoInput `json:"endereco_principal,omitempty"`
	DataNascimento    string         `json:"data_nascimento,omitempty"`
	Nacionalidade     string         `json:"nacionalidade,omitempty"`
}

type ListResponse struct {
	Items      []ListItem `json:"items"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
}

type User struct {
	ID                 string
	CPF                string
	Email              string
	NomeCompleto       string
	SenhaHash          string
	Papeis             []string
	PrecisaTrocarSenha bool
	Ativo              bool
	StatusAcesso       string
	ClienteAtivo       bool
}

type PersistInput struct {
	ID                 string
	CPF                string
	Email              string
	NomeCompleto       string
	Foto               FotoInput
	Descricao          string
	Pseudonimo         string
	EnderecoPrincipal  *EnderecoInput
	WhatsApp           string
	DataNascimento     string
	Nacionalidade      string
	SenhaHash          string
	Papeis             []string
	OrigemCadastro     string
	PrecisaTrocarSenha bool
	UpdatePassword     bool
	StatusAcesso       string
	ClienteAtivo       bool
}

type StatusActionResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	StatusCodigo string `json:"status_codigo"`
	ClienteAtivo bool   `json:"cliente_ativo"`
}

type ResetPasswordResponse struct {
	ID                 string `json:"id"`
	SenhaTemporaria    string `json:"senha_temporaria"`
	PrecisaTrocarSenha bool   `json:"precisa_trocar_senha"`
}
