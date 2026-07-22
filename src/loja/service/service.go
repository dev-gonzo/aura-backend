package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"sistema-editorial/editora/backend/src/loja/entity"
)

type repository interface {
	GetSettings(ctx context.Context) (entity.SettingsRecord, error)
	GetDraftSettings(ctx context.Context) (entity.SettingsRecord, bool, error)
	SaveDraftSettings(ctx context.Context, input entity.SettingsRecord) error
	FinalizePublishedSettings(ctx context.Context) error
	UpsertSettings(ctx context.Context, input entity.SettingsRecord) error
	ListBanners(ctx context.Context) ([]entity.BannerPayload, error)
	ReplaceBanners(ctx context.Context, items []entity.BannerPayload) error
	ListCategories(ctx context.Context) ([]entity.AdminStoreCategoryListItem, error)
	CreateCategory(ctx context.Context, input entity.StoreCategoryRecord) (string, error)
	UpdateCategory(ctx context.Context, input entity.StoreCategoryRecord) error
	DeleteCategory(ctx context.Context, id string) error
	ListAdminProducts(ctx context.Context) ([]entity.AdminProductListItem, error)
	GetAdminProductByID(ctx context.Context, id string) (entity.AdminProductDetail, error)
	CreateProduct(ctx context.Context, input entity.ProductRecord) (string, error)
	UpdateProduct(ctx context.Context, input entity.ProductRecord) error
	ListPublicProducts(ctx context.Context) ([]entity.PublicProductListItem, error)
	GetPublicProductBySlug(ctx context.Context, slug string) (entity.PublicProductDetail, error)
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
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetSettings(ctx context.Context) (entity.SettingsResponse, error) {
	record, hasDraft, err := s.repository.GetDraftSettings(ctx)
	if err != nil {
		return entity.SettingsResponse{}, fmt.Errorf("erro ao carregar rascunho da loja: %w", err)
	}
	record = normalizeLoadedSettingsRecord(record)

	if !hasDraft {
		record, err = s.repository.GetSettings(ctx)
		if err != nil {
			return entity.SettingsResponse{}, fmt.Errorf("erro ao carregar configuracao da loja: %w", err)
		}
		banners, bannerErr := s.repository.ListBanners(ctx)
		if bannerErr != nil {
			return entity.SettingsResponse{}, fmt.Errorf("erro ao carregar banners da loja: %w", bannerErr)
		}
		record.Banners = banners
		record = normalizeLoadedSettingsRecord(record)
	}

	return mapSettingsResponse(record), nil
}

func (s *Service) UpdateSettings(ctx context.Context, request entity.UpdateSettingsRequest) error {
	record, err := normalizeSettingsRequest(request)
	if err != nil {
		return err
	}
	if err := s.repository.SaveDraftSettings(ctx, record); err != nil {
		return fmt.Errorf("erro ao salvar rascunho da loja: %w", err)
	}
	return nil
}

func (s *Service) UpdateIntegrations(ctx context.Context, settings entity.IntegrationSettings) error {
	normalized := normalizeIntegrationSettings(settings)

	record, err := s.repository.GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("erro ao carregar configuracao publicada da loja: %w", err)
	}
	record = normalizeLoadedSettingsRecord(record)
	record.Integrations = normalized
	if err := s.repository.UpsertSettings(ctx, record); err != nil {
		return fmt.Errorf("erro ao aplicar integracoes da loja: %w", err)
	}

	draftRecord, hasDraft, err := s.repository.GetDraftSettings(ctx)
	if err != nil {
		return fmt.Errorf("erro ao carregar rascunho da loja para sincronizar integracoes: %w", err)
	}
	if !hasDraft {
		return nil
	}

	draftRecord = normalizeLoadedSettingsRecord(draftRecord)
	draftRecord.Integrations = normalized
	if err := s.repository.SaveDraftSettings(ctx, draftRecord); err != nil {
		return fmt.Errorf("erro ao sincronizar integracoes com o rascunho da loja: %w", err)
	}
	return nil
}

func (s *Service) PublishSettings(ctx context.Context) error {
	record, hasDraft, err := s.repository.GetDraftSettings(ctx)
	if err != nil {
		return fmt.Errorf("erro ao carregar rascunho para publicacao: %w", err)
	}
	if !hasDraft {
		return ValidationError{Message: "nenhum rascunho pendente para publicar"}
	}
	record = normalizeLoadedSettingsRecord(record)
	if err := s.repository.UpsertSettings(ctx, record); err != nil {
		return fmt.Errorf("erro ao publicar configuracao da loja: %w", err)
	}
	if err := s.repository.ReplaceBanners(ctx, record.Banners); err != nil {
		return fmt.Errorf("erro ao publicar banners da loja: %w", err)
	}
	if err := s.repository.FinalizePublishedSettings(ctx); err != nil {
		return fmt.Errorf("erro ao finalizar publicacao da loja: %w", err)
	}
	return nil
}

func (s *Service) ListCategories(ctx context.Context) ([]entity.AdminStoreCategoryListItem, error) {
	items, err := s.repository.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias da loja: %w", err)
	}
	return items, nil
}

func (s *Service) CreateCategory(ctx context.Context, request entity.StoreCategoryPayload) (string, error) {
	record, err := normalizeCategoryPayload("", request)
	if err != nil {
		return "", err
	}
	if err := s.validateCategoryUniqueness(ctx, record); err != nil {
		return "", err
	}
	id, err := s.repository.CreateCategory(ctx, record)
	if err != nil {
		return "", fmt.Errorf("erro ao criar categoria da loja: %w", err)
	}
	return id, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id string, request entity.StoreCategoryPayload) error {
	record, err := normalizeCategoryPayload(id, request)
	if err != nil {
		return err
	}
	if err := s.validateCategoryUniqueness(ctx, record); err != nil {
		return err
	}
	if err := s.repository.UpdateCategory(ctx, record); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ValidationError{Message: "categoria da loja nao encontrada"}
		}
		return fmt.Errorf("erro ao atualizar categoria da loja: %w", err)
	}
	return nil
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return ValidationError{Message: "categoria da loja nao encontrada"}
	}
	if err := s.repository.DeleteCategory(ctx, trimmedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ValidationError{Message: "categoria da loja nao encontrada"}
		}
		return fmt.Errorf("erro ao excluir categoria da loja: %w", err)
	}
	return nil
}

func (s *Service) ListAdminProducts(ctx context.Context) ([]entity.AdminProductListItem, error) {
	items, err := s.repository.ListAdminProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar produtos da loja: %w", err)
	}
	return items, nil
}

func (s *Service) GetAdminProductByID(ctx context.Context, id string) (entity.AdminProductDetail, error) {
	item, err := s.repository.GetAdminProductByID(ctx, strings.TrimSpace(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.AdminProductDetail{}, ValidationError{Message: "produto da loja nao encontrado"}
		}
		return entity.AdminProductDetail{}, fmt.Errorf("erro ao carregar produto da loja: %w", err)
	}
	return item, nil
}

func (s *Service) CreateProduct(ctx context.Context, request entity.ProductPayload) (string, error) {
	record, err := normalizeProductPayload("", request)
	if err != nil {
		return "", err
	}
	if err := s.validateProductCategories(ctx, record.Categorias); err != nil {
		return "", err
	}
	id, err := s.repository.CreateProduct(ctx, record)
	if err != nil {
		return "", fmt.Errorf("erro ao criar produto da loja: %w", err)
	}
	return id, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id string, request entity.ProductPayload) error {
	record, err := normalizeProductPayload(id, request)
	if err != nil {
		return err
	}
	if err := s.validateProductCategories(ctx, record.Categorias); err != nil {
		return err
	}
	if err := s.repository.UpdateProduct(ctx, record); err != nil {
		return fmt.Errorf("erro ao atualizar produto da loja: %w", err)
	}
	return nil
}

func (s *Service) GetPublicConfig(ctx context.Context, preview bool) (entity.PublicConfigResponse, error) {
	var (
		record entity.SettingsRecord
		err    error
	)

	if preview {
		var hasDraft bool
		record, hasDraft, err = s.repository.GetDraftSettings(ctx)
		if err != nil {
			return entity.PublicConfigResponse{}, fmt.Errorf("erro ao carregar rascunho publico da loja: %w", err)
		}
		record = normalizeLoadedSettingsRecord(record)
		if !hasDraft {
			record, err = s.repository.GetSettings(ctx)
			if err != nil {
				return entity.PublicConfigResponse{}, fmt.Errorf("erro ao carregar configuracao publica da loja: %w", err)
			}
			banners, bannerErr := s.repository.ListBanners(ctx)
			if bannerErr != nil {
				return entity.PublicConfigResponse{}, fmt.Errorf("erro ao carregar banners publicos da loja: %w", bannerErr)
			}
			record.Banners = banners
			record = normalizeLoadedSettingsRecord(record)
		}
	} else {
		record, err = s.repository.GetSettings(ctx)
		if err != nil {
			return entity.PublicConfigResponse{}, fmt.Errorf("erro ao carregar configuracao publica da loja: %w", err)
		}
		banners, bannerErr := s.repository.ListBanners(ctx)
		if bannerErr != nil {
			return entity.PublicConfigResponse{}, fmt.Errorf("erro ao carregar banners publicos da loja: %w", bannerErr)
		}
		record.Banners = banners
		record = normalizeLoadedSettingsRecord(record)
	}

	return entity.PublicConfigResponse{
		StoreName:             record.StoreName,
		StoreSubtitle:         record.StoreSubtitle,
		StoreDescription:      record.StoreDescription,
		StoreTags:             record.StoreTags,
		StoreInMaintenance:    record.StoreInMaintenance,
		MaintenanceTitle:      record.MaintenanceTitle,
		MaintenanceMessage:    record.MaintenanceMessage,
		PrimaryColor:          record.PrimaryColor,
		AccentColor:           record.AccentColor,
		SecondaryColor:        record.SecondaryColor,
		BackgroundColor:       record.BackgroundColor,
		HeaderColor:           record.HeaderColor,
		MenuColor:             record.MenuColor,
		FooterColor:           record.FooterColor,
		HeaderTextColor:       record.HeaderTextColor,
		MenuTextColor:         record.MenuTextColor,
		FooterTextColor:       record.FooterTextColor,
		PrimaryTextColor:      record.PrimaryTextColor,
		HighlightColor:        record.HighlightColor,
		DiscountColor:         record.DiscountColor,
		LaunchBadgeColor:      record.LaunchBadgeColor,
		FeaturedBadgeColor:    record.FeaturedBadgeColor,
		PromotionBadgeColor:   record.PromotionBadgeColor,
		ButtonColor:           record.ButtonColor,
		ButtonTextColor:       record.ButtonTextColor,
		ButtonStyle:           record.ButtonStyle,
		FontFamily:            record.FontFamily,
		FontSecondaryFamily:   record.FontSecondaryFamily,
		FontMenuFamily:        record.FontMenuFamily,
		FontButtonFamily:      record.FontButtonFamily,
		FontHighlightFamily:   record.FontHighlightFamily,
		FontImportURL:         record.FontImportURL,
		FontEmbedCode:         record.FontEmbedCode,
		Integrations:          record.Integrations,
		CornerStyle:           record.CornerStyle,
		ContentWidthMode:      record.ContentWidthMode,
		MenuBackgroundMode:    record.MenuBackgroundMode,
		BannerWidthMode:       record.BannerWidthMode,
		BannerEffectMode:      record.BannerEffectMode,
		BannerRotationSeconds: record.BannerRotationSeconds,
		BrandDisplayMode:      record.BrandDisplayMode,
		BrandLogoSize:         record.BrandLogoSize,
		ProductsPerRowDesktop: record.ProductsPerRowDesktop,
		ProductsPerRowMobile:  record.ProductsPerRowMobile,
		HeroTitle:             record.HeroTitle,
		HeroSubtitle:          record.HeroSubtitle,
		Logo:                  record.Logo,
		Favicon:               record.Favicon,
		BannerDesktop:         record.BannerDesktop,
		BannerMobile:          record.BannerMobile,
		Banners:               record.Banners,
		MenuLinks:             record.MenuLinks,
		SectionOrder:          sanitizeSectionOrder(record.SectionOrder),
		VisibleSections:       sanitizeSectionVisibility(record.VisibleSections),
		CustomBlockEyebrow:    record.CustomBlockEyebrow,
		CustomBlockTitle:      record.CustomBlockTitle,
		CustomBlockDescription: record.CustomBlockDescription,
		CustomBlockHTML:       record.CustomBlockHTML,
		CustomBlockCSS:        record.CustomBlockCSS,
		CustomBlockJS:         record.CustomBlockJS,
		FeatureHighlights:     record.FeatureHighlights,
		InstitutionalSection:  record.InstitutionalSection,
		FooterLinks:           record.FooterLinks,
		FooterContactTitle:    record.FooterContactTitle,
		FooterContactText:     record.FooterContactText,
		LaunchesSection:       record.LaunchesSection,
		FeaturedSection:       record.FeaturedSection,
		PromotionsSection:     record.PromotionsSection,
	}, nil
}

func (s *Service) ListPublicProducts(ctx context.Context) ([]entity.PublicProductListItem, error) {
	items, err := s.repository.ListPublicProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar catalogo publico da loja: %w", err)
	}
	return items, nil
}

func (s *Service) GetPublicProductBySlug(ctx context.Context, slug string) (entity.PublicProductDetail, error) {
	item, err := s.repository.GetPublicProductBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.PublicProductDetail{}, ValidationError{Message: "produto da loja nao encontrado"}
		}
		return entity.PublicProductDetail{}, fmt.Errorf("erro ao carregar produto publico da loja: %w", err)
	}
	return item, nil
}

func mapSettingsResponse(record entity.SettingsRecord) entity.SettingsResponse {
	return entity.SettingsResponse{
		StoreName:             record.StoreName,
		StoreSubtitle:         record.StoreSubtitle,
		StoreDescription:      record.StoreDescription,
		StoreTags:             record.StoreTags,
		StoreInMaintenance:    record.StoreInMaintenance,
		MaintenanceTitle:      record.MaintenanceTitle,
		MaintenanceMessage:    record.MaintenanceMessage,
		PrimaryColor:          record.PrimaryColor,
		AccentColor:           record.AccentColor,
		SecondaryColor:        record.SecondaryColor,
		BackgroundColor:       record.BackgroundColor,
		HeaderColor:           record.HeaderColor,
		MenuColor:             record.MenuColor,
		FooterColor:           record.FooterColor,
		HeaderTextColor:       record.HeaderTextColor,
		MenuTextColor:         record.MenuTextColor,
		FooterTextColor:       record.FooterTextColor,
		PrimaryTextColor:      record.PrimaryTextColor,
		HighlightColor:        record.HighlightColor,
		DiscountColor:         record.DiscountColor,
		LaunchBadgeColor:      record.LaunchBadgeColor,
		FeaturedBadgeColor:    record.FeaturedBadgeColor,
		PromotionBadgeColor:   record.PromotionBadgeColor,
		ButtonColor:           record.ButtonColor,
		ButtonTextColor:       record.ButtonTextColor,
		ButtonStyle:           record.ButtonStyle,
		FontFamily:            record.FontFamily,
		FontSecondaryFamily:   record.FontSecondaryFamily,
		FontMenuFamily:        record.FontMenuFamily,
		FontButtonFamily:      record.FontButtonFamily,
		FontHighlightFamily:   record.FontHighlightFamily,
		FontImportURL:         record.FontImportURL,
		FontEmbedCode:         record.FontEmbedCode,
		Integrations:          record.Integrations,
		CornerStyle:           record.CornerStyle,
		ContentWidthMode:      record.ContentWidthMode,
		MenuBackgroundMode:    record.MenuBackgroundMode,
		BannerWidthMode:       record.BannerWidthMode,
		BannerEffectMode:      record.BannerEffectMode,
		BannerRotationSeconds: record.BannerRotationSeconds,
		BrandDisplayMode:      record.BrandDisplayMode,
		BrandLogoSize:         record.BrandLogoSize,
		ProductsPerRowDesktop: record.ProductsPerRowDesktop,
		ProductsPerRowMobile:  record.ProductsPerRowMobile,
		HeroTitle:             record.HeroTitle,
		HeroSubtitle:          record.HeroSubtitle,
		Logo:                  record.Logo,
		Favicon:               record.Favicon,
		BannerDesktop:         record.BannerDesktop,
		BannerMobile:          record.BannerMobile,
		Banners:               record.Banners,
		MenuLinks:             record.MenuLinks,
		SectionOrder:          sanitizeSectionOrder(record.SectionOrder),
		VisibleSections:       sanitizeSectionVisibility(record.VisibleSections),
		CustomBlockEyebrow:    record.CustomBlockEyebrow,
		CustomBlockTitle:      record.CustomBlockTitle,
		CustomBlockDescription: record.CustomBlockDescription,
		CustomBlockHTML:       record.CustomBlockHTML,
		CustomBlockCSS:        record.CustomBlockCSS,
		CustomBlockJS:         record.CustomBlockJS,
		FeatureHighlights:     record.FeatureHighlights,
		InstitutionalSection:  record.InstitutionalSection,
		FooterLinks:           record.FooterLinks,
		FooterContactTitle:    record.FooterContactTitle,
		FooterContactText:     record.FooterContactText,
		LaunchesSection:       record.LaunchesSection,
		FeaturedSection:       record.FeaturedSection,
		PromotionsSection:     record.PromotionsSection,
		CreatedAt:             formatOptionalTime(record.CreatedAt),
		UpdatedAt:             formatOptionalTime(record.UpdatedAt),
		DraftUpdatedAt:        formatOptionalTime(record.DraftUpdatedAt),
		PublishedAt:           formatOptionalTime(record.PublishedAt),
		HasUnpublishedChanges: record.HasUnpublishedChanges,
	}
}

func normalizeSettingsRequest(request entity.UpdateSettingsRequest) (entity.SettingsRecord, error) {
	record := entity.SettingsRecord{
		StoreName:             strings.TrimSpace(request.StoreName),
		StoreSubtitle:         strings.TrimSpace(request.StoreSubtitle),
		StoreDescription:      strings.TrimSpace(request.StoreDescription),
		StoreTags:             strings.TrimSpace(request.StoreTags),
		StoreInMaintenance:    request.StoreInMaintenance,
		MaintenanceTitle:      strings.TrimSpace(request.MaintenanceTitle),
		MaintenanceMessage:    strings.TrimSpace(request.MaintenanceMessage),
		PrimaryColor:          normalizeHexColor(request.PrimaryColor, "#7c3aed"),
		AccentColor:           normalizeHexColor(request.AccentColor, "#22c55e"),
		SecondaryColor:        normalizeHexColor(request.SecondaryColor, "#A78BFA"),
		BackgroundColor:       normalizeHexColor(request.BackgroundColor, "#020617"),
		HeaderColor:           normalizeHexColor(request.HeaderColor, "#020617"),
		MenuColor:             normalizeHexColor(request.MenuColor, "#0F172A"),
		FooterColor:           normalizeHexColor(request.FooterColor, "#111827"),
		HeaderTextColor:       normalizeHexColor(request.HeaderTextColor, "#F8FAFC"),
		MenuTextColor:         normalizeHexColor(request.MenuTextColor, "#E2E8F0"),
		FooterTextColor:       normalizeHexColor(request.FooterTextColor, "#CBD5E1"),
		PrimaryTextColor:      normalizeHexColor(request.PrimaryTextColor, "#F8FAFC"),
		HighlightColor:        normalizeHexColor(request.HighlightColor, "#F59E0B"),
		DiscountColor:         normalizeHexColor(request.DiscountColor, "#EF4444"),
		LaunchBadgeColor:      normalizeHexColor(request.LaunchBadgeColor, "#1D4ED8"),
		FeaturedBadgeColor:    normalizeHexColor(request.FeaturedBadgeColor, "#166534"),
		PromotionBadgeColor:   normalizeHexColor(request.PromotionBadgeColor, "#C2410C"),
		ButtonColor:           normalizeHexColor(request.ButtonColor, "#7C3AED"),
		ButtonTextColor:       normalizeHexColor(request.ButtonTextColor, "#F8FAFC"),
		ButtonStyle:           normalizeButtonStyle(request.ButtonStyle),
		FontFamily:            normalizeFontFamily(request.FontFamily),
		FontSecondaryFamily:   normalizeFontFamily(request.FontSecondaryFamily),
		FontMenuFamily:        normalizeFontFamily(request.FontMenuFamily),
		FontButtonFamily:      normalizeFontFamily(request.FontButtonFamily),
		FontHighlightFamily:   normalizeFontFamily(request.FontHighlightFamily),
		FontImportURL:         normalizeGoogleFontImportURL(request.FontImportURL),
		FontEmbedCode:         strings.TrimSpace(request.FontEmbedCode),
		Integrations:          normalizeIntegrationSettings(request.Integrations),
		CornerStyle:           normalizeCornerStyle(request.CornerStyle),
		ContentWidthMode:      normalizeContentWidthMode(request.ContentWidthMode),
		MenuBackgroundMode:    normalizeMenuBackgroundMode(request.MenuBackgroundMode),
		BannerWidthMode:       normalizeBannerWidthMode(request.BannerWidthMode, normalizeContentWidthMode(request.ContentWidthMode)),
		BannerEffectMode:      normalizeBannerEffectMode(request.BannerEffectMode),
		BannerRotationSeconds: normalizeBannerRotationSeconds(request.BannerRotationSeconds),
		BrandDisplayMode:      normalizeBrandDisplayMode(request.BrandDisplayMode),
		BrandLogoSize:         normalizeBrandLogoSize(request.BrandLogoSize),
		HeroTitle:             strings.TrimSpace(request.HeroTitle),
		HeroSubtitle:          strings.TrimSpace(request.HeroSubtitle),
		Logo:                  sanitizeImageAsset(request.Logo),
		Favicon:               sanitizeImageAsset(request.Favicon),
		BannerDesktop:         sanitizeImageAsset(request.BannerDesktop),
		BannerMobile:          sanitizeImageAsset(request.BannerMobile),
		Banners:               sanitizeBanners(request.Banners),
		MenuLinks:             sanitizeNavigationLinks(request.MenuLinks),
		SectionOrder:          sanitizeSectionOrder(request.SectionOrder),
		VisibleSections:       sanitizeSectionVisibility(request.VisibleSections),
		CustomBlockEyebrow:    strings.TrimSpace(request.CustomBlockEyebrow),
		CustomBlockTitle:      strings.TrimSpace(request.CustomBlockTitle),
		CustomBlockDescription: strings.TrimSpace(request.CustomBlockDescription),
		CustomBlockHTML:       strings.TrimSpace(request.CustomBlockHTML),
		CustomBlockCSS:        strings.TrimSpace(request.CustomBlockCSS),
		CustomBlockJS:         strings.TrimSpace(request.CustomBlockJS),
		FeatureHighlights:     sanitizeFeatureHighlights(request.FeatureHighlights),
		InstitutionalSection:  normalizeInstitutionalSectionConfig(request.InstitutionalSection),
		FooterLinks:           sanitizeNavigationLinks(request.FooterLinks),
		FooterContactTitle:    strings.TrimSpace(request.FooterContactTitle),
		FooterContactText:     strings.TrimSpace(request.FooterContactText),
		LaunchesSection:       normalizeProductSectionConfig(request.LaunchesSection, defaultLaunchesSectionConfig()),
		FeaturedSection:       normalizeProductSectionConfig(request.FeaturedSection, defaultFeaturedSectionConfig()),
		PromotionsSection:     normalizeProductSectionConfig(request.PromotionsSection, defaultPromotionsSectionConfig()),
	}

	if record.StoreName == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "nome da loja e obrigatorio"}
	}
	if strings.TrimSpace(request.LaunchesSection.Title) == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "o titulo da vitrine de lancamentos e obrigatorio"}
	}
	if strings.TrimSpace(request.FeaturedSection.Title) == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "o titulo da vitrine de destaque e obrigatorio"}
	}
	if strings.TrimSpace(request.PromotionsSection.Title) == "" {
		return entity.SettingsRecord{}, ValidationError{Message: "o titulo da vitrine de promocoes e obrigatorio"}
	}
	if request.BannerRotationSeconds <= 0 {
		return entity.SettingsRecord{}, ValidationError{Message: "o tempo de exibicao de cada banner e obrigatorio"}
	}
	if request.BannerRotationSeconds < 2 || request.BannerRotationSeconds > 15 {
		return entity.SettingsRecord{}, ValidationError{Message: "o tempo de exibicao de cada banner deve ficar entre 2 e 15 segundos"}
	}
	record.ProductsPerRowDesktop = normalizeProductsPerRowDesktop(request.ProductsPerRowDesktop, record.ContentWidthMode)
	record.ProductsPerRowMobile = normalizeProductsPerRowMobile(request.ProductsPerRowMobile)
	if record.StoreInMaintenance && record.MaintenanceTitle == "" {
		record.MaintenanceTitle = "Loja em manutenção"
	}
	if record.StoreInMaintenance && record.MaintenanceMessage == "" {
		record.MaintenanceMessage = "Estamos em manutenção no momento. Em breve retornaremos."
	}

	if record.FontSecondaryFamily == "" {
		record.FontSecondaryFamily = record.FontFamily
	}
	if record.FontMenuFamily == "" {
		record.FontMenuFamily = firstNonEmpty(record.FontSecondaryFamily, record.FontFamily)
	}
	if record.FontButtonFamily == "" {
		record.FontButtonFamily = record.FontFamily
	}
	if record.FontHighlightFamily == "" {
		record.FontHighlightFamily = firstNonEmpty(record.FontSecondaryFamily, record.FontFamily)
	}
	return record, nil
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() || value.Unix() <= 0 {
		return ""
	}
	return value.Format(time.RFC3339)
}

func (s *Service) validateCategoryUniqueness(ctx context.Context, input entity.StoreCategoryRecord) error {
	items, err := s.repository.ListCategories(ctx)
	if err != nil {
		return fmt.Errorf("erro ao validar categorias da loja: %w", err)
	}

	for _, item := range items {
		if item.ID == input.ID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.Nome), input.Nome) {
			return ValidationError{Message: "ja existe uma categoria com esse nome"}
		}
		if strings.EqualFold(strings.TrimSpace(item.Slug), input.Slug) {
			return ValidationError{Message: "ja existe uma categoria com esse slug"}
		}
	}

	return nil
}

func (s *Service) validateProductCategories(ctx context.Context, categories []string) error {
	if len(categories) == 0 {
		return nil
	}

	items, err := s.repository.ListCategories(ctx)
	if err != nil {
		return fmt.Errorf("erro ao validar categorias do produto da loja: %w", err)
	}

	available := make(map[string]struct{}, len(items))
	for _, item := range items {
		available[strings.ToLower(strings.TrimSpace(item.Nome))] = struct{}{}
	}

	for _, category := range categories {
		if _, exists := available[strings.ToLower(strings.TrimSpace(category))]; !exists {
			return ValidationError{Message: fmt.Sprintf("categoria da loja nao encontrada: %s", category)}
		}
	}

	return nil
}

func normalizeCategoryPayload(id string, request entity.StoreCategoryPayload) (entity.StoreCategoryRecord, error) {
	record := entity.StoreCategoryRecord{
		ID: strings.TrimSpace(id),
		StoreCategoryPayload: entity.StoreCategoryPayload{
			Nome:      strings.TrimSpace(request.Nome),
			Slug:      normalizeSlug(request.Slug),
			Descricao: strings.TrimSpace(request.Descricao),
			Ordem:     request.Ordem,
			Ativa:     request.Ativa,
		},
	}

	if record.Nome == "" {
		return entity.StoreCategoryRecord{}, ValidationError{Message: "nome da categoria e obrigatorio"}
	}
	if record.Slug == "" {
		record.Slug = normalizeSlug(record.Nome)
	}
	if record.Slug == "" {
		return entity.StoreCategoryRecord{}, ValidationError{Message: "slug da categoria e obrigatorio"}
	}
	if record.Ordem < 0 {
		return entity.StoreCategoryRecord{}, ValidationError{Message: "ordem da categoria nao pode ser negativa"}
	}

	return record, nil
}

func normalizeProductPayload(id string, request entity.ProductPayload) (entity.ProductRecord, error) {
	categories := normalizeProductCategories(request.Categorias)
	if trimmedLegacyCategory := strings.TrimSpace(request.Categoria); trimmedLegacyCategory != "" {
		categories = normalizeProductCategories(append(categories, trimmedLegacyCategory))
	}

	record := entity.ProductRecord{
		ID: strings.TrimSpace(id),
		ProductPayload: entity.ProductPayload{
			LivroID:          strings.TrimSpace(request.LivroID),
			IsBook:           shouldTreatProductAsBook(request),
			Authors:          normalizeProductAuthors(request.Authors, request.AutorNome),
			AutorNome:        strings.TrimSpace(request.AutorNome),
			Marca:            strings.TrimSpace(request.Marca),
			Editora:          strings.TrimSpace(request.Editora),
			Subtitulo:        strings.TrimSpace(request.Subtitulo),
			Sinopse:          strings.TrimSpace(request.Sinopse),
			ISBN:             onlyDigits(request.ISBN),
			CodigoBarra:      onlyDigits(request.CodigoBarra),
			Edicao:           strings.TrimSpace(request.Edicao),
			Idioma:           strings.TrimSpace(request.Idioma),
			NumeroPaginas:    normalizeOptionalInt(request.NumeroPaginas),
			Genero:           strings.TrimSpace(request.Genero),
			DataPublicacao:   normalizeOptionalDate(request.DataPublicacao),
			TipoCapa:         strings.TrimSpace(request.TipoCapa),
			PesoGramas:       normalizeOptionalInt(request.PesoGramas),
			LarguraCm:        normalizeOptionalFloat(request.LarguraCm),
			AlturaCm:         normalizeOptionalFloat(request.AlturaCm),
			ProfundidadeCm:   normalizeOptionalFloat(request.ProfundidadeCm),
			Slug:             normalizeSlug(request.Slug),
			NomeExibicao:     strings.TrimSpace(request.NomeExibicao),
			DescricaoCurta:   strings.TrimSpace(request.DescricaoCurta),
			Categoria:        strings.Join(categories, ", "),
			Categorias:       categories,
			PrecoVenda:       request.PrecoVenda,
			EmPromocao:       request.EmPromocao,
			PrecoPromocional: request.PrecoPromocional,
			Destaque:         request.Destaque,
			Lancamento:       request.Lancamento,
			Ativo:            request.Ativo,
			Ordem:            request.Ordem,
			Fotos:            sanitizeProductPhotos(request.Fotos),
		},
	}

	record.AutorNome = joinProductAuthorNames(record.Authors)

	if record.Slug == "" {
		return entity.ProductRecord{}, ValidationError{Message: "slug do produto da loja e obrigatorio"}
	}
	if record.NomeExibicao == "" {
		return entity.ProductRecord{}, ValidationError{Message: "nome de exibicao do produto da loja e obrigatorio"}
	}
	if record.IsBook && len(record.Authors) == 0 {
		return entity.ProductRecord{}, ValidationError{Message: "selecione ao menos um autor para livros"}
	}
	if record.ISBN != "" && len(record.ISBN) != 13 {
		return entity.ProductRecord{}, ValidationError{Message: "isbn deve conter 13 digitos"}
	}
	if record.CodigoBarra != "" && len(record.CodigoBarra) != 13 {
		return entity.ProductRecord{}, ValidationError{Message: "codigo de barras deve conter 13 digitos"}
	}
	if record.PrecoVenda < 0 {
		return entity.ProductRecord{}, ValidationError{Message: "preco de venda do produto da loja nao pode ser negativo"}
	}
	if record.PrecoPromocional < 0 {
		return entity.ProductRecord{}, ValidationError{Message: "preco promocional do produto da loja nao pode ser negativo"}
	}
	if record.EmPromocao && record.PrecoPromocional <= 0 {
		return entity.ProductRecord{}, ValidationError{Message: "informe um preco promocional valido"}
	}
	return record, nil
}

func shouldTreatProductAsBook(request entity.ProductPayload) bool {
	if request.IsBook {
		return true
	}
	if strings.TrimSpace(request.LivroID) != "" {
		return true
	}
	if len(request.Authors) > 0 {
		return true
	}
	return strings.TrimSpace(request.AutorNome) != ""
}

func normalizeProductAuthors(items []entity.ProductAuthor, fallbackAuthor string) []entity.ProductAuthor {
	normalized := make([]entity.ProductAuthor, 0, len(items)+1)
	seen := make(map[string]struct{}, len(items)+1)

	appendAuthor := func(id string, nome string) {
		trimmedID := strings.TrimSpace(id)
		trimmedNome := strings.TrimSpace(nome)
		if trimmedNome == "" {
			return
		}
		key := strings.ToLower(trimmedID + "::" + trimmedNome)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		normalized = append(normalized, entity.ProductAuthor{
			ID:   trimmedID,
			Nome: trimmedNome,
		})
	}

	for _, item := range items {
		appendAuthor(item.ID, item.Nome)
	}

	if len(normalized) == 0 {
		appendAuthor("", fallbackAuthor)
	}

	return normalized
}

func joinProductAuthorNames(items []entity.ProductAuthor) string {
	if len(items) == 0 {
		return ""
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item.Nome)
		if trimmed == "" {
			continue
		}
		names = append(names, trimmed)
	}
	return strings.Join(names, ", ")
}

func onlyDigits(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func normalizeOptionalInt(value *int) *int {
	if value == nil || *value <= 0 {
		return nil
	}
	normalized := *value
	return &normalized
}

func normalizeOptionalFloat(value *float64) *float64 {
	if value == nil || *value <= 0 {
		return nil
	}
	normalized := *value
	return &normalized
}

func normalizeOptionalDate(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return ""
	}
	return trimmed
}

func normalizeProductCategories(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(items))
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func sanitizeImageAsset(asset *entity.ImageAsset) *entity.ImageAsset {
	if asset == nil || strings.TrimSpace(asset.Base64) == "" {
		return nil
	}
	return &entity.ImageAsset{
		Base64:       strings.TrimSpace(asset.Base64),
		Mime:         strings.TrimSpace(asset.Mime),
		Largura:      asset.Largura,
		Altura:       asset.Altura,
		TamanhoBytes: asset.TamanhoBytes,
		HashSHA256:   strings.TrimSpace(asset.HashSHA256),
	}
}

func sanitizeProductPhotos(items []entity.ProductPhoto) []entity.ProductPhoto {
	if len(items) == 0 {
		return []entity.ProductPhoto{}
	}

	photos := make([]entity.ProductPhoto, 0, len(items))
	for index, item := range items {
		if len(photos) >= 5 {
			break
		}

		image := sanitizeImageAsset(item.Image)
		if image == nil {
			continue
		}

		photos = append(photos, entity.ProductPhoto{
			ID:        strings.TrimSpace(item.ID),
			Order:     index,
			IsPrimary: item.IsPrimary,
			Image:     image,
		})
	}

	if len(photos) == 0 {
		return []entity.ProductPhoto{}
	}

	primaryIndex := 0
	for index, item := range photos {
		if item.IsPrimary {
			primaryIndex = index
			break
		}
	}

	for index := range photos {
		photos[index].Order = index
		photos[index].IsPrimary = index == primaryIndex
	}

	return photos
}

func sanitizeBanners(items []entity.BannerPayload) []entity.BannerPayload {
	if len(items) == 0 {
		return []entity.BannerPayload{}
	}

	banners := make([]entity.BannerPayload, 0, len(items))
	for index, item := range items {
		desktop := sanitizeImageAsset(item.Desktop)
		mobile := sanitizeImageAsset(item.Mobile)
		if desktop == nil && mobile == nil {
			continue
		}

		banners = append(banners, entity.BannerPayload{
			ID:               strings.TrimSpace(item.ID),
			Title:            strings.TrimSpace(item.Title),
			Subtitle:         strings.TrimSpace(item.Subtitle),
			Link:             strings.TrimSpace(item.Link),
			Order:            index,
			Active:           item.Active,
			ShowContent:      normalizeBannerShowContent(item.ShowContent),
			LinkMode:         normalizeBannerLinkMode(item.LinkMode, item.UseButton, item.Link),
			ContentPosition:  synthesizeBannerLegacyContentPosition(item.ContentPositionX, item.ContentPositionY, item.ContentPosition),
			ContentPositionX: normalizeBannerContentPositionX(item.ContentPositionX, item.ContentPosition),
			ContentPositionY: normalizeBannerContentPositionY(item.ContentPositionY, item.ContentPosition),
			UseButton:        normalizeBannerUseButton(item.UseButton, item.ButtonLabel, item.Link),
			ButtonLabel:      normalizeBannerButtonLabel(item.ButtonLabel),
			Desktop:          desktop,
			Mobile:           mobile,
		})
	}

	return banners
}

func normalizeBannerContentPosition(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "top", "bottom", "left", "right", "center":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "center"
	}
}

func normalizeBannerContentPositionX(value string, legacy string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "left", "center", "right":
		return strings.TrimSpace(strings.ToLower(value))
	}

	switch normalizeBannerContentPosition(legacy) {
	case "left":
		return "left"
	case "right":
		return "right"
	default:
		return "center"
	}
}

func normalizeBannerContentPositionY(value string, legacy string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "top", "center", "bottom":
		return strings.TrimSpace(strings.ToLower(value))
	}

	switch normalizeBannerContentPosition(legacy) {
	case "top":
		return "top"
	case "bottom":
		return "bottom"
	default:
		return "center"
	}
}

func synthesizeBannerLegacyContentPosition(positionX string, positionY string, legacy string) string {
	normalizedX := normalizeBannerContentPositionX(positionX, legacy)
	normalizedY := normalizeBannerContentPositionY(positionY, legacy)

	switch {
	case normalizedY == "top" && normalizedX == "center":
		return "top"
	case normalizedY == "bottom" && normalizedX == "center":
		return "bottom"
	case normalizedX == "left" && normalizedY == "center":
		return "left"
	case normalizedX == "right" && normalizedY == "center":
		return "right"
	default:
		return "center"
	}
}

func normalizeBannerButtonLabel(value string) string {
	return strings.TrimSpace(value)
}

func normalizeBannerShowContent(value bool) bool {
	return value
}

func normalizeBannerLinkMode(value string, useButton bool, link string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "banner", "button":
		if strings.TrimSpace(link) == "" {
			return "none"
		}
		return strings.TrimSpace(strings.ToLower(value))
	case "none":
		return "none"
	}

	if normalizeBannerUseButton(useButton, "", link) {
		return "button"
	}
	if strings.TrimSpace(link) != "" {
		return "banner"
	}
	return "none"
}

func normalizeBannerUseButton(useButton bool, buttonLabel string, link string) bool {
	return useButton && strings.TrimSpace(buttonLabel) != "" && strings.TrimSpace(link) != ""
}

func sanitizeNavigationLinks(items []entity.NavigationLink) []entity.NavigationLink {
	if len(items) == 0 {
		return []entity.NavigationLink{}
	}

	links := make([]entity.NavigationLink, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		url := strings.TrimSpace(item.URL)
		if label == "" || url == "" {
			continue
		}
		links = append(links, entity.NavigationLink{
			Label:   label,
			URL:     url,
			Visible: item.Visible,
			Kind:    strings.TrimSpace(item.Kind),
		})
	}
	return links
}

func sanitizeFeatureHighlights(items []entity.FeatureHighlight) []entity.FeatureHighlight {
	if len(items) == 0 {
		return []entity.FeatureHighlight{}
	}

	highlights := make([]entity.FeatureHighlight, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		text := strings.TrimSpace(item.Text)
		icon := strings.TrimSpace(item.Icon)
		if title == "" || text == "" {
			continue
		}
		highlights = append(highlights, entity.FeatureHighlight{
			Title:      title,
			Text:       text,
			Icon:       icon,
			TextAlign:  normalizeTextAlign(item.TextAlign),
			IconSize:   normalizeIconSize(item.IconSize),
			FontFamily: normalizeFontFamily(item.FontFamily),
		})
	}
	return highlights
}

func sanitizeSectionOrder(items []entity.StorefrontSection) []entity.StorefrontSection {
	defaults := defaultStorefrontSections()

	allowed := make(map[entity.StorefrontSection]struct{}, len(defaults))
	for _, item := range defaults {
		allowed[item] = struct{}{}
	}

	used := make(map[entity.StorefrontSection]struct{}, len(defaults))
	result := make([]entity.StorefrontSection, 0, len(defaults))
	for _, item := range items {
		if _, ok := allowed[item]; !ok {
			continue
		}
		if _, exists := used[item]; exists {
			continue
		}
		used[item] = struct{}{}
		result = append(result, item)
	}

	for _, item := range defaults {
		if _, exists := used[item]; exists {
			continue
		}
		result = append(result, item)
	}

	return result
}

func sanitizeSectionVisibility(items []entity.StorefrontSection) []entity.StorefrontSection {
	if items == nil {
		return defaultStorefrontSections()
	}

	allowedItems := defaultStorefrontSections()
	allowed := make(map[entity.StorefrontSection]struct{}, len(allowedItems))
	for _, item := range allowedItems {
		allowed[item] = struct{}{}
	}

	used := make(map[entity.StorefrontSection]struct{}, len(items))
	result := make([]entity.StorefrontSection, 0, len(items))
	for _, item := range items {
		if _, ok := allowed[item]; !ok {
			continue
		}
		if _, exists := used[item]; exists {
			continue
		}
		used[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

func defaultStorefrontSections() []entity.StorefrontSection {
	return []entity.StorefrontSection{
		entity.StorefrontSectionBanner,
		entity.StorefrontSectionCustomBlock,
		entity.StorefrontSectionLaunches,
		entity.StorefrontSectionFeatured,
		entity.StorefrontSectionInstitutional,
		entity.StorefrontSectionPromotions,
	}
}

func normalizeHexColor(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	matched, _ := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, trimmed)
	if !matched {
		return fallback
	}
	return strings.ToUpper(trimmed)
}

func normalizeCornerStyle(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "sharp":
		return "sharp"
	case "soft":
		return "soft"
	default:
		return "accentuated"
	}
}

func normalizeContentWidthMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "full_width":
		return "full_width"
	default:
		return "contained"
	}
}

func normalizeBannerWidthMode(value string, contentWidthMode string) string {
	if contentWidthMode == "full_width" {
		return "full_width"
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "full_width":
		return "full_width"
	default:
		return "contained"
	}
}

func normalizeBannerEffectMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none":
		return "none"
	case "fade":
		return "fade"
	case "zoom":
		return "zoom"
	default:
		return "slider"
	}
}

func normalizeBannerRotationSeconds(value int) int {
	switch {
	case value >= 2 && value <= 15:
		return value
	default:
		return 5
	}
}

func normalizeBrandDisplayMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "logo_only":
		return "logo_only"
	default:
		return "logo_and_name"
	}
}

func normalizeBrandLogoSize(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "medium":
		return "medium"
	case "large":
		return "large"
	default:
		return "small"
	}
}

func normalizeMenuBackgroundMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "transparent":
		return "transparent"
	default:
		return "solid"
	}
}

func normalizeButtonStyle(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "solid":
		return "solid"
	default:
		return "gradient"
	}
}

func normalizeFontFamily(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeGoogleFontImportURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	stylesheetMatch := regexp.MustCompile(`https://fonts\.googleapis\.com/[^"')\s]+`).FindString(trimmed)
	if stylesheetMatch != "" {
		trimmed = strings.TrimSpace(stylesheetMatch)
	} else {
		hrefMatches := regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`).FindAllStringSubmatch(trimmed, -1)
		for _, match := range hrefMatches {
			if len(match) == 2 && strings.HasPrefix(strings.ToLower(strings.TrimSpace(match[1])), "https://fonts.googleapis.com/") {
				trimmed = strings.TrimSpace(match[1])
				break
			}
		}
	}

	trimmed = strings.TrimPrefix(trimmed, "url(")
	trimmed = strings.TrimSuffix(trimmed, ")")
	trimmed = strings.Trim(trimmed, `"'`)

	if !strings.HasPrefix(strings.ToLower(trimmed), "https://fonts.googleapis.com/") {
		return ""
	}

	return trimmed
}

func normalizeProductsPerRowDesktop(value int, contentWidthMode string) int {
	if contentWidthMode == "full_width" {
		switch value {
		case 3, 4, 5:
			return value
		default:
			return 5
		}
	}

	switch value {
	case 3:
		return 3
	default:
		return 4
	}
}

func normalizeProductsPerRowMobile(value int) int {
	switch value {
	case 2:
		return 2
	default:
		return 1
	}
}

func normalizeProductSectionConfig(
	value entity.ProductSectionConfig,
	defaults entity.ProductSectionConfig,
) entity.ProductSectionConfig {
	config := entity.ProductSectionConfig{
		Eyebrow:     strings.TrimSpace(value.Eyebrow),
		Title:       strings.TrimSpace(value.Title),
		Description: strings.TrimSpace(value.Description),
		DisplayMode: normalizeProductSectionDisplayMode(value.DisplayMode),
	}

	if config.Title == "" {
		config.Title = defaults.Title
	}
	if config.DisplayMode == "" {
		config.DisplayMode = defaults.DisplayMode
	}

	return config
}

func normalizeInstitutionalSectionConfig(value entity.InstitutionalSectionConfig) entity.InstitutionalSectionConfig {
	return entity.InstitutionalSectionConfig{
		Eyebrow:         strings.TrimSpace(value.Eyebrow),
		Title:           strings.TrimSpace(value.Title),
		Description:     strings.TrimSpace(value.Description),
		DisplayMode:     normalizeInstitutionalSectionDisplayMode(value.DisplayMode),
		WidthMode:       normalizeInstitutionalSectionWidthMode(value.WidthMode),
		BackgroundColor: normalizeHexColor(value.BackgroundColor, "#0F172A"),
	}
}

func normalizeInstitutionalSectionDisplayMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "continuous":
		return "continuous"
	default:
		return "cards"
	}
}

func normalizeInstitutionalSectionWidthMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "full_width":
		return "full_width"
	default:
		return "contained"
	}
}

func normalizeProductSectionDisplayMode(value entity.ProductSectionDisplayMode) entity.ProductSectionDisplayMode {
	switch strings.TrimSpace(string(value)) {
	case string(entity.ProductSectionDisplayModeHorizontalScroll):
		return entity.ProductSectionDisplayModeHorizontalScroll
	default:
		return entity.ProductSectionDisplayModeWrap
	}
}

func defaultLaunchesSectionConfig() entity.ProductSectionConfig {
	return entity.ProductSectionConfig{
		Eyebrow:     "",
		Title:       "Novidades da loja",
		Description: "",
		DisplayMode: entity.ProductSectionDisplayModeWrap,
	}
}

func defaultFeaturedSectionConfig() entity.ProductSectionConfig {
	return entity.ProductSectionConfig{
		Eyebrow:     "",
		Title:       "Títulos em evidência",
		Description: "",
		DisplayMode: entity.ProductSectionDisplayModeWrap,
	}
}

func defaultPromotionsSectionConfig() entity.ProductSectionConfig {
	return entity.ProductSectionConfig{
		Eyebrow:     "",
		Title:       "Ofertas e descontos",
		Description: "",
		DisplayMode: entity.ProductSectionDisplayModeWrap,
	}
}

func normalizeTextAlign(value string) string {
	switch strings.TrimSpace(value) {
	case "center":
		return "center"
	case "right":
		return "right"
	default:
		return "left"
	}
}

func normalizeIconSize(value string) string {
	switch strings.TrimSpace(value) {
	case "small":
		return "small"
	case "large":
		return "large"
	default:
		return "medium"
	}
}

func normalizeLoadedSettingsRecord(record entity.SettingsRecord) entity.SettingsRecord {
	record.ContentWidthMode = normalizeContentWidthMode(record.ContentWidthMode)
	record.BannerWidthMode = normalizeBannerWidthMode(record.BannerWidthMode, record.ContentWidthMode)
	record.BannerEffectMode = normalizeBannerEffectMode(record.BannerEffectMode)
	record.BannerRotationSeconds = normalizeBannerRotationSeconds(record.BannerRotationSeconds)
	record.CornerStyle = normalizeCornerStyle(record.CornerStyle)
	record.BrandDisplayMode = normalizeBrandDisplayMode(record.BrandDisplayMode)
	record.BrandLogoSize = normalizeBrandLogoSize(record.BrandLogoSize)
	record.HeaderColor = normalizeHexColor(record.HeaderColor, "#020617")
	record.HeaderTextColor = normalizeHexColor(record.HeaderTextColor, "#F8FAFC")
	record.LaunchBadgeColor = normalizeHexColor(record.LaunchBadgeColor, "#1D4ED8")
	record.FeaturedBadgeColor = normalizeHexColor(record.FeaturedBadgeColor, "#166534")
	record.PromotionBadgeColor = normalizeHexColor(record.PromotionBadgeColor, "#C2410C")
	record.ButtonColor = normalizeHexColor(record.ButtonColor, "#7C3AED")
	record.ButtonTextColor = normalizeHexColor(record.ButtonTextColor, "#F8FAFC")
	record.ButtonStyle = normalizeButtonStyle(record.ButtonStyle)
	record.FontFamily = normalizeFontFamily(record.FontFamily)
	record.FontSecondaryFamily = normalizeFontFamily(record.FontSecondaryFamily)
	record.FontMenuFamily = normalizeFontFamily(record.FontMenuFamily)
	record.FontButtonFamily = normalizeFontFamily(record.FontButtonFamily)
	record.FontHighlightFamily = normalizeFontFamily(record.FontHighlightFamily)
	record.FontImportURL = normalizeGoogleFontImportURL(record.FontImportURL)
	record.FontEmbedCode = strings.TrimSpace(record.FontEmbedCode)
	record.Integrations = normalizeIntegrationSettings(record.Integrations)
	if record.FontSecondaryFamily == "" {
		record.FontSecondaryFamily = record.FontFamily
	}
	if record.FontMenuFamily == "" {
		record.FontMenuFamily = firstNonEmpty(record.FontSecondaryFamily, record.FontFamily)
	}
	if record.FontButtonFamily == "" {
		record.FontButtonFamily = record.FontFamily
	}
	if record.FontHighlightFamily == "" {
		record.FontHighlightFamily = firstNonEmpty(record.FontSecondaryFamily, record.FontFamily)
	}
	record.MenuBackgroundMode = normalizeMenuBackgroundMode(record.MenuBackgroundMode)
	record.ProductsPerRowDesktop = normalizeProductsPerRowDesktop(record.ProductsPerRowDesktop, record.ContentWidthMode)
	record.ProductsPerRowMobile = normalizeProductsPerRowMobile(record.ProductsPerRowMobile)
	record.MenuLinks = normalizeLoadedMenuLinks(record.MenuLinks)
	record.SectionOrder = sanitizeSectionOrder(record.SectionOrder)
	record.VisibleSections = sanitizeSectionVisibility(record.VisibleSections)
	record.LaunchesSection = normalizeProductSectionConfig(record.LaunchesSection, defaultLaunchesSectionConfig())
	record.FeaturedSection = normalizeProductSectionConfig(record.FeaturedSection, defaultFeaturedSectionConfig())
	record.PromotionsSection = normalizeProductSectionConfig(record.PromotionsSection, defaultPromotionsSectionConfig())
	record.InstitutionalSection = normalizeInstitutionalSectionConfig(record.InstitutionalSection)
	record.FeatureHighlights = sanitizeFeatureHighlights(record.FeatureHighlights)
	return record
}

func normalizeIntegrationSettings(settings entity.IntegrationSettings) entity.IntegrationSettings {
	return entity.IntegrationSettings{
		FacebookPixelID:          strings.TrimSpace(settings.FacebookPixelID),
		GoogleAdsID:              strings.TrimSpace(settings.GoogleAdsID),
		GoogleAdsConversionLabel: strings.TrimSpace(settings.GoogleAdsConversionLabel),
		GoogleAnalyticsID:        strings.TrimSpace(settings.GoogleAnalyticsID),
		GoogleTagManagerID:       strings.TrimSpace(settings.GoogleTagManagerID),
		MicrosoftClarityID:       strings.TrimSpace(settings.MicrosoftClarityID),
		TikTokPixelID:            strings.TrimSpace(settings.TikTokPixelID),
	}
}

func normalizeLoadedMenuLinks(items []entity.NavigationLink) []entity.NavigationLink {
	if len(items) == 0 {
		return []entity.NavigationLink{}
	}

	normalized := make([]entity.NavigationLink, 0, len(items))
	usedKinds := map[string]struct{}{}
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		url := strings.TrimSpace(item.URL)
		if label == "" || url == "" {
			continue
		}

		kind := normalizeNavigationLinkKind(item.Kind, label, url)
		visible := item.Visible
		if strings.TrimSpace(item.Kind) == "" {
			visible = true
		}

		if kind != "custom" {
			if _, exists := usedKinds[kind]; exists {
				continue
			}
			usedKinds[kind] = struct{}{}
			label, url = defaultNavigationLinkValues(kind)
		}

		normalized = append(normalized, entity.NavigationLink{
			Label:   label,
			URL:     url,
			Visible: visible,
			Kind:    kind,
		})
	}

	for _, kind := range []string{"home", "products", "categories", "contact"} {
		if _, exists := usedKinds[kind]; exists {
			continue
		}
		label, url := defaultNavigationLinkValues(kind)
		normalized = append(normalized, entity.NavigationLink{
			Label:   label,
			URL:     url,
			Visible: true,
			Kind:    kind,
		})
	}

	return normalized
}

func normalizeNavigationLinkKind(value string, label string, url string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "home":
		return "home"
	case "products":
		return "products"
	case "categories":
		return "categories"
	case "contact":
		return "contact"
	default:
		return inferNavigationLinkKind(label, url)
	}
}

func inferNavigationLinkKind(label string, url string) string {
	normalizedLabel := strings.TrimSpace(strings.ToLower(label))
	normalizedURL := strings.TrimSpace(strings.ToLower(url))
	switch {
	case normalizedLabel == "início" || normalizedLabel == "inicio" || normalizedURL == "#topo":
		return "home"
	case normalizedLabel == "produtos" || normalizedURL == "/produtos":
		return "products"
	case normalizedLabel == "categorias" || normalizedURL == "__categories__":
		return "categories"
	case normalizedLabel == "contato" || normalizedURL == "#rodape":
		return "contact"
	default:
		return "custom"
	}
}

func defaultNavigationLinkValues(kind string) (string, string) {
	switch kind {
	case "home":
		return "Início", "#topo"
	case "products":
		return "Produtos", "/produtos"
	case "categories":
		return "Categorias", "__categories__"
	case "contact":
		return "Contato", "#rodape"
	default:
		return "", ""
	}
}

func normalizeSlug(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return normalized
}
