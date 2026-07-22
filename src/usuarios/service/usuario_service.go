package service

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sistema-editorial/editora/backend/src/usuarios/entity"
)

var allowedRoles = []string{"ADMIN", "EDITOR", "FUNC", "ESCRITOR", "CLIENTE"}

type repository interface {
	ExistsByCPFOrEmail(ctx context.Context, cpf string, email string) (bool, error)
	Create(ctx context.Context, input entity.PersistInput) (entity.CreateResponse, error)
	List(ctx context.Context, query entity.ListQuery) (entity.ListResponse, error)
	FindByID(ctx context.Context, id string) (entity.ListItem, error)
	ExistsEmailByDifferentID(ctx context.Context, id string, email string) (bool, error)
	Update(ctx context.Context, input entity.PersistInput) (entity.CreateResponse, error)
	UpdateAccessStatus(ctx context.Context, id string, status string, clienteAtivo bool) error
	ResetPassword(ctx context.Context, id string, passwordHash string) error
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func IsValidationError(err error) bool {
	var target ValidationError
	return errors.As(err, &target)
}

type Service struct {
	repo  repository
	admin InitialAdminConfig
}

type InitialAdminConfig struct {
	Email    string
	CPF      string
	Password string
}

func NewService(admin InitialAdminConfig, repo repository) *Service {
	return &Service{
		repo:  repo,
		admin: admin,
	}
}

func (s *Service) Create(ctx context.Context, input entity.CreateRequest) (entity.CreateResponse, error) {
	persistInput, err := s.normalizeAndValidateCreate(input)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	exists, err := s.repo.ExistsByCPFOrEmail(ctx, persistInput.CPF, persistInput.Email)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	if exists {
		return entity.CreateResponse{}, ValidationError{Message: "cpf ou email ja cadastrado"}
	}

	return s.repo.Create(ctx, persistInput)
}

func (s *Service) Update(ctx context.Context, id string, input entity.UpdateRequest) (entity.CreateResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return entity.CreateResponse{}, ValidationError{Message: "id do usuario e obrigatorio"}
	}

	currentUser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	persistInput, err := s.normalizeAndValidateUpdate(id, currentUser, input)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	emailInUse, err := s.repo.ExistsEmailByDifferentID(ctx, id, persistInput.Email)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	if emailInUse {
		return entity.CreateResponse{}, ValidationError{Message: "email ja cadastrado"}
	}

	return s.repo.Update(ctx, persistInput)
}

func (s *Service) List(ctx context.Context, query entity.ListQuery) (entity.ListResponse, error) {
	query.Search = strings.TrimSpace(query.Search)
	query.Role = strings.ToUpper(strings.TrimSpace(query.Role))
	if query.Role == "TODOS" {
		query.Role = ""
	}

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	if query.PageSize > 100 {
		query.PageSize = 100
	}

	if query.Role != "" && !slices.Contains(allowedRoles, query.Role) {
		return entity.ListResponse{}, ValidationError{Message: fmt.Sprintf("perfil invalido: %s", query.Role)}
	}

	return s.repo.List(ctx, query)
}

func (s *Service) FindByID(ctx context.Context, id string) (entity.ListItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return entity.ListItem{}, ValidationError{Message: "id do usuario e obrigatorio"}
	}

	return s.repo.FindByID(ctx, id)
}

func (s *Service) Block(ctx context.Context, id string) (entity.StatusActionResponse, error) {
	user, err := s.FindByID(ctx, id)
	if err != nil {
		return entity.StatusActionResponse{}, err
	}

	if err := s.repo.UpdateAccessStatus(ctx, user.ID, entity.StatusAcessoBloqueado, false); err != nil {
		return entity.StatusActionResponse{}, err
	}

	return entity.StatusActionResponse{
		ID:           user.ID,
		Status:       "Bloqueado",
		StatusCodigo: entity.StatusAcessoBloqueado,
		ClienteAtivo: false,
	}, nil
}

func (s *Service) Activate(ctx context.Context, id string) (entity.StatusActionResponse, error) {
	user, err := s.FindByID(ctx, id)
	if err != nil {
		return entity.StatusActionResponse{}, err
	}

	if err := s.repo.UpdateAccessStatus(ctx, user.ID, entity.StatusAcessoAtivo, true); err != nil {
		return entity.StatusActionResponse{}, err
	}

	return entity.StatusActionResponse{
		ID:           user.ID,
		Status:       "Ativo",
		StatusCodigo: entity.StatusAcessoAtivo,
		ClienteAtivo: true,
	}, nil
}

func (s *Service) ResetPassword(ctx context.Context, id string) (entity.ResetPasswordResponse, error) {
	user, err := s.FindByID(ctx, id)
	if err != nil {
		return entity.ResetPasswordResponse{}, err
	}

	temporaryPassword, err := generateTemporaryPassword(10)
	if err != nil {
		return entity.ResetPasswordResponse{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		return entity.ResetPasswordResponse{}, err
	}

	if err := s.repo.ResetPassword(ctx, user.ID, string(passwordHash)); err != nil {
		return entity.ResetPasswordResponse{}, err
	}

	return entity.ResetPasswordResponse{
		ID:                 user.ID,
		SenhaTemporaria:    temporaryPassword,
		PrecisaTrocarSenha: true,
	}, nil
}

func (s *Service) EnsureInitialAdmin(ctx context.Context) error {
	exists, err := s.repo.ExistsByCPFOrEmail(ctx, s.admin.CPF, s.admin.Email)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(s.admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	placeholderPhoto := base64.StdEncoding.EncodeToString([]byte("aura-admin-placeholder"))
	hash := sha256.Sum256([]byte(placeholderPhoto))

	_, err = s.repo.Create(ctx, entity.PersistInput{
		CPF:                s.admin.CPF,
		Email:              s.admin.Email,
		NomeCompleto:       "Administrador Aura",
		Foto:               entity.FotoInput{Base64: placeholderPhoto, Mime: "image/webp", Largura: 1024, Altura: 1024, TamanhoBytes: len(placeholderPhoto), HashSHA256: hex.EncodeToString(hash[:])},
		WhatsApp:           "11999999999",
		DataNascimento:     "1990-01-01",
		SenhaHash:          string(passwordHash),
		Papeis:             []string{"ADMIN", "CLIENTE"},
		OrigemCadastro:     "EDITORA",
		PrecisaTrocarSenha: false,
		StatusAcesso:       entity.StatusAcessoAtivo,
		ClienteAtivo:       true,
	})

	return err
}

func (s *Service) normalizeAndValidateCreate(input entity.CreateRequest) (entity.PersistInput, error) {
	cpf := digitsOnly(input.CPF)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	nomeCompleto := strings.TrimSpace(input.NomeCompleto)
	descricao := strings.TrimSpace(input.Descricao)
	pseudonimo := strings.TrimSpace(input.Pseudonimo)
	whatsApp := digitsOnly(input.WhatsApp)
	dataNascimento := strings.TrimSpace(input.DataNascimento)
	nacionalidade := strings.TrimSpace(input.Nacionalidade)
	origemCadastro := strings.ToUpper(strings.TrimSpace(input.OrigemCadastro))
	if origemCadastro == "" {
		origemCadastro = "EDITORA"
	}

	if !isValidCPF(cpf) {
		return entity.PersistInput{}, ValidationError{Message: "cpf invalido"}
	}

	if email == "" || !strings.Contains(email, "@") {
		return entity.PersistInput{}, ValidationError{Message: "email invalido"}
	}

	if nomeCompleto == "" {
		return entity.PersistInput{}, ValidationError{Message: "nome_completo e obrigatorio"}
	}

	if len(whatsApp) < 10 || len(whatsApp) > 13 {
		return entity.PersistInput{}, ValidationError{Message: "whatsapp invalido"}
	}

	if _, err := time.Parse("2006-01-02", dataNascimento); err != nil {
		return entity.PersistInput{}, ValidationError{Message: "data_nascimento deve estar no formato yyyy-mm-dd"}
	}

	if wordCount(descricao) > 110 {
		return entity.PersistInput{}, ValidationError{Message: "descricao deve ter no maximo 110 palavras"}
	}

	if len(nacionalidade) > 100 {
		return entity.PersistInput{}, ValidationError{Message: "nacionalidade deve ter no maximo 100 caracteres"}
	}

	roles := normalizeRoles(input.Papeis)
	if len(roles) == 0 {
		return entity.PersistInput{}, ValidationError{Message: "pelo menos um perfil deve ser informado"}
	}

	for _, role := range roles {
		if !slices.Contains(allowedRoles, role) {
			return entity.PersistInput{}, ValidationError{Message: fmt.Sprintf("perfil invalido: %s", role)}
		}
	}

	if err := validatePhoto(input.Foto); err != nil {
		return entity.PersistInput{}, err
	}

	normalizedAddress, err := normalizeAddress(input.EnderecoPrincipal)
	if err != nil {
		return entity.PersistInput{}, err
	}

	password := strings.TrimSpace(input.Senha)
	if password == "" || len(password) < 5 {
		return entity.PersistInput{}, ValidationError{Message: "senha deve ter pelo menos 5 caracteres"}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.PersistInput{}, err
	}

	photoHash := strings.TrimSpace(input.Foto.HashSHA256)
	if photoHash == "" {
		sum := sha256.Sum256([]byte(input.Foto.Base64))
		photoHash = hex.EncodeToString(sum[:])
	}

	return entity.PersistInput{
		CPF:                cpf,
		Email:              email,
		NomeCompleto:       nomeCompleto,
		Foto:               entity.FotoInput{Base64: strings.TrimSpace(input.Foto.Base64), Mime: strings.TrimSpace(input.Foto.Mime), Largura: input.Foto.Largura, Altura: input.Foto.Altura, TamanhoBytes: input.Foto.TamanhoBytes, HashSHA256: photoHash},
		Descricao:          descricao,
		Pseudonimo:         pseudonimo,
		EnderecoPrincipal:  normalizedAddress,
		WhatsApp:           whatsApp,
		DataNascimento:     dataNascimento,
		Nacionalidade:      nacionalidade,
		SenhaHash:          string(passwordHash),
		Papeis:             roles,
		OrigemCadastro:     origemCadastro,
		PrecisaTrocarSenha: origemCadastro == "EDITORA",
		UpdatePassword:     true,
		StatusAcesso:       determineInitialAccessStatus(origemCadastro, roles),
		ClienteAtivo:       true,
	}, nil
}

func (s *Service) normalizeAndValidateUpdate(
	id string,
	current entity.ListItem,
	input entity.UpdateRequest,
) (entity.PersistInput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	nomeCompleto := strings.TrimSpace(input.NomeCompleto)
	descricao := strings.TrimSpace(input.Descricao)
	pseudonimo := strings.TrimSpace(input.Pseudonimo)
	whatsApp := digitsOnly(input.WhatsApp)
	dataNascimento := strings.TrimSpace(input.DataNascimento)
	nacionalidade := strings.TrimSpace(input.Nacionalidade)

	if email == "" || !strings.Contains(email, "@") {
		return entity.PersistInput{}, ValidationError{Message: "email invalido"}
	}

	if nomeCompleto == "" {
		return entity.PersistInput{}, ValidationError{Message: "nome_completo e obrigatorio"}
	}

	if len(whatsApp) < 10 || len(whatsApp) > 13 {
		return entity.PersistInput{}, ValidationError{Message: "whatsapp invalido"}
	}

	if _, err := time.Parse("2006-01-02", dataNascimento); err != nil {
		return entity.PersistInput{}, ValidationError{Message: "data_nascimento deve estar no formato yyyy-mm-dd"}
	}

	if wordCount(descricao) > 110 {
		return entity.PersistInput{}, ValidationError{Message: "descricao deve ter no maximo 110 palavras"}
	}

	if len(nacionalidade) > 100 {
		return entity.PersistInput{}, ValidationError{Message: "nacionalidade deve ter no maximo 100 caracteres"}
	}

	roles := normalizeRoles(input.Papeis)
	if len(roles) == 0 {
		return entity.PersistInput{}, ValidationError{Message: "pelo menos um perfil deve ser informado"}
	}

	for _, role := range roles {
		if !slices.Contains(allowedRoles, role) {
			return entity.PersistInput{}, ValidationError{Message: fmt.Sprintf("perfil invalido: %s", role)}
		}
	}

	if err := validatePhoto(input.Foto); err != nil {
		return entity.PersistInput{}, err
	}

	normalizedAddress, err := normalizeAddress(input.EnderecoPrincipal)
	if err != nil {
		return entity.PersistInput{}, err
	}

	password := strings.TrimSpace(input.Senha)
	updatePassword := password != ""
	passwordHash := ""
	if updatePassword {
		if len(password) < 5 {
			return entity.PersistInput{}, ValidationError{Message: "senha deve ter pelo menos 5 caracteres"}
		}

		generatedHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return entity.PersistInput{}, err
		}
		passwordHash = string(generatedHash)
	}

	photoHash := strings.TrimSpace(input.Foto.HashSHA256)
	if photoHash == "" {
		sum := sha256.Sum256([]byte(input.Foto.Base64))
		photoHash = hex.EncodeToString(sum[:])
	}

	return entity.PersistInput{
		ID:                 id,
		CPF:                current.CPF,
		Email:              email,
		NomeCompleto:       nomeCompleto,
		Foto:               entity.FotoInput{Base64: strings.TrimSpace(input.Foto.Base64), Mime: strings.TrimSpace(input.Foto.Mime), Largura: input.Foto.Largura, Altura: input.Foto.Altura, TamanhoBytes: input.Foto.TamanhoBytes, HashSHA256: photoHash},
		Descricao:          descricao,
		Pseudonimo:         pseudonimo,
		EnderecoPrincipal:  normalizedAddress,
		WhatsApp:           whatsApp,
		DataNascimento:     dataNascimento,
		Nacionalidade:      nacionalidade,
		SenhaHash:          passwordHash,
		Papeis:             roles,
		OrigemCadastro:     current.OrigemCadastro,
		PrecisaTrocarSenha: updatePassword && current.OrigemCadastro == "EDITORA",
		UpdatePassword:     updatePassword,
		StatusAcesso:       normalizeAccessStatus(current.StatusCodigo),
		ClienteAtivo:       current.ClienteAtivo,
	}, nil
}

func normalizeRoles(input []string) []string {
	if len(input) == 0 {
		input = []string{"CLIENTE"}
	}

	unique := make(map[string]struct{})
	for _, role := range input {
		normalized := strings.ToUpper(strings.TrimSpace(role))
		if normalized == "" {
			continue
		}

		unique[normalized] = struct{}{}
	}

	unique["CLIENTE"] = struct{}{}

	result := make([]string, 0, len(unique))
	for role := range unique {
		result = append(result, role)
	}

	slices.Sort(result)
	return result
}

func determineInitialAccessStatus(origemCadastro string, roles []string) string {
	if strings.TrimSpace(strings.ToUpper(origemCadastro)) != "EDITORA" && hasElevatedRole(roles) {
		return entity.StatusAcessoPendenteAprovacao
	}

	return entity.StatusAcessoAtivo
}

func hasElevatedRole(roles []string) bool {
	for _, role := range roles {
		if strings.TrimSpace(strings.ToUpper(role)) != "CLIENTE" {
			return true
		}
	}

	return false
}

func normalizeAccessStatus(status string) string {
	switch strings.TrimSpace(strings.ToUpper(status)) {
	case entity.StatusAcessoBloqueado:
		return entity.StatusAcessoBloqueado
	case entity.StatusAcessoPendenteAprovacao:
		return entity.StatusAcessoPendenteAprovacao
	default:
		return entity.StatusAcessoAtivo
	}
}

func normalizeAddress(address *entity.EnderecoInput) (*entity.EnderecoInput, error) {
	if address == nil {
		return nil, nil
	}

	normalized := &entity.EnderecoInput{
		CEP:         digitsOnly(address.CEP),
		Logradouro:  strings.TrimSpace(address.Logradouro),
		Numero:      strings.TrimSpace(address.Numero),
		Complemento: strings.TrimSpace(address.Complemento),
		Bairro:      strings.TrimSpace(address.Bairro),
		Cidade:      strings.TrimSpace(address.Cidade),
		UF:          strings.ToUpper(strings.TrimSpace(address.UF)),
		Pais:        strings.ToUpper(strings.TrimSpace(address.Pais)),
	}

	if normalized.Pais == "" {
		normalized.Pais = "BRASIL"
	}

	if normalized.CEP == "" &&
		normalized.Logradouro == "" &&
		normalized.Numero == "" &&
		normalized.Bairro == "" &&
		normalized.Cidade == "" &&
		normalized.UF == "" {
		return nil, nil
	}

	if len(normalized.CEP) != 8 || normalized.Logradouro == "" || normalized.Numero == "" ||
		normalized.Bairro == "" || normalized.Cidade == "" || len(normalized.UF) != 2 {
		return nil, ValidationError{Message: "endereco_principal incompleto"}
	}

	return normalized, nil
}

func validatePhoto(photo entity.FotoInput) error {
	if strings.TrimSpace(photo.Base64) == "" {
		return ValidationError{Message: "foto e obrigatoria"}
	}

	if strings.TrimSpace(photo.Mime) != "image/webp" {
		return ValidationError{Message: "foto deve estar em image/webp"}
	}

	if photo.Largura != 1024 || photo.Altura != 1024 {
		return ValidationError{Message: "foto deve ter exatamente 1024x1024"}
	}

	if photo.TamanhoBytes <= 0 || photo.TamanhoBytes > 150*1024 {
		return ValidationError{Message: "foto final deve ter no maximo 150 KB"}
	}

	if _, err := base64.StdEncoding.DecodeString(photo.Base64); err != nil {
		return ValidationError{Message: "foto em base64 invalida"}
	}

	return nil
}

func digitsOnly(value string) string {
	regex := regexp.MustCompile(`\D`)
	return regex.ReplaceAllString(value, "")
}

func wordCount(value string) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}

	return len(strings.Fields(trimmed))
}

func isValidCPF(cpf string) bool {
	if len(cpf) != 11 {
		return false
	}

	allEqual := true
	for index := 1; index < len(cpf); index++ {
		if cpf[index] != cpf[0] {
			allEqual = false
			break
		}
	}

	if allEqual {
		return false
	}

	sum := 0
	for index := 0; index < 9; index++ {
		sum += int(cpf[index]-'0') * (10 - index)
	}

	firstDigit := (sum * 10) % 11
	if firstDigit == 10 {
		firstDigit = 0
	}

	if firstDigit != int(cpf[9]-'0') {
		return false
	}

	sum = 0
	for index := 0; index < 10; index++ {
		sum += int(cpf[index]-'0') * (11 - index)
	}

	secondDigit := (sum * 10) % 11
	if secondDigit == 10 {
		secondDigit = 0
	}

	return secondDigit == int(cpf[10]-'0')
}

func generateTemporaryPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789@#"
	buffer := make([]byte, length)
	randomBytes := make([]byte, length)
	if _, err := cryptorand.Read(randomBytes); err != nil {
		return "", err
	}

	for index, value := range randomBytes {
		buffer[index] = charset[int(value)%len(charset)]
	}

	return string(buffer), nil
}
