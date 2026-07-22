package entity

type LoginRequest struct {
	Login string `json:"login"`
	Senha string `json:"senha"`
}

type LoginResponse struct {
	Token              string   `json:"token"`
	UserID             string   `json:"user_id"`
	NomeCompleto       string   `json:"nome_completo"`
	Email              string   `json:"email"`
	Papeis             []string `json:"papeis"`
	PrecisaTrocarSenha bool     `json:"precisa_trocar_senha"`
}

type ChangePasswordRequest struct {
	SenhaAtual string `json:"senha_atual"`
	NovaSenha  string `json:"nova_senha"`
}

type CurrentUserResponse struct {
	ID                 string   `json:"id"`
	CPF                string   `json:"cpf"`
	Email              string   `json:"email"`
	NomeCompleto       string   `json:"nome_completo"`
	Papeis             []string `json:"papeis"`
	PrecisaTrocarSenha bool     `json:"precisa_trocar_senha"`
	StatusAcesso       string   `json:"status_acesso"`
	ClienteAtivo       bool     `json:"cliente_ativo"`
}

type AuthenticatedUser struct {
	ID                 string
	CPF                string
	Email              string
	NomeCompleto       string
	Papeis             []string
	PrecisaTrocarSenha bool
	StatusAcesso       string
	ClienteAtivo       bool
}

type LoginUserRecord struct {
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
