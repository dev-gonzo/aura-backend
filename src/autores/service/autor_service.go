package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"sistema-editorial/editora/backend/src/autores/entity"
)

type repository interface {
	Create(ctx context.Context, input entity.PersistInput) (string, error)
	Update(ctx context.Context, input entity.PersistInput) error
	List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error)
	FindByID(ctx context.Context, id string) (entity.DetailResponse, error)
	ExistsByEmail(ctx context.Context, email string, excludeID string) (bool, error)
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error) {
	return s.repo.List(ctx, query)
}

func (s *Service) FindByID(ctx context.Context, id string) (entity.DetailResponse, error) {
	if strings.TrimSpace(id) == "" {
		return entity.DetailResponse{}, ValidationError{Message: "id do autor e obrigatorio"}
	}

	return s.repo.FindByID(ctx, strings.TrimSpace(id))
}

func (s *Service) Create(ctx context.Context, request entity.CreateRequest) (string, error) {
	input, err := s.normalizePersistInput(ctx, "", request)
	if err != nil {
		return "", err
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id string, request entity.UpdateRequest) error {
	input, err := s.normalizePersistInput(ctx, id, request)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, input)
}

func (s *Service) normalizePersistInput(ctx context.Context, id string, request entity.CreateRequest) (entity.PersistInput, error) {
	nomeCompleto := strings.TrimSpace(request.NomeCompleto)
	if nomeCompleto == "" {
		return entity.PersistInput{}, ValidationError{Message: "nome do autor e obrigatorio"}
	}

	status := normalizeStatus(request.Status)
	if status == "" {
		status = "ATIVO"
	}

	if status != "ATIVO" && status != "INATIVO" {
		return entity.PersistInput{}, ValidationError{Message: "status do autor invalido"}
	}

	if err := validatePhoto(request.Foto); err != nil {
		return entity.PersistInput{}, err
	}

	email := normalizeOptionalString(request.Email)
	if email != nil {
		exists, err := s.repo.ExistsByEmail(ctx, *email, strings.TrimSpace(id))
		if err != nil {
			return entity.PersistInput{}, fmt.Errorf("erro ao validar email do autor: %w", err)
		}
		if exists {
			return entity.PersistInput{}, ValidationError{Message: "email do autor ja cadastrado"}
		}
	}

	return entity.PersistInput{
		ID:                 strings.TrimSpace(id),
		UsuarioID:          normalizeOptionalString(request.UsuarioID),
		NomeCompleto:       nomeCompleto,
		NomePublico:        normalizeOptionalString(request.NomePublico),
		Email:              email,
		EmailPrivado:       request.EmailPrivado,
		Whatsapp:           normalizeOptionalString(request.Whatsapp),
		WhatsappPrivado:    request.WhatsappPrivado,
		Instagram:          normalizeOptionalString(request.Instagram),
		InstagramPrivado:   request.InstagramPrivado,
		Wattpad:            normalizeOptionalString(request.Wattpad),
		WattpadPrivado:     request.WattpadPrivado,
		Facebook:           normalizeOptionalString(request.Facebook),
		FacebookPrivado:    request.FacebookPrivado,
		XTwitter:           normalizeOptionalString(request.XTwitter),
		XTwitterPrivado:    request.XTwitterPrivado,
		Tiktok:             normalizeOptionalString(request.Tiktok),
		TiktokPrivado:      request.TiktokPrivado,
		Youtube:            normalizeOptionalString(request.Youtube),
		YoutubePrivado:     request.YoutubePrivado,
		Linkedin:           normalizeOptionalString(request.Linkedin),
		LinkedinPrivado:    request.LinkedinPrivado,
		Nacionalidade:      normalizeOptionalString(request.Nacionalidade),
		Biografia:          normalizeOptionalString(request.Biografia),
		Foto:               request.Foto,
		Status:             status,
	}, nil
}

func normalizeStatus(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validatePhoto(photo *entity.PhotoInput) error {
	if photo == nil {
		return nil
	}

	if strings.TrimSpace(photo.Base64) == "" {
		return ValidationError{Message: "a foto do autor esta invalida"}
	}
	if strings.TrimSpace(photo.Mime) != "image/webp" {
		return ValidationError{Message: "a foto do autor deve estar em image/webp"}
	}
	if photo.Largura != 1024 || photo.Altura != 1024 {
		return ValidationError{Message: "a foto do autor deve ter exatamente 1024x1024"}
	}
	if photo.TamanhoBytes <= 0 || photo.TamanhoBytes > 150*1024 {
		return ValidationError{Message: "a foto do autor deve ter no maximo 150 KB"}
	}
	if strings.TrimSpace(photo.HashSHA256) == "" {
		return ValidationError{Message: "o hash da foto do autor e obrigatorio"}
	}

	return nil
}

func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
