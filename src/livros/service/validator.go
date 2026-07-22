package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"sistema-editorial/editora/backend/src/livros/entity"
)

type repository interface {
	IsCatalogIdentifierAvailable(ctx context.Context, normalizedISBN string, codigoBarra string, excludeID string) (bool, error)
	Create(ctx context.Context, input entity.PersistInput) (string, error)
	Update(ctx context.Context, input entity.PersistInput) error
	List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error)
	FindByID(ctx context.Context, id string) (entity.DetailResponse, error)
	RegisterStockMovement(ctx context.Context, livroID string, request entity.StockMovementRequest) error
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
		return entity.DetailResponse{}, ValidationError{Message: "id do livro e obrigatorio"}
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

func (s *Service) RegisterStockMovement(ctx context.Context, id string, request entity.StockMovementRequest) error {
	if strings.TrimSpace(id) == "" {
		return ValidationError{Message: "id do livro e obrigatorio"}
	}
	request.Tipo = strings.ToUpper(strings.TrimSpace(request.Tipo))
	if request.Tipo != "ENTRADA" && request.Tipo != "SAIDA" && request.Tipo != "AJUSTE" {
		return ValidationError{Message: "tipo de movimento de estoque invalido"}
	}
	if request.Quantidade <= 0 {
		return ValidationError{Message: "quantidade do movimento deve ser maior que zero"}
	}
	if strings.TrimSpace(request.Motivo) == "" {
		return ValidationError{Message: "motivo do movimento de estoque e obrigatorio"}
	}
	return s.repo.RegisterStockMovement(ctx, strings.TrimSpace(id), request)
}

func (s *Service) Validate(ctx context.Context, input entity.ValidationRequest) (entity.ValidationResponse, error) {
	normalizedISBN := onlyDigits(input.ISBN)
	codigoBarra := onlyDigits(input.CodigoBarra)
	status := normalizeStatus(input.Status)
	if status == "" {
		status = "EM_PRODUCAO"
	}

	response := entity.ValidationResponse{
		Status:      status,
		ISBN:        normalizedISBN,
		CodigoBarra: codigoBarra,
		Disponivel:  false,
	}

	var validationMessages []string
	isPublishing := status == "PUBLICADO"

	if isPublishing && normalizedISBN == "" {
		validationMessages = append(validationMessages, "isbn e obrigatorio para publicar o livro")
	}
	if isPublishing && codigoBarra == "" {
		validationMessages = append(validationMessages, "codigo_barra e obrigatorio para publicar o livro")
	}
	if normalizedISBN == "" && codigoBarra != "" {
		validationMessages = append(validationMessages, "isbn deve ser informado junto com codigo_barra")
	}
	if codigoBarra == "" && normalizedISBN != "" {
		validationMessages = append(validationMessages, "codigo_barra deve ser informado junto com isbn")
	}
	if len(normalizedISBN) > 0 && !isISBN13(normalizedISBN) {
		validationMessages = append(validationMessages, "isbn deve ser um ISBN-13 valido")
	}
	if len(codigoBarra) > 0 && !isEAN13(codigoBarra) {
		validationMessages = append(validationMessages, "codigo_barra deve ser um EAN-13 valido")
	}
	if normalizedISBN != "" && codigoBarra != "" && normalizedISBN != codigoBarra {
		validationMessages = append(validationMessages, "codigo_barra deve corresponder ao isbn numerico")
	}

	if len(validationMessages) > 0 {
		response.Erros = validationMessages
		return response, nil
	}
	if normalizedISBN == "" && codigoBarra == "" {
		response.Valid = true
		response.Disponivel = true
		return response, nil
	}

	available, err := s.repo.IsCatalogIdentifierAvailable(ctx, normalizedISBN, codigoBarra, strings.TrimSpace(input.ID))
	if err != nil {
		return entity.ValidationResponse{}, fmt.Errorf("erro ao validar disponibilidade do livro: %w", err)
	}

	response.Disponivel = available
	response.Valid = available
	if !available {
		response.Erros = []string{"isbn ou codigo_barra ja cadastrado"}
	}
	return response, nil
}

func (s *Service) normalizePersistInput(
	ctx context.Context,
	id string,
	request entity.CreateRequest,
) (entity.PersistInput, error) {
	titulo := strings.TrimSpace(request.Titulo)
	if titulo == "" {
		return entity.PersistInput{}, ValidationError{Message: "titulo do livro e obrigatorio"}
	}
	autorID := strings.TrimSpace(request.AutorID)
	if autorID == "" {
		return entity.PersistInput{}, ValidationError{Message: "autor do livro e obrigatorio"}
	}

	status := normalizeStatus(request.Status)
	if status == "" {
		status = "RASCUNHO"
	}
	if status != "RASCUNHO" && status != "EM_PRODUCAO" && status != "PRONTO_PARA_VENDA" && status != "PUBLICADO" && status != "ESGOTADO" && status != "INATIVO" {
		return entity.PersistInput{}, ValidationError{Message: "status do livro invalido"}
	}
	if err := validateCover(request.Capa); err != nil {
		return entity.PersistInput{}, err
	}

	possuiFormatoFisico := request.PossuiFormatoFisico
	possuiFormatoDigital := request.PossuiFormatoDigital
	if !possuiFormatoFisico && !possuiFormatoDigital {
		return entity.PersistInput{}, ValidationError{Message: "ao menos um formato de venda do livro deve ser selecionado"}
	}

	formato := buildFormatoResumo(possuiFormatoFisico, possuiFormatoDigital)
	precoVendaFisico := readFloat(request.PrecoVendaFisico)
	precoVendaDigital := readFloat(request.PrecoVendaDigital)
	precoVenda := resolvePrecoResumo(possuiFormatoFisico, precoVendaFisico, possuiFormatoDigital, precoVendaDigital, request.PrecoVenda)
	custoImpressao := readFloat(request.CustoImpressao)
	if precoVenda < 0 || precoVendaFisico < 0 || precoVendaDigital < 0 || custoImpressao < 0 {
		return entity.PersistInput{}, ValidationError{Message: "valores do livro nao podem ser negativos"}
	}

	estoqueDisponivel := readInt(request.EstoqueDisponivel)
	estoqueMinimo := readInt(request.EstoqueMinimo)
	controlarEstoque := request.ControlarEstoque && possuiFormatoFisico
	if !controlarEstoque {
		controlarEstoque = false
		estoqueDisponivel = 0
		estoqueMinimo = 0
	}
	if estoqueDisponivel < 0 || estoqueMinimo < 0 {
		return entity.PersistInput{}, ValidationError{Message: "estoque do livro nao pode ser negativo"}
	}

	tipoCapa, err := normalizeTipoCapa(request.TipoCapa, possuiFormatoFisico)
	if err != nil {
		return entity.PersistInput{}, err
	}
	possuiBox := request.PossuiBox && possuiFormatoFisico

	if !possuiFormatoFisico {
		custoImpressao = 0
	}

	canalVendaDigital, urlCompraDigital, err := normalizeDigitalSale(
		request.CanalVendaDigital,
		request.URLCompraDigital,
		possuiFormatoDigital,
	)
	if err != nil {
		return entity.PersistInput{}, err
	}

	validation, err := s.Validate(ctx, entity.ValidationRequest{
		ID:          strings.TrimSpace(id),
		Status:      status,
		ISBN:        request.ISBN,
		CodigoBarra: request.CodigoBarra,
	})
	if err != nil {
		return entity.PersistInput{}, err
	}
	if !validation.Valid {
		return entity.PersistInput{}, ValidationError{Message: strings.Join(validation.Erros, "; ")}
	}

	return entity.PersistInput{
		ID:                     strings.TrimSpace(id),
		AutorID:                autorID,
		Titulo:                 titulo,
		Subtitulo:              normalizeOptionalString(request.Subtitulo),
		Sinopse:                normalizeOptionalString(request.Sinopse),
		ISBN:                   normalizeDigitsOptional(validation.ISBN),
		CodigoBarra:            normalizeDigitsOptional(validation.CodigoBarra),
		Status:                 status,
		Formato:                formato,
		PossuiFormatoFisico:    possuiFormatoFisico,
		PossuiFormatoDigital:   possuiFormatoDigital,
		Edicao:                 normalizeOptionalString(request.Edicao),
		Idioma:                 normalizeOptionalString(request.Idioma),
		NumeroPaginas:          normalizeOptionalPositiveInt(request.NumeroPaginas),
		Genero:                 normalizeOptionalString(request.Genero),
		PrecoVenda:             precoVenda,
		PrecoVendaFisico:       priceOrZero(possuiFormatoFisico, precoVendaFisico),
		PrecoVendaDigital:      priceOrZero(possuiFormatoDigital, precoVendaDigital),
		CanalVendaDigital:      canalVendaDigital,
		URLCompraDigital:       urlCompraDigital,
		CustoImpressao:         custoImpressao,
		VendaInfinita:          possuiFormatoFisico && !controlarEstoque,
		ControlarEstoque:       controlarEstoque,
		EstoqueDisponivel:      estoqueDisponivel,
		EstoqueMinimo:          estoqueMinimo,
		PesoGramas:             normalizeOptionalPositiveInt(request.PesoGramas),
		LarguraCm:              normalizeOptionalPositiveFloat(request.LarguraCm),
		AlturaCm:               normalizeOptionalPositiveFloat(request.AlturaCm),
		ProfundidadeCm:         normalizeOptionalPositiveFloat(request.ProfundidadeCm),
		TipoCapa:               tipoCapa,
		PossuiBox:              possuiBox,
		DetalhesEdicao:         normalizeOptionalString(request.DetalhesEdicao),
		DataPublicacaoPrevista: normalizeOptionalDate(request.DataPublicacaoPrevista),
		DataPublicacao:         normalizeOptionalDate(request.DataPublicacao),
		Capa:                   request.Capa,
		Ativo:                  request.Ativo,
	}, nil
}

func onlyDigits(value string) string {
	var builder strings.Builder

	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}

	return builder.String()
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

func normalizeDigitsOptional(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalDate(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalPositiveInt(value *int) *int {
	if value == nil || *value <= 0 {
		return nil
	}
	return value
}

func normalizeOptionalPositiveFloat(value *float64) *float64 {
	if value == nil || *value <= 0 {
		return nil
	}
	return value
}

func readFloat(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func readInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func buildFormatoResumo(possuiFormatoFisico bool, possuiFormatoDigital bool) string {
	if possuiFormatoFisico && possuiFormatoDigital {
		return "HIBRIDO"
	}
	if possuiFormatoDigital {
		return "DIGITAL"
	}
	return "FISICO"
}

func resolvePrecoResumo(
	possuiFormatoFisico bool,
	precoVendaFisico float64,
	possuiFormatoDigital bool,
	precoVendaDigital float64,
	precoVendaLegado *float64,
) float64 {
	if possuiFormatoFisico && precoVendaFisico > 0 {
		return precoVendaFisico
	}
	if possuiFormatoDigital && precoVendaDigital > 0 {
		return precoVendaDigital
	}
	return readFloat(precoVendaLegado)
}

func priceOrZero(enabled bool, value float64) float64 {
	if !enabled {
		return 0
	}
	return value
}

func normalizeTipoCapa(value string, possuiFormatoFisico bool) (*string, error) {
	if !possuiFormatoFisico {
		return nil, nil
	}

	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return nil, nil
	}

	switch trimmed {
	case "BROCHURA", "CAPA_DURA", "ESPIRAL", "GRAMPEADO", "LUXO", "OUTRO":
		return &trimmed, nil
	default:
		return nil, ValidationError{Message: "tipo de capa do livro invalido"}
	}
}

func normalizeDigitalSale(
	canalValue string,
	urlValue string,
	possuiFormatoDigital bool,
) (*string, *string, error) {
	if !possuiFormatoDigital {
		return nil, nil, nil
	}

	canal := strings.ToUpper(strings.TrimSpace(canalValue))
	link := strings.TrimSpace(urlValue)

	if canal == "" && link == "" {
		return nil, nil, nil
	}

	if canal == "" {
		canal = "AMAZON"
	}

	switch canal {
	case "AMAZON", "LINK_EXTERNO":
	default:
		return nil, nil, ValidationError{Message: "canal de venda digital invalido"}
	}

	if link == "" {
		return stringPointer(canal), nil, nil
	}

	parsed, err := url.ParseRequestURI(link)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return nil, nil, ValidationError{Message: "url de compra digital invalida"}
	}

	return stringPointer(canal), &link, nil
}

func stringPointer(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validateCover(capa *entity.CoverInput) error {
	if capa == nil {
		return nil
	}
	if strings.TrimSpace(capa.Base64) == "" {
		return ValidationError{Message: "a capa do livro esta invalida"}
	}
	if strings.TrimSpace(capa.Mime) != "image/webp" {
		return ValidationError{Message: "a capa do livro deve estar em image/webp"}
	}
	if capa.Largura != 1024 || capa.Altura != 1024 {
		return ValidationError{Message: "a capa do livro deve ter exatamente 1024x1024"}
	}
	if capa.TamanhoBytes <= 0 || capa.TamanhoBytes > 150*1024 {
		return ValidationError{Message: "a capa do livro deve ter no maximo 150 KB"}
	}
	if strings.TrimSpace(capa.HashSHA256) == "" {
		return ValidationError{Message: "o hash da capa do livro e obrigatorio"}
	}
	return nil
}

func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}

func isISBN13(value string) bool {
	return len(value) == 13 && hasISBNPrefix(value) && hasValidEAN13Checksum(value)
}

func isEAN13(value string) bool {
	return len(value) == 13 && hasISBNPrefix(value) && hasValidEAN13Checksum(value)
}

func hasISBNPrefix(value string) bool {
	return strings.HasPrefix(value, "978") || strings.HasPrefix(value, "979")
}

func hasValidEAN13Checksum(value string) bool {
	if len(value) != 13 {
		return false
	}

	sum := 0

	for index := 0; index < 12; index++ {
		digit := int(value[index] - '0')
		if index%2 == 0 {
			sum += digit
			continue
		}

		sum += digit * 3
	}

	checkDigit := (10 - (sum % 10)) % 10
	return checkDigit == int(value[12]-'0')
}
