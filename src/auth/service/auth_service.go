package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"sistema-editorial/editora/backend/src/auth/entity"
	usuariosentity "sistema-editorial/editora/backend/src/usuarios/entity"
)

type repository interface {
	FindByLogin(ctx context.Context, login string) (entity.LoginUserRecord, error)
	FindByID(ctx context.Context, userID string) (entity.AuthenticatedUser, error)
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
}

type Claims struct {
	Papeis             []string `json:"papeis"`
	NomeCompleto       string   `json:"nome_completo"`
	Email              string   `json:"email"`
	PrecisaTrocarSenha bool     `json:"precisa_trocar_senha"`
	jwt.RegisteredClaims
}

type Service struct {
	repo      repository
	jwtSecret string
}

func NewService(jwtSecret string, repo repository) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Login(ctx context.Context, input entity.LoginRequest) (entity.LoginResponse, error) {
	login := normalizeLogin(input.Login)
	if login == "" || strings.TrimSpace(input.Senha) == "" {
		return entity.LoginResponse{}, errors.New("login e senha sao obrigatorios")
	}

	user, err := s.repo.FindByLogin(ctx, login)
	if err != nil {
		return entity.LoginResponse{}, errors.New("credenciais invalidas")
	}

	switch strings.TrimSpace(strings.ToUpper(user.StatusAcesso)) {
	case usuariosentity.StatusAcessoBloqueado:
		return entity.LoginResponse{}, errors.New("usuario bloqueado")
	case usuariosentity.StatusAcessoPendenteAprovacao:
		return entity.LoginResponse{}, errors.New("usuario pendente de aprovacao")
	}

	if !user.Ativo {
		return entity.LoginResponse{}, errors.New("usuario bloqueado")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.SenhaHash), []byte(strings.TrimSpace(input.Senha))); err != nil {
		return entity.LoginResponse{}, errors.New("credenciais invalidas")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return entity.LoginResponse{}, err
	}

	return entity.LoginResponse{
		Token:              token,
		UserID:             user.ID,
		NomeCompleto:       user.NomeCompleto,
		Email:              user.Email,
		Papeis:             user.Papeis,
		PrecisaTrocarSenha: user.PrecisaTrocarSenha,
	}, nil
}

func (s *Service) Me(ctx context.Context, userID string) (entity.CurrentUserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return entity.CurrentUserResponse{}, err
	}

	return entity.CurrentUserResponse{
		ID:                 user.ID,
		CPF:                user.CPF,
		Email:              user.Email,
		NomeCompleto:       user.NomeCompleto,
		Papeis:             user.Papeis,
		PrecisaTrocarSenha: user.PrecisaTrocarSenha,
		StatusAcesso:       user.StatusAcesso,
		ClienteAtivo:       user.ClienteAtivo,
	}, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID string, input entity.ChangePasswordRequest) error {
	if len(strings.TrimSpace(input.NovaSenha)) < 5 {
		return errors.New("nova_senha deve ter pelo menos 5 caracteres")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	loginUser, err := s.repo.FindByLogin(ctx, user.Email)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(loginUser.SenhaHash), []byte(strings.TrimSpace(input.SenhaAtual))); err != nil {
		return errors.New("senha_atual invalida")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(input.NovaSenha)), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, string(passwordHash))
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo de assinatura invalido")
		}

		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token invalido")
	}

	return claims, nil
}

func (s *Service) generateToken(user entity.LoginUserRecord) (string, error) {
	now := time.Now()
	claims := Claims{
		Papeis:             user.Papeis,
		NomeCompleto:       user.NomeCompleto,
		Email:              user.Email,
		PrecisaTrocarSenha: user.PrecisaTrocarSenha,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    "aura-editora",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func normalizeLogin(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if strings.Contains(trimmed, "@") {
		return trimmed
	}

	return strings.NewReplacer(".", "", "-", "", "/", "", " ", "", "(", "", ")", "", "+", "").Replace(trimmed)
}
