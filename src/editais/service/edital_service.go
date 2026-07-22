package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"sistema-editorial/editora/backend/src/editais/entity"
)

var invalidFileNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
var allowedStatuses = []string{
	entity.StatusRascunho,
	entity.StatusAgendado,
	entity.StatusPublicado,
	entity.StatusEncerrado,
	entity.StatusCancelado,
}

type objectStorage interface {
	Upload(
		ctx context.Context,
		key string,
		body io.Reader,
		contentType string,
		cacheControl string,
		contentLength int64,
	) (StoredObject, error)
	Delete(ctx context.Context, key string) error
}

type repository interface {
	Create(ctx context.Context, input entity.PersistInput) (string, error)
	Update(ctx context.Context, input entity.PersistInput) error
	List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error)
	FindByID(ctx context.Context, id string) (entity.DetailResponse, error)
}

type StoredObject struct {
	Bucket string
	Key    string
	URL    string
}

type ValidationError struct {
	Message string
}

func (error ValidationError) Error() string {
	return error.Message
}

func IsValidationError(err error) bool {
	var target ValidationError
	return errors.As(err, &target)
}

type UploadService struct {
	storage objectStorage
	repo    repository
}

func NewUploadService(storage objectStorage, repo repository) *UploadService {
	return &UploadService{
		storage: storage,
		repo:    repo,
	}
}

func (service *UploadService) Upload(ctx context.Context, input entity.UploadRequest) (entity.UploadResponse, error) {
	if service.storage == nil {
		return entity.UploadResponse{}, ValidationError{Message: "storage do edital nao configurado"}
	}

	fileName := strings.TrimSpace(input.FileName)
	if fileName == "" {
		return entity.UploadResponse{}, ValidationError{Message: "selecione um arquivo"}
	}

	if input.Body == nil {
		return entity.UploadResponse{}, ValidationError{Message: "arquivo invalido"}
	}

	if input.Size <= 0 {
		return entity.UploadResponse{}, ValidationError{Message: "arquivo vazio"}
	}

	key := buildObjectKey(fileName)
	object, err := service.storage.Upload(
		ctx,
		key,
		input.Body,
		strings.TrimSpace(input.ContentType),
		"public, max-age=31536000",
		input.Size,
	)
	if err != nil {
		return entity.UploadResponse{}, err
	}

	return entity.UploadResponse{
		FileName:    fileName,
		ContentType: strings.TrimSpace(input.ContentType),
		Size:        input.Size,
		Bucket:      object.Bucket,
		Key:         object.Key,
		URL:         object.URL,
	}, nil
}

func (service *UploadService) List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error) {
	if service.repo == nil {
		return nil, ValidationError{Message: "repositorio do edital nao configurado"}
	}

	query.Search = strings.TrimSpace(query.Search)
	query.Status = normalizeStatus(query.Status)
	if query.Status != "" && !slices.Contains(allowedStatuses, query.Status) {
		return nil, ValidationError{Message: "status do edital invalido"}
	}

	return service.repo.List(ctx, query)
}

func (service *UploadService) FindByID(ctx context.Context, id string) (entity.DetailResponse, error) {
	if service.repo == nil {
		return entity.DetailResponse{}, ValidationError{Message: "repositorio do edital nao configurado"}
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return entity.DetailResponse{}, ValidationError{Message: "id do edital e obrigatorio"}
	}

	return service.repo.FindByID(ctx, id)
}

func (service *UploadService) Create(ctx context.Context, input entity.CreateRequest) (entity.DetailResponse, error) {
	if service.repo == nil {
		return entity.DetailResponse{}, ValidationError{Message: "repositorio do edital nao configurado"}
	}

	persistInput, err := normalizeAndValidatePersistInput(input)
	if err != nil {
		return entity.DetailResponse{}, err
	}

	id, err := service.repo.Create(ctx, persistInput)
	if err != nil {
		service.cleanupUploadedAttachment(ctx, persistInput.Anexo)
		return entity.DetailResponse{}, err
	}

	return service.repo.FindByID(ctx, id)
}

func (service *UploadService) Update(ctx context.Context, id string, input entity.UpdateRequest) (entity.DetailResponse, error) {
	if service.repo == nil {
		return entity.DetailResponse{}, ValidationError{Message: "repositorio do edital nao configurado"}
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return entity.DetailResponse{}, ValidationError{Message: "id do edital e obrigatorio"}
	}

	current, err := service.repo.FindByID(ctx, id)
	if err != nil {
		return entity.DetailResponse{}, err
	}

	persistInput, err := normalizeAndValidatePersistInput(input)
	if err != nil {
		return entity.DetailResponse{}, err
	}

	persistInput.ID = id
	if err := service.repo.Update(ctx, persistInput); err != nil {
		if hasAttachmentChanged(current.Anexo, persistInput.Anexo) {
			service.cleanupUploadedAttachment(ctx, persistInput.Anexo)
		}

		return entity.DetailResponse{}, err
	}

	if hasAttachmentChanged(current.Anexo, persistInput.Anexo) {
		if err := service.deleteStoredAttachment(ctx, current.Anexo); err != nil {
			return entity.DetailResponse{}, err
		}
	}

	return service.repo.FindByID(ctx, id)
}

func buildObjectKey(fileName string) string {
	trimmedName := strings.TrimSpace(fileName)
	extension := strings.ToLower(filepath.Ext(trimmedName))
	baseName := strings.TrimSuffix(trimmedName, extension)
	baseName = invalidFileNamePattern.ReplaceAllString(baseName, "-")
	baseName = strings.Trim(baseName, "-._ ")
	if baseName == "" {
		baseName = "arquivo"
	}

	if extension == "" {
		extension = ".bin"
	}

	timestamp := time.Now().UTC().Format("20060102-150405")
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	datePath := time.Now().UTC().Format("2006/01/02")

	return fmt.Sprintf("editais/testes/%s/%s-%s%s", datePath, timestamp, uniqueSuffix, extensionWithBase(baseName, extension))
}

func extensionWithBase(baseName string, extension string) string {
	return "-" + baseName + extension
}

func normalizeAndValidatePersistInput(input entity.CreateRequest) (entity.PersistInput, error) {
	titulo := strings.TrimSpace(input.Titulo)
	descricao := strings.TrimSpace(input.Descricao)
	status := normalizeStatus(input.Status)

	if err := validateCover(input.Capa); err != nil {
		return entity.PersistInput{}, err
	}

	if titulo == "" {
		return entity.PersistInput{}, ValidationError{Message: "titulo do edital e obrigatorio"}
	}

	if len(titulo) > 255 {
		return entity.PersistInput{}, ValidationError{Message: "titulo do edital deve ter no maximo 255 caracteres"}
	}

	if descricao == "" {
		return entity.PersistInput{}, ValidationError{Message: "descricao do edital e obrigatoria"}
	}

	if status == "" {
		status = entity.StatusRascunho
	}

	if !slices.Contains(allowedStatuses, status) {
		return entity.PersistInput{}, ValidationError{Message: "status do edital invalido"}
	}

	dataInicio, err := normalizeOptionalDate(input.DataInicio, "data de inicio")
	if err != nil {
		return entity.PersistInput{}, err
	}

	dataFim, err := normalizeOptionalDate(input.DataFim, "data de fim")
	if err != nil {
		return entity.PersistInput{}, err
	}

	dataPrevistaPublicacao, err := normalizeOptionalDate(input.DataPrevistaPublicacao, "data prevista de publicacao")
	if err != nil {
		return entity.PersistInput{}, err
	}

	if dataInicio != nil && dataFim != nil && *dataFim < *dataInicio {
		return entity.PersistInput{}, ValidationError{Message: "a data de fim deve ser maior ou igual a data de inicio"}
	}

	if input.TotalVagas != nil && *input.TotalVagas <= 0 {
		return entity.PersistInput{}, ValidationError{Message: "o total de vagas deve ser maior que zero"}
	}

	taxaInscricao, err := normalizeMoney(input.TaxaInscricao, "taxa de inscricao")
	if err != nil {
		return entity.PersistInput{}, err
	}

	taxaPublicacao, err := normalizeMoney(input.TaxaPublicacao, "taxa de publicacao")
	if err != nil {
		return entity.PersistInput{}, err
	}

	anexo, err := normalizeAttachment(input.Anexo)
	if err != nil {
		return entity.PersistInput{}, err
	}

	capa := input.Capa
	if strings.TrimSpace(capa.HashSHA256) == "" {
		sum := sha256.Sum256([]byte(capa.Base64))
		capa.HashSHA256 = hex.EncodeToString(sum[:])
	}

	return entity.PersistInput{
		Capa:                   capa,
		Titulo:                 titulo,
		Descricao:              descricao,
		Anexo:                  anexo,
		TaxaInscricao:          taxaInscricao,
		TaxaPublicacao:         taxaPublicacao,
		Status:                 status,
		DataInicio:             dataInicio,
		DataFim:                dataFim,
		TotalVagas:             input.TotalVagas,
		DataPrevistaPublicacao: dataPrevistaPublicacao,
	}, nil
}

func validateCover(capa entity.CapaInput) error {
	if strings.TrimSpace(capa.Base64) == "" {
		return ValidationError{Message: "arte de capa e obrigatoria"}
	}

	if strings.TrimSpace(capa.Mime) != "image/webp" {
		return ValidationError{Message: "a arte de capa deve estar em image/webp"}
	}

	if capa.Largura != 1024 || capa.Altura != 1024 {
		return ValidationError{Message: "a arte de capa deve ter exatamente 1024x1024"}
	}

	if capa.TamanhoBytes <= 0 || capa.TamanhoBytes > 150*1024 {
		return ValidationError{Message: "a arte de capa final deve ter no maximo 150 KB"}
	}

	if _, err := base64.StdEncoding.DecodeString(capa.Base64); err != nil {
		return ValidationError{Message: "a arte de capa em base64 e invalida"}
	}

	return nil
}

func normalizeOptionalDate(value string, fieldLabel string) (*string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return nil, ValidationError{Message: fmt.Sprintf("%s deve estar no formato yyyy-mm-dd", fieldLabel)}
	}

	return &trimmed, nil
}

func normalizeMoney(value *float64, fieldLabel string) (*float64, error) {
	if value == nil {
		return nil, nil
	}

	if *value < 0 {
		return nil, ValidationError{Message: fmt.Sprintf("%s nao pode ser negativa", fieldLabel)}
	}

	normalized := float64(int((*value)*100+0.5)) / 100
	return &normalized, nil
}

func normalizeAttachment(anexo *entity.AnexoInput) (*entity.AnexoInput, error) {
	if anexo == nil {
		return nil, nil
	}

	normalized := &entity.AnexoInput{
		NomeArquivo:  strings.TrimSpace(anexo.NomeArquivo),
		ContentType:  strings.TrimSpace(anexo.ContentType),
		TamanhoBytes: anexo.TamanhoBytes,
		Bucket:       strings.TrimSpace(anexo.Bucket),
		Key:          strings.TrimSpace(anexo.Key),
		URL:          strings.TrimSpace(anexo.URL),
	}

	if normalized.NomeArquivo == "" &&
		normalized.ContentType == "" &&
		normalized.TamanhoBytes == 0 &&
		normalized.Bucket == "" &&
		normalized.Key == "" &&
		normalized.URL == "" {
		return nil, nil
	}

	if normalized.NomeArquivo == "" ||
		normalized.ContentType == "" ||
		normalized.TamanhoBytes <= 0 ||
		normalized.Bucket == "" ||
		normalized.Key == "" ||
		normalized.URL == "" {
		return nil, ValidationError{Message: "o anexo do edital esta incompleto"}
	}

	return normalized, nil
}

func normalizeStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

func hasAttachmentChanged(current *entity.AnexoInput, next *entity.AnexoInput) bool {
	currentKey := attachmentKey(current)
	nextKey := attachmentKey(next)

	return currentKey != nextKey
}

func attachmentKey(anexo *entity.AnexoInput) string {
	if anexo == nil {
		return ""
	}

	return strings.TrimSpace(anexo.Key)
}

func (service *UploadService) cleanupUploadedAttachment(ctx context.Context, anexo *entity.AnexoInput) {
	_ = service.deleteStoredAttachment(ctx, anexo)
}

func (service *UploadService) deleteStoredAttachment(ctx context.Context, anexo *entity.AnexoInput) error {
	if service.storage == nil || anexo == nil {
		return nil
	}

	key := strings.TrimSpace(anexo.Key)
	if key == "" {
		return nil
	}

	return service.storage.Delete(ctx, key)
}
