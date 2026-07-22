package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"sistema-editorial/editora/backend/src/logistica/entity"
)

type catalogRepository interface {
	LookupCatalogItems(ctx context.Context, ids []string) (map[string]entity.CatalogItem, error)
}

type settingsRepository interface {
	Get(ctx context.Context) (entity.SettingsRecord, error)
	Upsert(ctx context.Context, input entity.SettingsRecord) error
}

type provider interface {
	Code() string
	Label() string
	IsEnabled(cfg entity.LogisticsConfig) bool
	IsConfigured(cfg entity.LogisticsConfig) bool
	IsSandbox(cfg entity.LogisticsConfig) bool
	Quote(ctx context.Context, cfg entity.LogisticsConfig, input ProviderQuoteInput) ([]entity.QuoteOption, error)
}

type ProviderQuoteInput struct {
	Origin         entity.OriginConfig
	DestinationCEP string
	Items          []ProviderQuoteItem
	DeclaredValue  float64
	Services       []int
	OwnHand        bool
	Receipt        bool
}

type ProviderQuoteItem struct {
	ID             string
	Title          string
	WidthCM        float64
	HeightCM       float64
	LengthCM       float64
	WeightKG       float64
	InsuranceValue float64
	Quantity       int
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type Service struct {
	settings  settingsRepository
	catalog   catalogRepository
	providers map[string]provider
}

func NewService(
	settings settingsRepository,
	catalog catalogRepository,
	providers ...provider,
) *Service {
	registry := make(map[string]provider, len(providers))
	for _, current := range providers {
		if current == nil {
			continue
		}
		registry[current.Code()] = current
	}

	return &Service{
		settings:  settings,
		catalog:   catalog,
		providers: registry,
	}
}

func (s *Service) GetConfig(ctx context.Context) (entity.SettingsResponse, error) {
	record, err := s.settings.Get(ctx)
	if err != nil {
		return entity.SettingsResponse{}, fmt.Errorf("erro ao carregar configuracao de logistica: %w", err)
	}

	providerStatus, err := s.GetProviderStatus(ctx)
	if err != nil {
		return entity.SettingsResponse{}, err
	}

	return entity.SettingsResponse{
		DefaultProvider:    record.DefaultProvider,
		TimeoutSeconds:     record.TimeoutSeconds,
		ContactEmail:       record.ContactEmail,
		Origin:             record.Origin,
		MelhorEnvioEnabled: record.MelhorEnvioEnabled,
		MelhorEnvio:        record.MelhorEnvio,
		SuperFreteEnabled:  record.SuperFreteEnabled,
		SuperFrete:         record.SuperFrete,
		LoggiEnabled:       record.LoggiEnabled,
		Loggi:              record.Loggi,
		FrenetEnabled:      record.FrenetEnabled,
		Frenet:             record.Frenet,
		Providers:          providerStatus.Providers,
		CreatedAt:          record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          record.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) GetProviderStatus(ctx context.Context) (entity.ConfigResponse, error) {
	settings, err := s.loadRuntimeConfig(ctx)
	if err != nil {
		return entity.ConfigResponse{}, err
	}

	response := entity.ConfigResponse{
		DefaultProvider: s.resolveProviderCode(settings, ""),
		OriginCEP:       settings.Origin.CEP,
		OriginCity:      settings.Origin.City,
		OriginState:     settings.Origin.State,
		Providers:       make([]entity.ProviderStatus, 0, len(s.providers)),
	}

	providerCodes := make([]string, 0, len(s.providers))
	for code := range s.providers {
		providerCodes = append(providerCodes, code)
	}
	slices.Sort(providerCodes)

	for _, code := range providerCodes {
		current := s.providers[code]
		response.Providers = append(response.Providers, entity.ProviderStatus{
			Code:       current.Code(),
			Label:      current.Label(),
			Enabled:    current.IsEnabled(settings),
			IsDefault:  code == response.DefaultProvider,
			Configured: current.IsConfigured(settings),
			Sandbox:    current.IsSandbox(settings),
		})
	}

	return response, nil
}

func (s *Service) UpdateConfig(ctx context.Context, request entity.UpdateSettingsRequest) error {
	record, err := s.normalizeSettingsRequest(request)
	if err != nil {
		return err
	}

	if err := s.settings.Upsert(ctx, record); err != nil {
		return fmt.Errorf("erro ao salvar configuracao de logistica: %w", err)
	}

	return nil
}

func (s *Service) CalculateQuote(ctx context.Context, request entity.QuoteRequest) (entity.QuoteResponse, error) {
	settings, err := s.loadRuntimeConfig(ctx)
	if err != nil {
		return entity.QuoteResponse{}, err
	}

	if !settings.HasOriginAddress() {
		return entity.QuoteResponse{}, ValidationError{Message: "origem da logistica nao configurada"}
	}

	destinationCEP := digitsOnly(request.CEPDestino)
	if len(destinationCEP) != 8 {
		return entity.QuoteResponse{}, ValidationError{Message: "cep de destino invalido"}
	}
	providerItems, subtotal, err := s.resolveQuoteItems(ctx, request)
	if err != nil {
		return entity.QuoteResponse{}, err
	}

	declaredValue := request.ValorDeclarado
	if declaredValue <= 0 {
		declaredValue = subtotal
	}

	quoteInput := ProviderQuoteInput{
		Origin:         settings.Origin,
		DestinationCEP: destinationCEP,
		Items:          providerItems,
		DeclaredValue:  declaredValue,
		Services:       request.Servicos,
		OwnHand:        request.MaosProprias,
		Receipt:        request.AvisoRecebimento,
	}

	providerCode := strings.ToUpper(strings.TrimSpace(request.Provider))
	responseProvider := providerCode
	options := make([]entity.QuoteOption, 0)

	if providerCode != "" {
		current, ok := s.providers[providerCode]
		if !ok {
			return entity.QuoteResponse{}, ValidationError{Message: "provider logistico nao suportado"}
		}
		if !current.IsEnabled(settings) {
			return entity.QuoteResponse{}, ValidationError{Message: "provider logistico nao esta habilitado"}
		}

		options, err = current.Quote(ctx, settings, quoteInput)
		if err != nil {
			return entity.QuoteResponse{}, err
		}
	} else {
		responseProvider = "MULTI"
		providerCodes := make([]string, 0, len(s.providers))
		for code := range s.providers {
			providerCodes = append(providerCodes, code)
		}
		slices.Sort(providerCodes)

		for _, code := range providerCodes {
			current := s.providers[code]
			if !current.IsEnabled(settings) {
				continue
			}

			providerOptions, providerErr := current.Quote(ctx, settings, quoteInput)
			if providerErr != nil {
				options = append(options, entity.QuoteOption{
					Provider:    current.Code(),
					ServiceCode: current.Code(),
					ServiceName: current.Label(),
					CarrierName: current.Label(),
					Error:       providerErr.Error(),
				})
				continue
			}

			options = append(options, providerOptions...)
		}

		if len(options) == 0 {
			return entity.QuoteResponse{}, ValidationError{Message: "nenhum provider habilitado retornou opcoes de frete"}
		}
	}

	return entity.QuoteResponse{
		Provider:       responseProvider,
		CEPOrigem:      settings.Origin.CEP,
		CEPDestino:     destinationCEP,
		SubtotalFisico: subtotal,
		Opcoes:         options,
	}, nil
}

func (s *Service) resolveQuoteItems(ctx context.Context, request entity.QuoteRequest) ([]ProviderQuoteItem, float64, error) {
	if request.Pacote != nil {
		pacote := request.Pacote
		if pacote.PesoKG <= 0 || pacote.LarguraCM <= 0 || pacote.AlturaCM <= 0 || pacote.ComprimentoCM <= 0 {
			return nil, 0, ValidationError{Message: "informe peso e dimensoes validos para a cotacao"}
		}

		subtotal := request.ValorDeclarado
		return []ProviderQuoteItem{
			{
				ID:             "PACOTE_MANUAL",
				Title:          "Pacote manual",
				WidthCM:        pacote.LarguraCM,
				HeightCM:       pacote.AlturaCM,
				LengthCM:       pacote.ComprimentoCM,
				WeightKG:       pacote.PesoKG,
				InsuranceValue: subtotal,
				Quantity:       1,
			},
		}, subtotal, nil
	}

	if len(request.Itens) == 0 {
		return nil, 0, ValidationError{Message: "informe as dimensoes do pacote ou ao menos um livro para cotacao"}
	}

	ids := make([]string, 0, len(request.Itens))
	for _, item := range request.Itens {
		livroID := strings.TrimSpace(item.LivroID)
		if livroID == "" {
			return nil, 0, ValidationError{Message: "livro da cotacao e obrigatorio"}
		}
		if item.Quantidade <= 0 {
			return nil, 0, ValidationError{Message: "quantidade do livro deve ser maior que zero"}
		}
		ids = append(ids, livroID)
	}

	catalogItems, err := s.catalog.LookupCatalogItems(ctx, ids)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao carregar livros da cotacao: %w", err)
	}

	providerItems := make([]ProviderQuoteItem, 0, len(request.Itens))
	subtotal := 0.0
	for _, requestedItem := range request.Itens {
		catalogItem, ok := catalogItems[strings.TrimSpace(requestedItem.LivroID)]
		if !ok {
			return nil, 0, ValidationError{Message: "livro informado nao foi encontrado para cotacao"}
		}
		if !catalogItem.Ativo {
			return nil, 0, ValidationError{Message: "livro inativo nao pode ser cotado"}
		}
		if !catalogItem.PossuiFormatoFisico {
			return nil, 0, ValidationError{Message: "livro sem formato fisico nao pode ser cotado na logistica"}
		}
		if catalogItem.PesoGramas <= 0 || catalogItem.LarguraCM <= 0 || catalogItem.AlturaCM <= 0 || catalogItem.ProfundidadeCM <= 0 {
			return nil, 0, ValidationError{Message: "livro sem peso ou dimensoes completos nao pode ser cotado"}
		}

		unitPrice := resolveDeclaredUnitPrice(catalogItem)
		subtotal += unitPrice * float64(requestedItem.Quantidade)
		providerItems = append(providerItems, ProviderQuoteItem{
			ID:             catalogItem.LivroID,
			Title:          catalogItem.Titulo,
			WidthCM:        catalogItem.LarguraCM,
			HeightCM:       catalogItem.AlturaCM,
			LengthCM:       catalogItem.ProfundidadeCM,
			WeightKG:       float64(catalogItem.PesoGramas) / 1000,
			InsuranceValue: unitPrice,
			Quantity:       requestedItem.Quantidade,
		})
	}

	return providerItems, subtotal, nil
}

func resolveDeclaredUnitPrice(item entity.CatalogItem) float64 {
	if item.PrecoVendaFisico > 0 {
		return item.PrecoVendaFisico
	}
	if item.PrecoVenda > 0 {
		return item.PrecoVenda
	}
	return 0
}

func (s *Service) loadRuntimeConfig(ctx context.Context) (entity.LogisticsConfig, error) {
	record, err := s.settings.Get(ctx)
	if err != nil {
		return entity.LogisticsConfig{}, fmt.Errorf("erro ao carregar configuracao de logistica: %w", err)
	}

	timeoutSeconds := record.TimeoutSeconds
	if timeoutSeconds < 5 {
		timeoutSeconds = 5
	}

	return entity.LogisticsConfig{
		DefaultProvider:    strings.ToUpper(strings.TrimSpace(record.DefaultProvider)),
		Timeout:            time.Duration(timeoutSeconds) * time.Second,
		ContactEmail:       strings.TrimSpace(record.ContactEmail),
		Origin:             normalizeOrigin(record.Origin),
		MelhorEnvioEnabled: record.MelhorEnvioEnabled,
		MelhorEnvio:        normalizeMelhorEnvioConfig(record),
		SuperFreteEnabled:  record.SuperFreteEnabled,
		SuperFrete:         normalizeSuperFreteConfig(record),
		LoggiEnabled:       record.LoggiEnabled,
		Loggi:              normalizeLoggiConfig(record),
		FrenetEnabled:      record.FrenetEnabled,
		Frenet:             normalizeFrenetConfig(record),
	}, nil
}

func (s *Service) normalizeSettingsRequest(request entity.UpdateSettingsRequest) (entity.SettingsRecord, error) {
	defaultProvider := strings.ToUpper(strings.TrimSpace(request.DefaultProvider))
	if defaultProvider == "" {
		defaultProvider = entity.ProviderSuperFrete
	}
	if defaultProvider != entity.ProviderSuperFrete &&
		defaultProvider != entity.ProviderMelhorEnvio &&
		defaultProvider != entity.ProviderLoggi &&
		defaultProvider != entity.ProviderFrenet {
		return entity.SettingsRecord{}, ValidationError{Message: "provider padrao invalido"}
	}

	timeoutSeconds := request.TimeoutSeconds
	if timeoutSeconds < 5 {
		timeoutSeconds = 15
	}

	record := entity.SettingsRecord{
		DefaultProvider:    defaultProvider,
		TimeoutSeconds:     timeoutSeconds,
		ContactEmail:       strings.TrimSpace(request.ContactEmail),
		Origin:             normalizeOrigin(request.Origin),
		MelhorEnvioEnabled: request.MelhorEnvioEnabled,
		MelhorEnvio: entity.MelhorEnvioConfig{
			Sandbox:      request.MelhorEnvio.Sandbox,
			BaseURL:      strings.TrimSpace(request.MelhorEnvio.BaseURL),
			AccessToken:  strings.TrimSpace(request.MelhorEnvio.AccessToken),
			RefreshToken: strings.TrimSpace(request.MelhorEnvio.RefreshToken),
			ClientID:     strings.TrimSpace(request.MelhorEnvio.ClientID),
			ClientSecret: strings.TrimSpace(request.MelhorEnvio.ClientSecret),
			RedirectURL:  strings.TrimSpace(request.MelhorEnvio.RedirectURL),
			UserAgent:    strings.TrimSpace(request.MelhorEnvio.UserAgent),
		},
		SuperFreteEnabled: request.SuperFreteEnabled,
		SuperFrete: entity.SuperFreteConfig{
			Sandbox:   request.SuperFrete.Sandbox,
			BaseURL:   strings.TrimSpace(request.SuperFrete.BaseURL),
			Token:     strings.TrimSpace(request.SuperFrete.Token),
			UserAgent: strings.TrimSpace(request.SuperFrete.UserAgent),
			Services:  strings.TrimSpace(request.SuperFrete.Services),
		},
		LoggiEnabled: request.LoggiEnabled,
		Loggi: entity.LoggiConfig{
			Sandbox:           request.Loggi.Sandbox,
			BaseURL:           strings.TrimSpace(request.Loggi.BaseURL),
			CompanyID:         strings.TrimSpace(request.Loggi.CompanyID),
			ClientID:          strings.TrimSpace(request.Loggi.ClientID),
			ClientSecret:      strings.TrimSpace(request.Loggi.ClientSecret),
			PickupType:        strings.TrimSpace(request.Loggi.PickupType),
			ExternalServiceID: strings.TrimSpace(request.Loggi.ExternalServiceID),
		},
		FrenetEnabled: request.FrenetEnabled,
		Frenet: entity.FrenetConfig{
			Sandbox:     request.Frenet.Sandbox,
			BaseURL:     strings.TrimSpace(request.Frenet.BaseURL),
			Token:       strings.TrimSpace(request.Frenet.Token),
			Platform:    strings.TrimSpace(request.Frenet.Platform),
			PlatformVer: strings.TrimSpace(request.Frenet.PlatformVer),
		},
	}

	if record.Origin.Name == "" {
		record.Origin.Name = "Aura Editora"
	}
	if record.ContactEmail == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "email de contato da logistica e obrigatorio"}
	}
	if record.Origin.CEP == "" || len(record.Origin.CEP) != 8 {
		return entity.SettingsRecord{}, ValidationError{Message: "cep de origem invalido"}
	}
	if record.Origin.Address == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "endereco de origem e obrigatorio"}
	}
	if record.Origin.Number == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "numero do endereco de origem e obrigatorio"}
	}
	if record.Origin.District == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "bairro de origem e obrigatorio"}
	}
	if record.Origin.City == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "cidade de origem e obrigatoria"}
	}
	if len(record.Origin.State) != 2 {
		return entity.SettingsRecord{}, ValidationError{Message: "uf de origem invalida"}
	}
	if record.MelhorEnvioEnabled {
		melhorEnvioConfig := normalizeMelhorEnvioConfig(record)
		if melhorEnvioConfig.BaseURL == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "base url do Melhor Envio e obrigatoria"}
		}
		if record.MelhorEnvio.AccessToken == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "access token do Melhor Envio e obrigatorio"}
		}
		if record.MelhorEnvio.UserAgent == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "user agent do Melhor Envio e obrigatorio"}
		}
	}
	if record.SuperFreteEnabled {
		superFreteConfig := normalizeSuperFreteConfig(record)
		if superFreteConfig.BaseURL == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "base url do SuperFrete e obrigatoria"}
		}
		if superFreteConfig.Token == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "token do SuperFrete e obrigatorio"}
		}
		if superFreteConfig.UserAgent == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "user agent do SuperFrete e obrigatorio"}
		}
	}
	if record.LoggiEnabled {
		loggiConfig := normalizeLoggiConfig(record)
		if loggiConfig.BaseURL == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "base url da Loggi e obrigatoria"}
		}
		if loggiConfig.CompanyID == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "company id da Loggi e obrigatorio"}
		}
		if loggiConfig.ClientID == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "client id da Loggi e obrigatorio"}
		}
		if loggiConfig.ClientSecret == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "client secret da Loggi e obrigatorio"}
		}
		if loggiConfig.PickupType == "" && loggiConfig.ExternalServiceID == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "pickup type ou external service id da Loggi e obrigatorio"}
		}
	}
	if record.FrenetEnabled {
		frenetConfig := normalizeFrenetConfig(record)
		if frenetConfig.BaseURL == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "base url da Frenet e obrigatoria"}
		}
		if frenetConfig.Token == "" {
			return entity.SettingsRecord{}, ValidationError{Message: "token da Frenet e obrigatorio"}
		}
	}

	return record, nil
}

func normalizeOrigin(origin entity.OriginConfig) entity.OriginConfig {
	return entity.OriginConfig{
		Name:     strings.TrimSpace(origin.Name),
		CEP:      digitsOnly(origin.CEP),
		Address:  strings.TrimSpace(origin.Address),
		Number:   strings.TrimSpace(origin.Number),
		District: strings.TrimSpace(origin.District),
		City:     strings.TrimSpace(origin.City),
		State:    strings.ToUpper(strings.TrimSpace(origin.State)),
	}
}

func normalizeMelhorEnvioConfig(record entity.SettingsRecord) entity.MelhorEnvioConfig {
	baseURL := strings.TrimSpace(record.MelhorEnvio.BaseURL)
	if baseURL == "" {
		if record.MelhorEnvio.Sandbox {
			baseURL = "https://sandbox.melhorenvio.com.br/api/v2"
		} else {
			baseURL = "https://www.melhorenvio.com.br/api/v2"
		}
	}

	userAgent := strings.TrimSpace(record.MelhorEnvio.UserAgent)
	if userAgent == "" && strings.TrimSpace(record.ContactEmail) != "" {
		userAgent = "Aura Editora (" + strings.TrimSpace(record.ContactEmail) + ")"
	}

	return entity.MelhorEnvioConfig{
		Sandbox:      record.MelhorEnvio.Sandbox,
		BaseURL:      strings.TrimRight(baseURL, "/"),
		AccessToken:  strings.TrimSpace(record.MelhorEnvio.AccessToken),
		RefreshToken: strings.TrimSpace(record.MelhorEnvio.RefreshToken),
		ClientID:     strings.TrimSpace(record.MelhorEnvio.ClientID),
		ClientSecret: strings.TrimSpace(record.MelhorEnvio.ClientSecret),
		RedirectURL:  strings.TrimSpace(record.MelhorEnvio.RedirectURL),
		UserAgent:    userAgent,
	}
}

func normalizeSuperFreteConfig(record entity.SettingsRecord) entity.SuperFreteConfig {
	baseURL := strings.TrimSpace(record.SuperFrete.BaseURL)
	if baseURL == "" {
		if record.SuperFrete.Sandbox {
			baseURL = "https://sandbox.superfrete.com"
		} else {
			baseURL = "https://api.superfrete.com"
		}
	}

	userAgent := strings.TrimSpace(record.SuperFrete.UserAgent)
	if userAgent == "" && strings.TrimSpace(record.ContactEmail) != "" {
		userAgent = "Aura Editora (" + strings.TrimSpace(record.ContactEmail) + ")"
	}

	services := strings.TrimSpace(record.SuperFrete.Services)
	if services == "" {
		services = "1,2,17"
	}

	return entity.SuperFreteConfig{
		Sandbox:   record.SuperFrete.Sandbox,
		BaseURL:   strings.TrimRight(baseURL, "/"),
		Token:     strings.TrimSpace(record.SuperFrete.Token),
		UserAgent: userAgent,
		Services:  services,
	}
}

func normalizeLoggiConfig(record entity.SettingsRecord) entity.LoggiConfig {
	baseURL := strings.TrimSpace(record.Loggi.BaseURL)
	if baseURL == "" {
		if record.Loggi.Sandbox {
			baseURL = "https://stg.api.loggi.com"
		} else {
			baseURL = "https://api.loggi.com"
		}
	}

	return entity.LoggiConfig{
		Sandbox:           record.Loggi.Sandbox,
		BaseURL:           strings.TrimRight(baseURL, "/"),
		CompanyID:         strings.TrimSpace(record.Loggi.CompanyID),
		ClientID:          strings.TrimSpace(record.Loggi.ClientID),
		ClientSecret:      strings.TrimSpace(record.Loggi.ClientSecret),
		PickupType:        strings.TrimSpace(record.Loggi.PickupType),
		ExternalServiceID: strings.TrimSpace(record.Loggi.ExternalServiceID),
	}
}

func normalizeFrenetConfig(record entity.SettingsRecord) entity.FrenetConfig {
	baseURL := strings.TrimSpace(record.Frenet.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.frenet.com.br"
	}

	platform := strings.TrimSpace(record.Frenet.Platform)
	if platform == "" {
		platform = "AURA"
	}

	platformVer := strings.TrimSpace(record.Frenet.PlatformVer)
	if platformVer == "" {
		platformVer = "1.0"
	}

	return entity.FrenetConfig{
		Sandbox:     record.Frenet.Sandbox,
		BaseURL:     strings.TrimRight(baseURL, "/"),
		Token:       strings.TrimSpace(record.Frenet.Token),
		Platform:    platform,
		PlatformVer: platformVer,
	}
}

func (s *Service) resolveProviderCode(settings entity.LogisticsConfig, requested string) string {
	code := strings.ToUpper(strings.TrimSpace(requested))
	if code != "" {
		return code
	}
	if strings.TrimSpace(settings.DefaultProvider) != "" {
		return strings.ToUpper(strings.TrimSpace(settings.DefaultProvider))
	}
	if _, ok := s.providers[entity.ProviderSuperFrete]; ok {
		return entity.ProviderSuperFrete
	}
	if _, ok := s.providers[entity.ProviderMelhorEnvio]; ok {
		return entity.ProviderMelhorEnvio
	}
	if _, ok := s.providers[entity.ProviderLoggi]; ok {
		return entity.ProviderLoggi
	}
	if _, ok := s.providers[entity.ProviderFrenet]; ok {
		return entity.ProviderFrenet
	}
	return entity.ProviderNone
}

func digitsOnly(value string) string {
	builder := strings.Builder{}
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
