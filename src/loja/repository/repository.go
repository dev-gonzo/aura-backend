package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sistema-editorial/editora/backend/src/loja/entity"
)

type Repository struct {
	pool *pgxpool.Pool
}

type draftSettingsPayload struct {
	StoreName             string                      `json:"store_name"`
	StoreSubtitle         string                      `json:"store_subtitle"`
	StoreDescription      string                      `json:"store_description"`
	StoreTags             string                      `json:"store_tags"`
	StoreInMaintenance    bool                        `json:"store_in_maintenance"`
	MaintenanceTitle      string                      `json:"maintenance_title"`
	MaintenanceMessage    string                      `json:"maintenance_message"`
	PrimaryColor          string                      `json:"primary_color"`
	AccentColor           string                      `json:"accent_color"`
	SecondaryColor        string                      `json:"secondary_color"`
	BackgroundColor       string                      `json:"background_color"`
	HeaderColor           string                      `json:"header_color"`
	MenuColor             string                      `json:"menu_color"`
	FooterColor           string                      `json:"footer_color"`
	HeaderTextColor       string                      `json:"header_text_color"`
	MenuTextColor         string                      `json:"menu_text_color"`
	FooterTextColor       string                      `json:"footer_text_color"`
	PrimaryTextColor      string                      `json:"primary_text_color"`
	HighlightColor        string                      `json:"highlight_color"`
	DiscountColor         string                      `json:"discount_color"`
	LaunchBadgeColor      string                      `json:"launch_badge_color"`
	FeaturedBadgeColor    string                      `json:"featured_badge_color"`
	PromotionBadgeColor   string                      `json:"promotion_badge_color"`
	ButtonColor           string                      `json:"button_color"`
	ButtonTextColor       string                      `json:"button_text_color"`
	ButtonStyle           string                      `json:"button_style"`
	FontFamily            string                      `json:"font_family"`
	FontSecondaryFamily   string                      `json:"font_secondary_family"`
	FontMenuFamily        string                      `json:"font_menu_family"`
	FontButtonFamily      string                      `json:"font_button_family"`
	FontHighlightFamily   string                      `json:"font_highlight_family"`
	FontImportURL         string                      `json:"font_import_url"`
	FontEmbedCode         string                      `json:"font_embed_code"`
	Integrations          entity.IntegrationSettings  `json:"integrations"`
	CornerStyle           string                      `json:"corner_style"`
	ContentWidthMode      string                      `json:"content_width_mode"`
	MenuBackgroundMode    string                      `json:"menu_background_mode"`
	BannerWidthMode       string                      `json:"banner_width_mode"`
	BannerEffectMode      string                      `json:"banner_effect_mode"`
	BannerRotationSeconds int                         `json:"banner_rotation_seconds"`
	BrandDisplayMode      string                      `json:"brand_display_mode"`
	BrandLogoSize         string                      `json:"brand_logo_size"`
	ProductsPerRowDesktop int                         `json:"products_per_row_desktop"`
	ProductsPerRowMobile  int                         `json:"products_per_row_mobile"`
	HeroTitle             string                      `json:"hero_title"`
	HeroSubtitle          string                      `json:"hero_subtitle"`
	Logo                  *entity.ImageAsset          `json:"logo,omitempty"`
	Favicon               *entity.ImageAsset          `json:"favicon,omitempty"`
	BannerDesktop         *entity.ImageAsset          `json:"banner_desktop,omitempty"`
	BannerMobile          *entity.ImageAsset          `json:"banner_mobile,omitempty"`
	Banners               []entity.BannerPayload      `json:"banners"`
	MenuLinks             []entity.NavigationLink     `json:"menu_links"`
	SectionOrder          []entity.StorefrontSection  `json:"section_order"`
	VisibleSections       []entity.StorefrontSection  `json:"visible_sections"`
	CustomBlockEyebrow    string                      `json:"custom_block_eyebrow"`
	CustomBlockTitle      string                      `json:"custom_block_title"`
	CustomBlockDescription string                     `json:"custom_block_description"`
	CustomBlockHTML       string                      `json:"custom_block_html"`
	CustomBlockCSS        string                      `json:"custom_block_css"`
	CustomBlockJS         string                      `json:"custom_block_js"`
	FeatureHighlights     []entity.FeatureHighlight   `json:"feature_highlights"`
	InstitutionalSection  entity.InstitutionalSectionConfig `json:"institutional_section"`
	ProductListingConfig  entity.ProductListingConfig `json:"product_listing_config"`
	FooterLinks           []entity.NavigationLink     `json:"footer_links"`
	FooterContactTitle    string                      `json:"footer_contact_title"`
	FooterContactText     string                      `json:"footer_contact_text"`
	LaunchesSection       entity.ProductSectionConfig `json:"launches_section"`
	FeaturedSection       entity.ProductSectionConfig `json:"featured_section"`
	PromotionsSection     entity.ProductSectionConfig `json:"promotions_section"`
}

type legacyDraftSettingsPayload struct {
	StoreName             string                      `json:"StoreName"`
	StoreSubtitle         string                      `json:"StoreSubtitle"`
	StoreDescription      string                      `json:"StoreDescription"`
	StoreTags             string                      `json:"StoreTags"`
	StoreInMaintenance    bool                        `json:"StoreInMaintenance"`
	MaintenanceTitle      string                      `json:"MaintenanceTitle"`
	MaintenanceMessage    string                      `json:"MaintenanceMessage"`
	PrimaryColor          string                      `json:"PrimaryColor"`
	AccentColor           string                      `json:"AccentColor"`
	SecondaryColor        string                      `json:"SecondaryColor"`
	BackgroundColor       string                      `json:"BackgroundColor"`
	HeaderColor           string                      `json:"HeaderColor"`
	MenuColor             string                      `json:"MenuColor"`
	FooterColor           string                      `json:"FooterColor"`
	HeaderTextColor       string                      `json:"HeaderTextColor"`
	MenuTextColor         string                      `json:"MenuTextColor"`
	FooterTextColor       string                      `json:"FooterTextColor"`
	PrimaryTextColor      string                      `json:"PrimaryTextColor"`
	HighlightColor        string                      `json:"HighlightColor"`
	DiscountColor         string                      `json:"DiscountColor"`
	LaunchBadgeColor      string                      `json:"LaunchBadgeColor"`
	FeaturedBadgeColor    string                      `json:"FeaturedBadgeColor"`
	PromotionBadgeColor   string                      `json:"PromotionBadgeColor"`
	ButtonColor           string                      `json:"ButtonColor"`
	ButtonTextColor       string                      `json:"ButtonTextColor"`
	ButtonStyle           string                      `json:"ButtonStyle"`
	FontFamily            string                      `json:"FontFamily"`
	FontSecondaryFamily   string                      `json:"FontSecondaryFamily"`
	FontMenuFamily        string                      `json:"FontMenuFamily"`
	FontButtonFamily      string                      `json:"FontButtonFamily"`
	FontHighlightFamily   string                      `json:"FontHighlightFamily"`
	FontImportURL         string                      `json:"FontImportURL"`
	FontEmbedCode         string                      `json:"FontEmbedCode"`
	Integrations          entity.IntegrationSettings  `json:"Integrations"`
	CornerStyle           string                      `json:"CornerStyle"`
	ContentWidthMode      string                      `json:"ContentWidthMode"`
	MenuBackgroundMode    string                      `json:"MenuBackgroundMode"`
	BannerWidthMode       string                      `json:"BannerWidthMode"`
	BannerEffectMode      string                      `json:"BannerEffectMode"`
	BannerRotationSeconds int                         `json:"BannerRotationSeconds"`
	BrandDisplayMode      string                      `json:"BrandDisplayMode"`
	BrandLogoSize         string                      `json:"BrandLogoSize"`
	ProductsPerRowDesktop int                         `json:"ProductsPerRowDesktop"`
	ProductsPerRowMobile  int                         `json:"ProductsPerRowMobile"`
	HeroTitle             string                      `json:"HeroTitle"`
	HeroSubtitle          string                      `json:"HeroSubtitle"`
	Logo                  *entity.ImageAsset          `json:"Logo,omitempty"`
	Favicon               *entity.ImageAsset          `json:"Favicon,omitempty"`
	BannerDesktop         *entity.ImageAsset          `json:"BannerDesktop,omitempty"`
	BannerMobile          *entity.ImageAsset          `json:"BannerMobile,omitempty"`
	Banners               []entity.BannerPayload      `json:"Banners"`
	MenuLinks             []entity.NavigationLink     `json:"MenuLinks"`
	SectionOrder          []entity.StorefrontSection  `json:"SectionOrder"`
	VisibleSections       []entity.StorefrontSection  `json:"VisibleSections"`
	CustomBlockEyebrow    string                      `json:"CustomBlockEyebrow"`
	CustomBlockTitle      string                      `json:"CustomBlockTitle"`
	CustomBlockDescription string                     `json:"CustomBlockDescription"`
	CustomBlockHTML       string                      `json:"CustomBlockHTML"`
	CustomBlockCSS        string                      `json:"CustomBlockCSS"`
	CustomBlockJS         string                      `json:"CustomBlockJS"`
	FeatureHighlights     []entity.FeatureHighlight   `json:"FeatureHighlights"`
	InstitutionalSection  entity.InstitutionalSectionConfig `json:"InstitutionalSection"`
	ProductListingConfig  entity.ProductListingConfig `json:"ProductListingConfig"`
	FooterLinks           []entity.NavigationLink     `json:"FooterLinks"`
	FooterContactTitle    string                      `json:"FooterContactTitle"`
	FooterContactText     string                      `json:"FooterContactText"`
	LaunchesSection       entity.ProductSectionConfig `json:"LaunchesSection"`
	FeaturedSection       entity.ProductSectionConfig `json:"FeaturedSection"`
	PromotionsSection     entity.ProductSectionConfig `json:"PromotionsSection"`
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func newDraftSettingsPayload(input entity.SettingsRecord) draftSettingsPayload {
	return draftSettingsPayload{
		StoreName:             input.StoreName,
		StoreSubtitle:         input.StoreSubtitle,
		StoreDescription:      input.StoreDescription,
		StoreTags:             input.StoreTags,
		StoreInMaintenance:    input.StoreInMaintenance,
		MaintenanceTitle:      input.MaintenanceTitle,
		MaintenanceMessage:    input.MaintenanceMessage,
		PrimaryColor:          input.PrimaryColor,
		AccentColor:           input.AccentColor,
		SecondaryColor:        input.SecondaryColor,
		BackgroundColor:       input.BackgroundColor,
		HeaderColor:           input.HeaderColor,
		MenuColor:             input.MenuColor,
		FooterColor:           input.FooterColor,
		HeaderTextColor:       input.HeaderTextColor,
		MenuTextColor:         input.MenuTextColor,
		FooterTextColor:       input.FooterTextColor,
		PrimaryTextColor:      input.PrimaryTextColor,
		HighlightColor:        input.HighlightColor,
		DiscountColor:         input.DiscountColor,
		LaunchBadgeColor:      input.LaunchBadgeColor,
		FeaturedBadgeColor:    input.FeaturedBadgeColor,
		PromotionBadgeColor:   input.PromotionBadgeColor,
		ButtonColor:           input.ButtonColor,
		ButtonTextColor:       input.ButtonTextColor,
		ButtonStyle:           input.ButtonStyle,
		FontFamily:            input.FontFamily,
		FontSecondaryFamily:   input.FontSecondaryFamily,
		FontMenuFamily:        input.FontMenuFamily,
		FontButtonFamily:      input.FontButtonFamily,
		FontHighlightFamily:   input.FontHighlightFamily,
		FontImportURL:         input.FontImportURL,
		FontEmbedCode:         input.FontEmbedCode,
		Integrations:          input.Integrations,
		CornerStyle:           input.CornerStyle,
		ContentWidthMode:      input.ContentWidthMode,
		MenuBackgroundMode:    input.MenuBackgroundMode,
		BannerWidthMode:       input.BannerWidthMode,
		BannerEffectMode:      input.BannerEffectMode,
		BannerRotationSeconds: input.BannerRotationSeconds,
		BrandDisplayMode:      input.BrandDisplayMode,
		BrandLogoSize:         input.BrandLogoSize,
		ProductsPerRowDesktop: input.ProductsPerRowDesktop,
		ProductsPerRowMobile:  input.ProductsPerRowMobile,
		HeroTitle:             input.HeroTitle,
		HeroSubtitle:          input.HeroSubtitle,
		Logo:                  input.Logo,
		Favicon:               input.Favicon,
		BannerDesktop:         input.BannerDesktop,
		BannerMobile:          input.BannerMobile,
		Banners:               input.Banners,
		MenuLinks:             input.MenuLinks,
		SectionOrder:          input.SectionOrder,
		VisibleSections:       input.VisibleSections,
		CustomBlockEyebrow:    input.CustomBlockEyebrow,
		CustomBlockTitle:      input.CustomBlockTitle,
		CustomBlockDescription: input.CustomBlockDescription,
		CustomBlockHTML:       input.CustomBlockHTML,
		CustomBlockCSS:        input.CustomBlockCSS,
		CustomBlockJS:         input.CustomBlockJS,
		FeatureHighlights:     input.FeatureHighlights,
		InstitutionalSection:  input.InstitutionalSection,
		ProductListingConfig:  input.ProductListingConfig,
		FooterLinks:           input.FooterLinks,
		FooterContactTitle:    input.FooterContactTitle,
		FooterContactText:     input.FooterContactText,
		LaunchesSection:       input.LaunchesSection,
		FeaturedSection:       input.FeaturedSection,
		PromotionsSection:     input.PromotionsSection,
	}
}

func (p draftSettingsPayload) toSettingsRecord() entity.SettingsRecord {
	return entity.SettingsRecord{
		StoreName:             p.StoreName,
		StoreSubtitle:         p.StoreSubtitle,
		StoreDescription:      p.StoreDescription,
		StoreTags:             p.StoreTags,
		StoreInMaintenance:    p.StoreInMaintenance,
		MaintenanceTitle:      p.MaintenanceTitle,
		MaintenanceMessage:    p.MaintenanceMessage,
		PrimaryColor:          p.PrimaryColor,
		AccentColor:           p.AccentColor,
		SecondaryColor:        p.SecondaryColor,
		BackgroundColor:       p.BackgroundColor,
		HeaderColor:           p.HeaderColor,
		MenuColor:             p.MenuColor,
		FooterColor:           p.FooterColor,
		HeaderTextColor:       p.HeaderTextColor,
		MenuTextColor:         p.MenuTextColor,
		FooterTextColor:       p.FooterTextColor,
		PrimaryTextColor:      p.PrimaryTextColor,
		HighlightColor:        p.HighlightColor,
		DiscountColor:         p.DiscountColor,
		LaunchBadgeColor:      p.LaunchBadgeColor,
		FeaturedBadgeColor:    p.FeaturedBadgeColor,
		PromotionBadgeColor:   p.PromotionBadgeColor,
		ButtonColor:           p.ButtonColor,
		ButtonTextColor:       p.ButtonTextColor,
		ButtonStyle:           p.ButtonStyle,
		FontFamily:            p.FontFamily,
		FontSecondaryFamily:   p.FontSecondaryFamily,
		FontMenuFamily:        p.FontMenuFamily,
		FontButtonFamily:      p.FontButtonFamily,
		FontHighlightFamily:   p.FontHighlightFamily,
		FontImportURL:         p.FontImportURL,
		FontEmbedCode:         p.FontEmbedCode,
		Integrations:          p.Integrations,
		CornerStyle:           p.CornerStyle,
		ContentWidthMode:      p.ContentWidthMode,
		MenuBackgroundMode:    p.MenuBackgroundMode,
		BannerWidthMode:       p.BannerWidthMode,
		BannerEffectMode:      p.BannerEffectMode,
		BannerRotationSeconds: p.BannerRotationSeconds,
		BrandDisplayMode:      p.BrandDisplayMode,
		BrandLogoSize:         p.BrandLogoSize,
		ProductsPerRowDesktop: p.ProductsPerRowDesktop,
		ProductsPerRowMobile:  p.ProductsPerRowMobile,
		HeroTitle:             p.HeroTitle,
		HeroSubtitle:          p.HeroSubtitle,
		Logo:                  p.Logo,
		Favicon:               p.Favicon,
		BannerDesktop:         p.BannerDesktop,
		BannerMobile:          p.BannerMobile,
		Banners:               p.Banners,
		MenuLinks:             p.MenuLinks,
		SectionOrder:          p.SectionOrder,
		VisibleSections:       p.VisibleSections,
		CustomBlockEyebrow:    p.CustomBlockEyebrow,
		CustomBlockTitle:      p.CustomBlockTitle,
		CustomBlockDescription: p.CustomBlockDescription,
		CustomBlockHTML:       p.CustomBlockHTML,
		CustomBlockCSS:        p.CustomBlockCSS,
		CustomBlockJS:         p.CustomBlockJS,
		FeatureHighlights:     p.FeatureHighlights,
		InstitutionalSection:  p.InstitutionalSection,
		ProductListingConfig:  p.ProductListingConfig,
		FooterLinks:           p.FooterLinks,
		FooterContactTitle:    p.FooterContactTitle,
		FooterContactText:     p.FooterContactText,
		LaunchesSection:       p.LaunchesSection,
		FeaturedSection:       p.FeaturedSection,
		PromotionsSection:     p.PromotionsSection,
	}
}

func (p legacyDraftSettingsPayload) toSettingsRecord() entity.SettingsRecord {
	return entity.SettingsRecord{
		StoreName:             p.StoreName,
		StoreSubtitle:         p.StoreSubtitle,
		StoreDescription:      p.StoreDescription,
		StoreTags:             p.StoreTags,
		StoreInMaintenance:    p.StoreInMaintenance,
		MaintenanceTitle:      p.MaintenanceTitle,
		MaintenanceMessage:    p.MaintenanceMessage,
		PrimaryColor:          p.PrimaryColor,
		AccentColor:           p.AccentColor,
		SecondaryColor:        p.SecondaryColor,
		BackgroundColor:       p.BackgroundColor,
		HeaderColor:           p.HeaderColor,
		MenuColor:             p.MenuColor,
		FooterColor:           p.FooterColor,
		HeaderTextColor:       p.HeaderTextColor,
		MenuTextColor:         p.MenuTextColor,
		FooterTextColor:       p.FooterTextColor,
		PrimaryTextColor:      p.PrimaryTextColor,
		HighlightColor:        p.HighlightColor,
		DiscountColor:         p.DiscountColor,
		LaunchBadgeColor:      p.LaunchBadgeColor,
		FeaturedBadgeColor:    p.FeaturedBadgeColor,
		PromotionBadgeColor:   p.PromotionBadgeColor,
		ButtonColor:           p.ButtonColor,
		ButtonTextColor:       p.ButtonTextColor,
		ButtonStyle:           p.ButtonStyle,
		FontFamily:            p.FontFamily,
		FontSecondaryFamily:   p.FontSecondaryFamily,
		FontMenuFamily:        p.FontMenuFamily,
		FontButtonFamily:      p.FontButtonFamily,
		FontHighlightFamily:   p.FontHighlightFamily,
		FontImportURL:         p.FontImportURL,
		FontEmbedCode:         p.FontEmbedCode,
		Integrations:          p.Integrations,
		CornerStyle:           p.CornerStyle,
		ContentWidthMode:      p.ContentWidthMode,
		MenuBackgroundMode:    p.MenuBackgroundMode,
		BannerWidthMode:       p.BannerWidthMode,
		BannerEffectMode:      p.BannerEffectMode,
		BannerRotationSeconds: p.BannerRotationSeconds,
		BrandDisplayMode:      p.BrandDisplayMode,
		BrandLogoSize:         p.BrandLogoSize,
		ProductsPerRowDesktop: p.ProductsPerRowDesktop,
		ProductsPerRowMobile:  p.ProductsPerRowMobile,
		HeroTitle:             p.HeroTitle,
		HeroSubtitle:          p.HeroSubtitle,
		Logo:                  p.Logo,
		Favicon:               p.Favicon,
		BannerDesktop:         p.BannerDesktop,
		BannerMobile:          p.BannerMobile,
		Banners:               p.Banners,
		MenuLinks:             p.MenuLinks,
		SectionOrder:          p.SectionOrder,
		VisibleSections:       p.VisibleSections,
		CustomBlockEyebrow:    p.CustomBlockEyebrow,
		CustomBlockTitle:      p.CustomBlockTitle,
		CustomBlockDescription: p.CustomBlockDescription,
		CustomBlockHTML:       p.CustomBlockHTML,
		CustomBlockCSS:        p.CustomBlockCSS,
		CustomBlockJS:         p.CustomBlockJS,
		FeatureHighlights:     p.FeatureHighlights,
		InstitutionalSection:  p.InstitutionalSection,
		ProductListingConfig:  p.ProductListingConfig,
		FooterLinks:           p.FooterLinks,
		FooterContactTitle:    p.FooterContactTitle,
		FooterContactText:     p.FooterContactText,
		LaunchesSection:       p.LaunchesSection,
		FeaturedSection:       p.FeaturedSection,
		PromotionsSection:     p.PromotionsSection,
	}
}

func (r *Repository) GetSettings(ctx context.Context) (entity.SettingsRecord, error) {
	const query = `
		SELECT
			nome_loja,
			coalesce(subtitulo_loja, ''),
			coalesce(descricao_loja, ''),
			coalesce(tags_loja, ''),
		coalesce(loja_em_construcao, false),
		coalesce(titulo_manutencao, ''),
		coalesce(mensagem_manutencao, ''),
			cor_principal,
			cor_primaria,
			cor_secundaria,
			cor_fundo,
			coalesce(cor_cabecalho, '#020617'),
			cor_menu,
			cor_rodape,
			coalesce(cor_texto_cabecalho, '#F8FAFC'),
			cor_texto_menu,
			cor_texto_rodape,
			cor_texto_principal,
			cor_destaque,
			cor_desconto,
			coalesce(cor_selo_lancamento, '#1D4ED8'),
			coalesce(cor_selo_destaque, '#166534'),
			coalesce(cor_selo_promocao, '#C2410C'),
			coalesce(cor_botao, '#7C3AED'),
			coalesce(cor_texto_botao, '#F8FAFC'),
			coalesce(estilo_botao, 'gradient'),
			coalesce(fonte_familia, ''),
			coalesce(fonte_secundaria_familia, ''),
			coalesce(fonte_menu_familia, ''),
			coalesce(fonte_botao_familia, ''),
			coalesce(fonte_destaque_familia, ''),
			coalesce(fonte_import_url, ''),
			coalesce(fonte_embed_code, ''),
			coalesce(integracoes_json, '{}'),
			coalesce(estilo_arredondamento, 'accentuated'),
			coalesce(modo_largura_layout, 'contained'),
			coalesce(modo_fundo_menu, 'solid'),
			coalesce(modo_largura_banner, 'contained'),
			coalesce(efeito_banner_principal, 'slider'),
			coalesce(tempo_rotacao_banner_segundos, 5),
			coalesce(modo_exibicao_marca, 'logo_and_name'),
			coalesce(produtos_por_linha_desktop, 4),
			coalesce(produtos_por_linha_mobile, 1),
			coalesce(hero_titulo, ''),
			coalesce(hero_subtitulo, ''),
			coalesce(menu_links_json, '[]'),
			coalesce(section_order_json, '[]'),
			coalesce(section_visibility_json, '[]'),
			coalesce(custom_block_eyebrow, ''),
			coalesce(custom_block_title, ''),
			coalesce(custom_block_description, ''),
			coalesce(custom_block_html, ''),
			coalesce(custom_block_css, ''),
			coalesce(custom_block_js, ''),
			coalesce(feature_highlights_json, '[]'),
			coalesce(institutional_section_json, ''),
			coalesce(product_listing_config_json, ''),
			coalesce(footer_links_json, '[]'),
			coalesce(footer_contact_title, ''),
			coalesce(footer_contact_text, ''),
			coalesce(launches_section_json, ''),
			coalesce(featured_section_json, ''),
			coalesce(promotions_section_json, ''),
			coalesce(logo_base64, ''), coalesce(logo_mime, ''), coalesce(logo_largura, 0), coalesce(logo_altura, 0), coalesce(logo_tamanho_bytes, 0), coalesce(logo_hash_sha256, ''),
			coalesce(favicon_base64, ''), coalesce(favicon_mime, ''), coalesce(favicon_largura, 0), coalesce(favicon_altura, 0), coalesce(favicon_tamanho_bytes, 0), coalesce(favicon_hash_sha256, ''),
			coalesce(banner_desktop_base64, ''), coalesce(banner_desktop_mime, ''), coalesce(banner_desktop_largura, 0), coalesce(banner_desktop_altura, 0), coalesce(banner_desktop_tamanho_bytes, 0), coalesce(banner_desktop_hash_sha256, ''),
			coalesce(banner_mobile_base64, ''), coalesce(banner_mobile_mime, ''), coalesce(banner_mobile_largura, 0), coalesce(banner_mobile_altura, 0), coalesce(banner_mobile_tamanho_bytes, 0), coalesce(banner_mobile_hash_sha256, ''),
			criado_em,
		atualizado_em,
		coalesce(published_at, atualizado_em),
		(draft_updated_at is not null)
		FROM loja_configuracoes
		WHERE id = 1
	`

	var record entity.SettingsRecord
	var logoBase64, logoMime, logoHash string
	var logoWidth, logoHeight, logoSize int
	var faviconBase64, faviconMime, faviconHash string
	var faviconWidth, faviconHeight, faviconSize int
	var bannerDesktopBase64, bannerDesktopMime, bannerDesktopHash string
	var bannerDesktopWidth, bannerDesktopHeight, bannerDesktopSize int
	var bannerMobileBase64, bannerMobileMime, bannerMobileHash string
	var bannerMobileWidth, bannerMobileHeight, bannerMobileSize int
	var menuLinksJSON, sectionOrderJSON, sectionVisibilityJSON, featureHighlightsJSON, institutionalSectionJSON, productListingConfigJSON, footerLinksJSON string
	var integrationsJSON string
	var launchesSectionJSON, featuredSectionJSON, promotionsSectionJSON string

	err := r.pool.QueryRow(ctx, query).Scan(
		&record.StoreName,
		&record.StoreSubtitle,
		&record.StoreDescription,
		&record.StoreTags,
		&record.StoreInMaintenance,
		&record.MaintenanceTitle,
		&record.MaintenanceMessage,
		&record.PrimaryColor,
		&record.AccentColor,
		&record.SecondaryColor,
		&record.BackgroundColor,
		&record.HeaderColor,
		&record.MenuColor,
		&record.FooterColor,
		&record.HeaderTextColor,
		&record.MenuTextColor,
		&record.FooterTextColor,
		&record.PrimaryTextColor,
		&record.HighlightColor,
		&record.DiscountColor,
		&record.LaunchBadgeColor,
		&record.FeaturedBadgeColor,
		&record.PromotionBadgeColor,
		&record.ButtonColor,
		&record.ButtonTextColor,
		&record.ButtonStyle,
		&record.FontFamily,
		&record.FontSecondaryFamily,
		&record.FontMenuFamily,
		&record.FontButtonFamily,
		&record.FontHighlightFamily,
		&record.FontImportURL,
		&record.FontEmbedCode,
		&integrationsJSON,
		&record.CornerStyle,
		&record.ContentWidthMode,
		&record.MenuBackgroundMode,
		&record.BannerWidthMode,
		&record.BannerEffectMode,
		&record.BannerRotationSeconds,
		&record.BrandDisplayMode,
		&record.ProductsPerRowDesktop,
		&record.ProductsPerRowMobile,
		&record.HeroTitle,
		&record.HeroSubtitle,
		&menuLinksJSON,
		&sectionOrderJSON,
		&sectionVisibilityJSON,
		&record.CustomBlockEyebrow,
		&record.CustomBlockTitle,
		&record.CustomBlockDescription,
		&record.CustomBlockHTML,
		&record.CustomBlockCSS,
		&record.CustomBlockJS,
		&featureHighlightsJSON,
		&institutionalSectionJSON,
		&productListingConfigJSON,
		&footerLinksJSON,
		&record.FooterContactTitle,
		&record.FooterContactText,
		&launchesSectionJSON,
		&featuredSectionJSON,
		&promotionsSectionJSON,
		&logoBase64, &logoMime, &logoWidth, &logoHeight, &logoSize, &logoHash,
		&faviconBase64, &faviconMime, &faviconWidth, &faviconHeight, &faviconSize, &faviconHash,
		&bannerDesktopBase64, &bannerDesktopMime, &bannerDesktopWidth, &bannerDesktopHeight, &bannerDesktopSize, &bannerDesktopHash,
		&bannerMobileBase64, &bannerMobileMime, &bannerMobileWidth, &bannerMobileHeight, &bannerMobileSize, &bannerMobileHash,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.PublishedAt,
		&record.HasUnpublishedChanges,
	)
	if err != nil {
		return entity.SettingsRecord{}, err
	}

	record.Logo = buildImageAsset(logoBase64, logoMime, logoWidth, logoHeight, logoSize, logoHash)
	record.Favicon = buildImageAsset(faviconBase64, faviconMime, faviconWidth, faviconHeight, faviconSize, faviconHash)
	record.BannerDesktop = buildImageAsset(
		bannerDesktopBase64,
		bannerDesktopMime,
		bannerDesktopWidth,
		bannerDesktopHeight,
		bannerDesktopSize,
		bannerDesktopHash,
	)
	record.BannerMobile = buildImageAsset(
		bannerMobileBase64,
		bannerMobileMime,
		bannerMobileWidth,
		bannerMobileHeight,
		bannerMobileSize,
		bannerMobileHash,
	)
	record.Integrations = unmarshalIntegrationSettings(integrationsJSON)
	record.MenuLinks = unmarshalNavigationLinks(menuLinksJSON)
	record.SectionOrder = unmarshalSectionOrder(sectionOrderJSON)
	record.VisibleSections = unmarshalSectionVisibility(sectionVisibilityJSON)
	record.FeatureHighlights = unmarshalFeatureHighlights(featureHighlightsJSON)
	record.InstitutionalSection = unmarshalInstitutionalSectionConfig(institutionalSectionJSON)
	record.ProductListingConfig = unmarshalProductListingConfig(productListingConfigJSON)
	record.FooterLinks = unmarshalNavigationLinks(footerLinksJSON)
	record.LaunchesSection = unmarshalProductSectionConfig(launchesSectionJSON)
	record.FeaturedSection = unmarshalProductSectionConfig(featuredSectionJSON)
	record.PromotionsSection = unmarshalProductSectionConfig(promotionsSectionJSON)

	return record, nil
}

func (r *Repository) GetDraftSettings(ctx context.Context) (entity.SettingsRecord, bool, error) {
	const query = `
		SELECT
			coalesce(draft_config_json, ''),
			coalesce(draft_updated_at, 'epoch'::timestamptz),
			coalesce(published_at, atualizado_em, NOW())
		FROM loja_configuracoes
		WHERE id = 1
	`

	var payload string
	var draftUpdatedAt, publishedAt time.Time
	if err := r.pool.QueryRow(ctx, query).Scan(&payload, &draftUpdatedAt, &publishedAt); err != nil {
		return entity.SettingsRecord{}, false, err
	}

	if strings.TrimSpace(payload) == "" {
		return entity.SettingsRecord{PublishedAt: publishedAt}, false, nil
	}

	var draftPayload draftSettingsPayload
	if err := json.Unmarshal([]byte(payload), &draftPayload); err == nil {
		record := draftPayload.toSettingsRecord()
		record.UpdatedAt = draftUpdatedAt
		record.DraftUpdatedAt = draftUpdatedAt
		record.PublishedAt = publishedAt
		record.HasUnpublishedChanges = true
		return record, true, nil
	}

	var legacyPayload legacyDraftSettingsPayload
	if err := json.Unmarshal([]byte(payload), &legacyPayload); err != nil {
		return entity.SettingsRecord{}, false, err
	}
	record := legacyPayload.toSettingsRecord()
	record.UpdatedAt = draftUpdatedAt
	record.DraftUpdatedAt = draftUpdatedAt
	record.PublishedAt = publishedAt
	record.HasUnpublishedChanges = true
	return record, true, nil
}

func (r *Repository) SaveDraftSettings(ctx context.Context, input entity.SettingsRecord) error {
	payload, err := json.Marshal(newDraftSettingsPayload(input))
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO loja_configuracoes (
			id,
			draft_config_json,
			draft_updated_at
		) VALUES (
			1,
			$1,
			NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			draft_config_json = EXCLUDED.draft_config_json,
			draft_updated_at = NOW()
	`

	_, err = r.pool.Exec(ctx, query, string(payload))
	return err
}

func (r *Repository) FinalizePublishedSettings(ctx context.Context) error {
	const query = `
		UPDATE loja_configuracoes
		SET
			draft_config_json = NULL,
			draft_updated_at = NULL,
			published_at = NOW(),
			atualizado_em = NOW()
		WHERE id = 1
	`

	_, err := r.pool.Exec(ctx, query)
	return err
}

func (r *Repository) UpsertSettings(ctx context.Context, input entity.SettingsRecord) error {
	const query = `
		INSERT INTO loja_configuracoes (
			id,
			nome_loja,
			subtitulo_loja,
			descricao_loja,
			tags_loja,
			loja_em_construcao,
			titulo_manutencao,
			mensagem_manutencao,
			cor_principal,
			cor_primaria,
			cor_secundaria,
			cor_fundo,
			cor_cabecalho,
			cor_menu,
			cor_rodape,
			cor_texto_cabecalho,
			cor_texto_menu,
			cor_texto_rodape,
			cor_texto_principal,
			cor_destaque,
			cor_desconto,
			cor_selo_lancamento,
			cor_selo_destaque,
			cor_selo_promocao,
			cor_botao,
			cor_texto_botao,
			estilo_botao,
			fonte_familia,
			fonte_secundaria_familia,
			fonte_menu_familia,
			fonte_botao_familia,
			fonte_destaque_familia,
			fonte_import_url,
			fonte_embed_code,
			estilo_arredondamento,
			modo_largura_layout,
			modo_fundo_menu,
			modo_largura_banner,
			efeito_banner_principal,
			tempo_rotacao_banner_segundos,
			modo_exibicao_marca,
			produtos_por_linha_desktop,
			produtos_por_linha_mobile,
			hero_titulo,
			hero_subtitulo,
			menu_links_json,
			section_order_json,
			section_visibility_json,
			custom_block_eyebrow,
			custom_block_title,
			custom_block_description,
			custom_block_html,
			custom_block_css,
			custom_block_js,
			feature_highlights_json,
			institutional_section_json,
			product_listing_config_json,
			footer_links_json,
			footer_contact_title,
			footer_contact_text,
			launches_section_json,
			featured_section_json,
			promotions_section_json,
			logo_base64, logo_mime, logo_largura, logo_altura, logo_tamanho_bytes, logo_hash_sha256,
			favicon_base64, favicon_mime, favicon_largura, favicon_altura, favicon_tamanho_bytes, favicon_hash_sha256,
			banner_desktop_base64, banner_desktop_mime, banner_desktop_largura, banner_desktop_altura, banner_desktop_tamanho_bytes, banner_desktop_hash_sha256,
			banner_mobile_base64, banner_mobile_mime, banner_mobile_largura, banner_mobile_altura, banner_mobile_tamanho_bytes, banner_mobile_hash_sha256,
			integracoes_json,
			atualizado_em
		) VALUES (
			1,
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,
			$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,
			$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,
			$42,$43,$44,$45,$46,$47,$48,$49,$50,$51,
			$52,$53,$54,$55,$56,$57,$58,$59,$60,$61,
			$62,$63,$64,$65,$66,$67,$68,$69,$70,$71,
			$72,$73,$74,$75,$76,$77,$78,$79,$80,$81,
			$82,$83,$84,$85,$86,$87,$88,
			NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			nome_loja = EXCLUDED.nome_loja,
			subtitulo_loja = EXCLUDED.subtitulo_loja,
			descricao_loja = EXCLUDED.descricao_loja,
			tags_loja = EXCLUDED.tags_loja,
			loja_em_construcao = EXCLUDED.loja_em_construcao,
			titulo_manutencao = EXCLUDED.titulo_manutencao,
			mensagem_manutencao = EXCLUDED.mensagem_manutencao,
			cor_principal = EXCLUDED.cor_principal,
			cor_primaria = EXCLUDED.cor_primaria,
			cor_secundaria = EXCLUDED.cor_secundaria,
			cor_fundo = EXCLUDED.cor_fundo,
			cor_cabecalho = EXCLUDED.cor_cabecalho,
			cor_menu = EXCLUDED.cor_menu,
			cor_rodape = EXCLUDED.cor_rodape,
			cor_texto_cabecalho = EXCLUDED.cor_texto_cabecalho,
			cor_texto_menu = EXCLUDED.cor_texto_menu,
			cor_texto_rodape = EXCLUDED.cor_texto_rodape,
			cor_texto_principal = EXCLUDED.cor_texto_principal,
			cor_destaque = EXCLUDED.cor_destaque,
			cor_desconto = EXCLUDED.cor_desconto,
			cor_selo_lancamento = EXCLUDED.cor_selo_lancamento,
			cor_selo_destaque = EXCLUDED.cor_selo_destaque,
			cor_selo_promocao = EXCLUDED.cor_selo_promocao,
			cor_botao = EXCLUDED.cor_botao,
			cor_texto_botao = EXCLUDED.cor_texto_botao,
			estilo_botao = EXCLUDED.estilo_botao,
			fonte_familia = EXCLUDED.fonte_familia,
			fonte_secundaria_familia = EXCLUDED.fonte_secundaria_familia,
			fonte_menu_familia = EXCLUDED.fonte_menu_familia,
			fonte_botao_familia = EXCLUDED.fonte_botao_familia,
			fonte_destaque_familia = EXCLUDED.fonte_destaque_familia,
			fonte_import_url = EXCLUDED.fonte_import_url,
			fonte_embed_code = EXCLUDED.fonte_embed_code,
			estilo_arredondamento = EXCLUDED.estilo_arredondamento,
			modo_largura_layout = EXCLUDED.modo_largura_layout,
			modo_fundo_menu = EXCLUDED.modo_fundo_menu,
			modo_largura_banner = EXCLUDED.modo_largura_banner,
			efeito_banner_principal = EXCLUDED.efeito_banner_principal,
			tempo_rotacao_banner_segundos = EXCLUDED.tempo_rotacao_banner_segundos,
			modo_exibicao_marca = EXCLUDED.modo_exibicao_marca,
			produtos_por_linha_desktop = EXCLUDED.produtos_por_linha_desktop,
			produtos_por_linha_mobile = EXCLUDED.produtos_por_linha_mobile,
			hero_titulo = EXCLUDED.hero_titulo,
			hero_subtitulo = EXCLUDED.hero_subtitulo,
			menu_links_json = EXCLUDED.menu_links_json,
			section_order_json = EXCLUDED.section_order_json,
			section_visibility_json = EXCLUDED.section_visibility_json,
			custom_block_eyebrow = EXCLUDED.custom_block_eyebrow,
			custom_block_title = EXCLUDED.custom_block_title,
			custom_block_description = EXCLUDED.custom_block_description,
			custom_block_html = EXCLUDED.custom_block_html,
			custom_block_css = EXCLUDED.custom_block_css,
			custom_block_js = EXCLUDED.custom_block_js,
			feature_highlights_json = EXCLUDED.feature_highlights_json,
			institutional_section_json = EXCLUDED.institutional_section_json,
			product_listing_config_json = EXCLUDED.product_listing_config_json,
			footer_links_json = EXCLUDED.footer_links_json,
			footer_contact_title = EXCLUDED.footer_contact_title,
			footer_contact_text = EXCLUDED.footer_contact_text,
			launches_section_json = EXCLUDED.launches_section_json,
			featured_section_json = EXCLUDED.featured_section_json,
			promotions_section_json = EXCLUDED.promotions_section_json,
			logo_base64 = EXCLUDED.logo_base64,
			logo_mime = EXCLUDED.logo_mime,
			logo_largura = EXCLUDED.logo_largura,
			logo_altura = EXCLUDED.logo_altura,
			logo_tamanho_bytes = EXCLUDED.logo_tamanho_bytes,
			logo_hash_sha256 = EXCLUDED.logo_hash_sha256,
			favicon_base64 = EXCLUDED.favicon_base64,
			favicon_mime = EXCLUDED.favicon_mime,
			favicon_largura = EXCLUDED.favicon_largura,
			favicon_altura = EXCLUDED.favicon_altura,
			favicon_tamanho_bytes = EXCLUDED.favicon_tamanho_bytes,
			favicon_hash_sha256 = EXCLUDED.favicon_hash_sha256,
			banner_desktop_base64 = EXCLUDED.banner_desktop_base64,
			banner_desktop_mime = EXCLUDED.banner_desktop_mime,
			banner_desktop_largura = EXCLUDED.banner_desktop_largura,
			banner_desktop_altura = EXCLUDED.banner_desktop_altura,
			banner_desktop_tamanho_bytes = EXCLUDED.banner_desktop_tamanho_bytes,
			banner_desktop_hash_sha256 = EXCLUDED.banner_desktop_hash_sha256,
			banner_mobile_base64 = EXCLUDED.banner_mobile_base64,
			banner_mobile_mime = EXCLUDED.banner_mobile_mime,
			banner_mobile_largura = EXCLUDED.banner_mobile_largura,
			banner_mobile_altura = EXCLUDED.banner_mobile_altura,
			banner_mobile_tamanho_bytes = EXCLUDED.banner_mobile_tamanho_bytes,
			banner_mobile_hash_sha256 = EXCLUDED.banner_mobile_hash_sha256,
			integracoes_json = EXCLUDED.integracoes_json,
			atualizado_em = NOW()
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		strings.TrimSpace(input.StoreName),
		nullableString(input.StoreSubtitle),
		nullableString(input.StoreDescription),
		nullableString(input.StoreTags),
		input.StoreInMaintenance,
		nullableString(input.MaintenanceTitle),
		nullableString(input.MaintenanceMessage),
		strings.TrimSpace(input.PrimaryColor),
		strings.TrimSpace(input.AccentColor),
		strings.TrimSpace(input.SecondaryColor),
		strings.TrimSpace(input.BackgroundColor),
		strings.TrimSpace(input.HeaderColor),
		strings.TrimSpace(input.MenuColor),
		strings.TrimSpace(input.FooterColor),
		strings.TrimSpace(input.HeaderTextColor),
		strings.TrimSpace(input.MenuTextColor),
		strings.TrimSpace(input.FooterTextColor),
		strings.TrimSpace(input.PrimaryTextColor),
		strings.TrimSpace(input.HighlightColor),
		strings.TrimSpace(input.DiscountColor),
		strings.TrimSpace(input.LaunchBadgeColor),
		strings.TrimSpace(input.FeaturedBadgeColor),
		strings.TrimSpace(input.PromotionBadgeColor),
		strings.TrimSpace(input.ButtonColor),
		strings.TrimSpace(input.ButtonTextColor),
		strings.TrimSpace(input.ButtonStyle),
		nullableString(input.FontFamily),
		nullableString(input.FontSecondaryFamily),
		nullableString(input.FontMenuFamily),
		nullableString(input.FontButtonFamily),
		nullableString(input.FontHighlightFamily),
		nullableString(input.FontImportURL),
		nullableString(input.FontEmbedCode),
		strings.TrimSpace(input.CornerStyle),
		strings.TrimSpace(input.ContentWidthMode),
		strings.TrimSpace(input.MenuBackgroundMode),
		strings.TrimSpace(input.BannerWidthMode),
		strings.TrimSpace(input.BannerEffectMode),
		input.BannerRotationSeconds,
		strings.TrimSpace(input.BrandDisplayMode),
		input.ProductsPerRowDesktop,
		input.ProductsPerRowMobile,
		nullableString(input.HeroTitle),
		nullableString(input.HeroSubtitle),
		marshalNavigationLinks(input.MenuLinks),
		marshalSectionOrder(input.SectionOrder),
		marshalSectionVisibility(input.VisibleSections),
		nullableString(input.CustomBlockEyebrow),
		nullableString(input.CustomBlockTitle),
		nullableString(input.CustomBlockDescription),
		nullableString(input.CustomBlockHTML),
		nullableString(input.CustomBlockCSS),
		nullableString(input.CustomBlockJS),
		marshalFeatureHighlights(input.FeatureHighlights),
		marshalInstitutionalSectionConfig(input.InstitutionalSection),
		marshalProductListingConfig(input.ProductListingConfig),
		marshalNavigationLinks(input.FooterLinks),
		nullableString(input.FooterContactTitle),
		nullableString(input.FooterContactText),
		marshalProductSectionConfig(input.LaunchesSection),
		marshalProductSectionConfig(input.FeaturedSection),
		marshalProductSectionConfig(input.PromotionsSection),
		readBase64(input.Logo), readMime(input.Logo), readWidth(input.Logo), readHeight(input.Logo), readSize(input.Logo), readHash(input.Logo),
		readBase64(input.Favicon), readMime(input.Favicon), readWidth(input.Favicon), readHeight(input.Favicon), readSize(input.Favicon), readHash(input.Favicon),
		readBase64(input.BannerDesktop), readMime(input.BannerDesktop), readWidth(input.BannerDesktop), readHeight(input.BannerDesktop), readSize(input.BannerDesktop), readHash(input.BannerDesktop),
		readBase64(input.BannerMobile), readMime(input.BannerMobile), readWidth(input.BannerMobile), readHeight(input.BannerMobile), readSize(input.BannerMobile), readHash(input.BannerMobile),
		marshalIntegrationSettings(input.Integrations),
	)
	return err
}

func (r *Repository) ListBanners(ctx context.Context) ([]entity.BannerPayload, error) {
	const query = `
		SELECT
			id::text,
			coalesce(titulo, ''),
			coalesce(subtitulo, ''),
			coalesce(link, ''),
			ordem,
			ativo,
			coalesce(show_content, true),
			coalesce(link_mode, CASE WHEN coalesce(use_button, false) = true THEN 'button' WHEN trim(coalesce(link, '')) <> '' THEN 'banner' ELSE 'none' END),
			coalesce(content_position, 'center'),
			coalesce(content_position_x, CASE WHEN content_position = 'left' THEN 'left' WHEN content_position = 'right' THEN 'right' ELSE 'center' END),
			coalesce(content_position_y, CASE WHEN content_position = 'top' THEN 'top' WHEN content_position = 'bottom' THEN 'bottom' ELSE 'center' END),
			coalesce(use_button, false),
			coalesce(button_label, ''),
			coalesce(desktop_base64, ''),
			coalesce(desktop_mime, ''),
			coalesce(desktop_largura, 0),
			coalesce(desktop_altura, 0),
			coalesce(desktop_tamanho_bytes, 0),
			coalesce(desktop_hash_sha256, ''),
			coalesce(mobile_base64, ''),
			coalesce(mobile_mime, ''),
			coalesce(mobile_largura, 0),
			coalesce(mobile_altura, 0),
			coalesce(mobile_tamanho_bytes, 0),
			coalesce(mobile_hash_sha256, ''),
			criado_em,
			atualizado_em
		FROM loja_banners
		ORDER BY ativo DESC, ordem ASC, criado_em DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.BannerPayload, 0)
	for rows.Next() {
		var item entity.BannerPayload
		var desktopBase64, desktopMime, desktopHash string
		var desktopWidth, desktopHeight, desktopSize int
		var mobileBase64, mobileMime, mobileHash string
		var mobileWidth, mobileHeight, mobileSize int
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Subtitle,
			&item.Link,
			&item.Order,
			&item.Active,
			&item.ShowContent,
			&item.LinkMode,
			&item.ContentPosition,
			&item.ContentPositionX,
			&item.ContentPositionY,
			&item.UseButton,
			&item.ButtonLabel,
			&desktopBase64,
			&desktopMime,
			&desktopWidth,
			&desktopHeight,
			&desktopSize,
			&desktopHash,
			&mobileBase64,
			&mobileMime,
			&mobileWidth,
			&mobileHeight,
			&mobileSize,
			&mobileHash,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.Desktop = buildImageAsset(desktopBase64, desktopMime, desktopWidth, desktopHeight, desktopSize, desktopHash)
		item.Mobile = buildImageAsset(mobileBase64, mobileMime, mobileWidth, mobileHeight, mobileSize, mobileHash)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) ReplaceBanners(ctx context.Context, items []entity.BannerPayload) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM loja_banners`); err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	const query = `
		INSERT INTO loja_banners (
			titulo,
			subtitulo,
			link,
			ordem,
			ativo,
			show_content,
			link_mode,
			content_position,
			content_position_x,
			content_position_y,
			use_button,
			button_label,
			desktop_base64, desktop_mime, desktop_largura, desktop_altura, desktop_tamanho_bytes, desktop_hash_sha256,
			mobile_base64, mobile_mime, mobile_largura, mobile_altura, mobile_tamanho_bytes, mobile_hash_sha256
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,
			$13,$14,$15,$16,$17,$18,
			$19,$20,$21,$22,$23,$24
		)
	`

	for _, item := range items {
		if _, err := r.pool.Exec(
			ctx,
			query,
			nullableString(item.Title),
			nullableString(item.Subtitle),
			nullableString(item.Link),
			item.Order,
			item.Active,
			item.ShowContent,
			item.LinkMode,
			item.ContentPosition,
			item.ContentPositionX,
			item.ContentPositionY,
			item.UseButton,
			nullableString(item.ButtonLabel),
			readBase64(item.Desktop), readMime(item.Desktop), readWidth(item.Desktop), readHeight(item.Desktop), readSize(item.Desktop), readHash(item.Desktop),
			readBase64(item.Mobile), readMime(item.Mobile), readWidth(item.Mobile), readHeight(item.Mobile), readSize(item.Mobile), readHash(item.Mobile),
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) ListCategories(ctx context.Context) ([]entity.AdminStoreCategoryListItem, error) {
	const query = `
		SELECT
			c.id,
			c.nome,
			c.slug,
			coalesce(c.descricao, ''),
			c.ordem,
			c.ativa,
			count(DISTINCT lpc.produto_id)::int AS produtos_count,
			c.criado_em,
			c.atualizado_em
		FROM loja_categorias c
		LEFT JOIN loja_produto_categorias lpc ON lpc.categoria_id = c.id
		GROUP BY c.id, c.nome, c.slug, c.descricao, c.ordem, c.ativa, c.criado_em, c.atualizado_em
		ORDER BY c.ordem ASC, c.nome ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.AdminStoreCategoryListItem, 0)
	for rows.Next() {
		var item entity.AdminStoreCategoryListItem
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Nome,
			&item.Slug,
			&item.Descricao,
			&item.Ordem,
			&item.Ativa,
			&item.ProdutosCount,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.CriadoEm = createdAt.Format(time.RFC3339)
		item.AtualizadoEm = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) CreateCategory(ctx context.Context, input entity.StoreCategoryRecord) (string, error) {
	const query = `
		INSERT INTO loja_categorias (
			nome,
			slug,
			descricao,
			ordem,
			ativa
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err := r.pool.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.Nome),
		strings.TrimSpace(input.Slug),
		nullableString(input.Descricao),
		input.Ordem,
		input.Ativa,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdateCategory(ctx context.Context, input entity.StoreCategoryRecord) error {
	commandTag, err := r.pool.Exec(
		ctx,
		`
			UPDATE loja_categorias
			SET
				nome = $2,
				slug = $3,
				descricao = $4,
				ordem = $5,
				ativa = $6,
				atualizado_em = NOW()
			WHERE id = $1
		`,
		input.ID,
		strings.TrimSpace(input.Nome),
		strings.TrimSpace(input.Slug),
		nullableString(input.Descricao),
		input.Ordem,
		input.Ativa,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `SELECT id FROM loja_categorias WHERE id = $1`, id).Scan(new(string)); err != nil {
		return err
	}

	if _, err := tx.Exec(
		ctx,
		`DELETE FROM loja_produto_categorias WHERE categoria_id = $1`,
		id,
	); err != nil {
		return err
	}

	commandTag, err := tx.Exec(ctx, `DELETE FROM loja_categorias WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
}

func (r *Repository) ListAdminProducts(ctx context.Context) ([]entity.AdminProductListItem, error) {
	const query = `
		SELECT
			lp.id,
			coalesce(lp.livro_id::text, ''),
			lp.is_book,
			coalesce(lp.authors_json, '[]'),
			coalesce(l.titulo, ''),
			l.subtitulo,
			coalesce(nullif(trim(lp.autor_nome), ''), a.nome_completo),
			lp.marca,
			lp.editora,
			lp.subtitulo,
			lp.sinopse,
			lp.isbn,
			lp.codigo_barra,
			lp.edicao,
			lp.idioma,
			lp.numero_paginas,
			lp.genero,
			to_char(lp.data_publicacao, 'YYYY-MM-DD'),
			lp.tipo_capa,
			lp.peso_gramas,
			lp.largura_cm,
			lp.altura_cm,
			lp.profundidade_cm,
			lp.slug,
			lp.nome_exibicao,
			coalesce(lp.descricao_curta, ''),
			coalesce(string_agg(distinct c.nome, ', ' ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), ''),
			coalesce(array_agg(distinct c.nome ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), '{}'),
			lp.preco_venda,
			lp.em_promocao,
			coalesce(lp.preco_promocional, 0),
			lp.destaque,
			lp.lancamento,
			lp.ativo,
			lp.ordem,
			coalesce(l.capa_base64, ''),
			coalesce(l.capa_mime, ''),
			coalesce(l.capa_largura, 0),
			coalesce(l.capa_altura, 0),
			coalesce(l.capa_tamanho_bytes, 0),
			coalesce(l.capa_hash_sha256, ''),
			coalesce(lp.fotos_json, '[]'),
			lp.criado_em,
			lp.atualizado_em
		FROM loja_produtos lp
		LEFT JOIN livros l ON l.id = lp.livro_id
		LEFT JOIN autores a ON a.id = l.autor_id
		LEFT JOIN loja_produto_categorias lpc ON lpc.produto_id = lp.id
		LEFT JOIN loja_categorias c ON c.id = lpc.categoria_id
		GROUP BY
			lp.id, lp.livro_id, lp.is_book, lp.authors_json, lp.autor_nome, lp.marca, lp.editora, lp.subtitulo, lp.sinopse, lp.isbn, lp.codigo_barra, lp.edicao,
			lp.idioma, lp.numero_paginas, lp.genero, lp.data_publicacao, lp.tipo_capa, lp.peso_gramas, lp.largura_cm, lp.altura_cm, lp.profundidade_cm,
			l.titulo, l.subtitulo, a.nome_completo, lp.slug, lp.nome_exibicao, lp.descricao_curta, lp.preco_venda, lp.em_promocao, lp.preco_promocional, lp.destaque,
			lp.lancamento, lp.ativo, lp.ordem, l.capa_base64, l.capa_mime, l.capa_largura, l.capa_altura,
			l.capa_tamanho_bytes, l.capa_hash_sha256, lp.fotos_json, lp.criado_em, lp.atualizado_em
		ORDER BY lp.destaque DESC, lp.ordem ASC, lp.criado_em DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.AdminProductListItem, 0)
	for rows.Next() {
		item, err := scanAdminProduct(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetAdminProductByID(ctx context.Context, id string) (entity.AdminProductDetail, error) {
	const query = `
		SELECT
			lp.id,
			coalesce(lp.livro_id::text, ''),
			lp.is_book,
			coalesce(lp.authors_json, '[]'),
			coalesce(l.titulo, ''),
			l.subtitulo,
			coalesce(nullif(trim(lp.autor_nome), ''), a.nome_completo),
			lp.marca,
			lp.editora,
			lp.subtitulo,
			lp.sinopse,
			lp.isbn,
			lp.codigo_barra,
			lp.edicao,
			lp.idioma,
			lp.numero_paginas,
			lp.genero,
			to_char(lp.data_publicacao, 'YYYY-MM-DD'),
			lp.tipo_capa,
			lp.peso_gramas,
			lp.largura_cm,
			lp.altura_cm,
			lp.profundidade_cm,
			lp.slug,
			lp.nome_exibicao,
			coalesce(lp.descricao_curta, ''),
			coalesce(string_agg(distinct c.nome, ', ' ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), ''),
			coalesce(array_agg(distinct c.nome ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), '{}'),
			lp.preco_venda,
			lp.em_promocao,
			coalesce(lp.preco_promocional, 0),
			lp.destaque,
			lp.lancamento,
			lp.ativo,
			lp.ordem,
			coalesce(l.capa_base64, ''),
			coalesce(l.capa_mime, ''),
			coalesce(l.capa_largura, 0),
			coalesce(l.capa_altura, 0),
			coalesce(l.capa_tamanho_bytes, 0),
			coalesce(l.capa_hash_sha256, ''),
			coalesce(lp.fotos_json, '[]'),
			lp.criado_em,
			lp.atualizado_em
		FROM loja_produtos lp
		LEFT JOIN livros l ON l.id = lp.livro_id
		LEFT JOIN autores a ON a.id = l.autor_id
		LEFT JOIN loja_produto_categorias lpc ON lpc.produto_id = lp.id
		LEFT JOIN loja_categorias c ON c.id = lpc.categoria_id
		WHERE lp.id = $1
		GROUP BY
			lp.id, lp.livro_id, lp.is_book, lp.authors_json, lp.autor_nome, lp.marca, lp.editora, lp.subtitulo, lp.sinopse, lp.isbn, lp.codigo_barra, lp.edicao,
			lp.idioma, lp.numero_paginas, lp.genero, lp.data_publicacao, lp.tipo_capa, lp.peso_gramas, lp.largura_cm, lp.altura_cm, lp.profundidade_cm,
			l.titulo, l.subtitulo, a.nome_completo, lp.slug, lp.nome_exibicao, lp.descricao_curta, lp.preco_venda, lp.em_promocao, lp.preco_promocional, lp.destaque,
			lp.lancamento, lp.ativo, lp.ordem, l.capa_base64, l.capa_mime, l.capa_largura, l.capa_altura,
			l.capa_tamanho_bytes, l.capa_hash_sha256, lp.fotos_json, lp.criado_em, lp.atualizado_em
	`

	row := r.pool.QueryRow(ctx, query, id)
	return scanAdminProduct(row)
}

func (r *Repository) CreateProduct(ctx context.Context, input entity.ProductRecord) (string, error) {
	const query = `
		INSERT INTO loja_produtos (
			livro_id,
			is_book,
			authors_json,
			autor_nome,
			marca,
			editora,
			subtitulo,
			sinopse,
			isbn,
			codigo_barra,
			edicao,
			idioma,
			numero_paginas,
			genero,
			data_publicacao,
			tipo_capa,
			peso_gramas,
			largura_cm,
			altura_cm,
			profundidade_cm,
			slug,
			nome_exibicao,
			descricao_curta,
			categoria,
			preco_venda,
			em_promocao,
			preco_promocional,
			destaque,
			lancamento,
			ativo,
			ordem,
			fotos_json
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32)
		RETURNING id
	`

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var id string
	err = tx.QueryRow(
		ctx,
		query,
		nullableUUID(input.LivroID),
		input.IsBook,
		marshalProductAuthors(input.Authors),
		nullableString(strings.TrimSpace(input.AutorNome)),
		nullableString(input.Marca),
		nullableString(input.Editora),
		nullableString(input.Subtitulo),
		nullableString(input.Sinopse),
		nullableString(input.ISBN),
		nullableString(input.CodigoBarra),
		nullableString(input.Edicao),
		nullableString(input.Idioma),
		nullableInt(input.NumeroPaginas),
		nullableString(input.Genero),
		nullableDate(input.DataPublicacao),
		nullableString(input.TipoCapa),
		nullableInt(input.PesoGramas),
		nullableFloat(input.LarguraCm),
		nullableFloat(input.AlturaCm),
		nullableFloat(input.ProfundidadeCm),
		strings.TrimSpace(input.Slug),
		strings.TrimSpace(input.NomeExibicao),
		nullableString(input.DescricaoCurta),
		nullableString(strings.Join(input.Categorias, ", ")),
		input.PrecoVenda,
		input.EmPromocao,
		nullableMoney(input.PrecoPromocional),
		input.Destaque,
		input.Lancamento,
		input.Ativo,
		input.Ordem,
		marshalProductPhotos(input.Fotos),
	).Scan(&id)
	if err != nil {
		return "", err
	}
	if err := syncProductCategories(ctx, tx, id, input.Categorias); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return id, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, input entity.ProductRecord) error {
	const query = `
		UPDATE loja_produtos
		SET
			livro_id = $2,
			is_book = $3,
			authors_json = $4,
			autor_nome = $5,
			marca = $6,
			editora = $7,
			subtitulo = $8,
			sinopse = $9,
			isbn = $10,
			codigo_barra = $11,
			edicao = $12,
			idioma = $13,
			numero_paginas = $14,
			genero = $15,
			data_publicacao = $16,
			tipo_capa = $17,
			peso_gramas = $18,
			largura_cm = $19,
			altura_cm = $20,
			profundidade_cm = $21,
			slug = $22,
			nome_exibicao = $23,
			descricao_curta = $24,
			categoria = $25,
			preco_venda = $26,
			em_promocao = $27,
			preco_promocional = $28,
			destaque = $29,
			lancamento = $30,
			ativo = $31,
			ordem = $32,
			fotos_json = $33,
			atualizado_em = NOW()
		WHERE id = $1
	`

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		query,
		input.ID,
		nullableUUID(input.LivroID),
		input.IsBook,
		marshalProductAuthors(input.Authors),
		nullableString(strings.TrimSpace(input.AutorNome)),
		nullableString(input.Marca),
		nullableString(input.Editora),
		nullableString(input.Subtitulo),
		nullableString(input.Sinopse),
		nullableString(input.ISBN),
		nullableString(input.CodigoBarra),
		nullableString(input.Edicao),
		nullableString(input.Idioma),
		nullableInt(input.NumeroPaginas),
		nullableString(input.Genero),
		nullableDate(input.DataPublicacao),
		nullableString(input.TipoCapa),
		nullableInt(input.PesoGramas),
		nullableFloat(input.LarguraCm),
		nullableFloat(input.AlturaCm),
		nullableFloat(input.ProfundidadeCm),
		strings.TrimSpace(input.Slug),
		strings.TrimSpace(input.NomeExibicao),
		nullableString(input.DescricaoCurta),
		nullableString(strings.Join(input.Categorias, ", ")),
		input.PrecoVenda,
		input.EmPromocao,
		nullableMoney(input.PrecoPromocional),
		input.Destaque,
		input.Lancamento,
		input.Ativo,
		input.Ordem,
		marshalProductPhotos(input.Fotos),
	)
	if err != nil {
		return err
	}
	if err := syncProductCategories(ctx, tx, input.ID, input.Categorias); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListPublicProducts(ctx context.Context) ([]entity.PublicProductListItem, error) {
	const query = `
		SELECT
			lp.id,
			lp.slug,
			lp.nome_exibicao,
			coalesce(lp.descricao_curta, ''),
			coalesce(string_agg(distinct c.nome, ', ' ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), ''),
			coalesce(array_agg(distinct c.nome ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), '{}'),
			lp.ordem,
			lp.preco_venda,
			lp.em_promocao,
			coalesce(lp.preco_promocional, 0),
			lp.destaque,
			lp.lancamento,
			lp.is_book,
			coalesce(lp.authors_json, '[]'),
			coalesce(nullif(trim(lp.autor_nome), ''), a.nome_completo, ''),
			coalesce(lp.marca, ''),
			coalesce(lp.editora, ''),
			coalesce(l.titulo, ''),
			l.subtitulo,
			coalesce(l.possui_formato_digital, false),
			coalesce(l.possui_formato_fisico, false),
			l.url_compra_digital,
			coalesce(l.capa_base64, ''),
			coalesce(l.capa_mime, ''),
			coalesce(l.capa_largura, 0),
			coalesce(l.capa_altura, 0),
			coalesce(l.capa_tamanho_bytes, 0),
			coalesce(l.capa_hash_sha256, ''),
			coalesce(lp.fotos_json, '[]')
		FROM loja_produtos lp
		LEFT JOIN livros l ON l.id = lp.livro_id
		LEFT JOIN autores a ON a.id = l.autor_id
		LEFT JOIN loja_produto_categorias lpc ON lpc.produto_id = lp.id
		LEFT JOIN loja_categorias c ON c.id = lpc.categoria_id
		WHERE lp.ativo = TRUE
		  AND (l.id IS NULL OR l.ativo = TRUE)
		GROUP BY
			lp.id, lp.slug, lp.nome_exibicao, lp.descricao_curta, lp.ordem, lp.preco_venda, lp.em_promocao,
			lp.preco_promocional, lp.destaque, lp.lancamento, lp.is_book, lp.authors_json, lp.autor_nome, lp.marca, lp.editora, a.nome_completo, l.titulo, l.subtitulo,
			l.possui_formato_digital, l.possui_formato_fisico, l.url_compra_digital, l.capa_base64,
			l.capa_mime, l.capa_largura, l.capa_altura, l.capa_tamanho_bytes, l.capa_hash_sha256, lp.fotos_json, lp.criado_em
		ORDER BY lp.destaque DESC, lp.ordem ASC, lp.criado_em DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.PublicProductListItem, 0)
	for rows.Next() {
		var item entity.PublicProductListItem
		var capaBase64, capaMime, capaHash, fotosJSON, authorsJSON string
		var capaWidth, capaHeight, capaSize int
		err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.NomeExibicao,
			&item.DescricaoCurta,
			&item.Categoria,
			&item.Categorias,
			&item.Ordem,
			&item.PrecoVenda,
			&item.EmPromocao,
			&item.PrecoPromocional,
			&item.Destaque,
			&item.Lancamento,
			&item.IsBook,
			&authorsJSON,
			&item.AutorNome,
			&item.Marca,
			&item.Editora,
			&item.LivroTitulo,
			&item.LivroSubtitulo,
			&item.PossuiDigital,
			&item.PossuiFisico,
			&item.URLCompraDigital,
			&capaBase64,
			&capaMime,
			&capaWidth,
			&capaHeight,
			&capaSize,
			&capaHash,
			&fotosJSON,
		)
		if err != nil {
			return nil, err
		}
		item.Authors = unmarshalProductAuthors(authorsJSON)
		item.Fotos = unmarshalProductPhotos(fotosJSON)
		item.Capa = preferredProductImage(item.Fotos, buildImageAsset(capaBase64, capaMime, capaWidth, capaHeight, capaSize, capaHash))
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetPublicProductBySlug(ctx context.Context, slug string) (entity.PublicProductDetail, error) {
	const query = `
		SELECT
			lp.id,
			lp.slug,
			lp.nome_exibicao,
			coalesce(lp.descricao_curta, ''),
			coalesce(string_agg(distinct c.nome, ', ' ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), ''),
			coalesce(array_agg(distinct c.nome ORDER BY c.nome) FILTER (WHERE c.nome IS NOT NULL), '{}'),
			lp.ordem,
			lp.preco_venda,
			lp.em_promocao,
			coalesce(lp.preco_promocional, 0),
			lp.destaque,
			lp.lancamento,
			lp.is_book,
			coalesce(lp.authors_json, '[]'),
			coalesce(nullif(trim(lp.autor_nome), ''), a.nome_completo, ''),
			coalesce(lp.marca, ''),
			coalesce(lp.editora, ''),
			coalesce(l.titulo, ''),
			l.subtitulo,
			l.sinopse,
			coalesce(l.possui_formato_digital, false),
			coalesce(l.possui_formato_fisico, false),
			l.url_compra_digital,
			coalesce(l.capa_base64, ''),
			coalesce(l.capa_mime, ''),
			coalesce(l.capa_largura, 0),
			coalesce(l.capa_altura, 0),
			coalesce(l.capa_tamanho_bytes, 0),
			coalesce(l.capa_hash_sha256, ''),
			coalesce(lp.fotos_json, '[]')
		FROM loja_produtos lp
		LEFT JOIN livros l ON l.id = lp.livro_id
		LEFT JOIN autores a ON a.id = l.autor_id
		LEFT JOIN loja_produto_categorias lpc ON lpc.produto_id = lp.id
		LEFT JOIN loja_categorias c ON c.id = lpc.categoria_id
		WHERE lp.slug = $1
		  AND lp.ativo = TRUE
		  AND (l.id IS NULL OR l.ativo = TRUE)
		GROUP BY
			lp.id, lp.slug, lp.nome_exibicao, lp.descricao_curta, lp.ordem, lp.preco_venda, lp.em_promocao,
			lp.preco_promocional, lp.destaque, lp.lancamento, lp.is_book, lp.authors_json, lp.autor_nome, lp.marca, lp.editora, a.nome_completo, l.titulo, l.subtitulo,
			l.sinopse, l.possui_formato_digital, l.possui_formato_fisico, l.url_compra_digital, l.capa_base64,
			l.capa_mime, l.capa_largura, l.capa_altura, l.capa_tamanho_bytes, l.capa_hash_sha256, lp.fotos_json
	`

	trimmedSlug := strings.TrimSpace(slug)
	var item entity.PublicProductDetail
	var capaBase64, capaMime, capaHash, fotosJSON, authorsJSON string
	var capaWidth, capaHeight, capaSize int
	err := r.pool.QueryRow(ctx, query, trimmedSlug).Scan(
		&item.ID,
		&item.Slug,
		&item.NomeExibicao,
		&item.DescricaoCurta,
		&item.Categoria,
		&item.Categorias,
		&item.Ordem,
		&item.PrecoVenda,
		&item.EmPromocao,
		&item.PrecoPromocional,
		&item.Destaque,
		&item.Lancamento,
		&item.IsBook,
		&authorsJSON,
		&item.AutorNome,
		&item.Marca,
		&item.Editora,
		&item.LivroTitulo,
		&item.LivroTitulo,
		&item.LivroTitulo,
		&item.LivroSubtitulo,
		&item.Sinopse,
		&item.PossuiDigital,
		&item.PossuiFisico,
		&item.URLCompraDigital,
		&capaBase64,
		&capaMime,
		&capaWidth,
		&capaHeight,
		&capaSize,
		&capaHash,
		&fotosJSON,
	)
	if err != nil {
		return entity.PublicProductDetail{}, err
	}

	item.Authors = unmarshalProductAuthors(authorsJSON)
	item.Fotos = unmarshalProductPhotos(fotosJSON)
	item.Capa = preferredProductImage(item.Fotos, buildImageAsset(capaBase64, capaMime, capaWidth, capaHeight, capaSize, capaHash))
	return item, nil
}

func scanAdminProduct(row interface {
	Scan(dest ...any) error
}) (entity.AdminProductDetail, error) {
	var item entity.AdminProductDetail
	var categorias []string
	var capaBase64, capaMime, capaHash, fotosJSON, authorsJSON string
	var capaWidth, capaHeight, capaSize int
	var criadoEm, atualizadoEm time.Time
	err := row.Scan(
		&item.ID,
		&item.LivroID,
		&item.IsBook,
		&authorsJSON,
		&item.LivroTitulo,
		&item.LivroSubtitulo,
		&item.AutorNome,
		&item.Marca,
		&item.Editora,
		&item.Subtitulo,
		&item.Sinopse,
		&item.ISBN,
		&item.CodigoBarra,
		&item.Edicao,
		&item.Idioma,
		&item.NumeroPaginas,
		&item.Genero,
		&item.DataPublicacao,
		&item.TipoCapa,
		&item.PesoGramas,
		&item.LarguraCm,
		&item.AlturaCm,
		&item.ProfundidadeCm,
		&item.Slug,
		&item.NomeExibicao,
		&item.DescricaoCurta,
		&item.Categoria,
		&categorias,
		&item.PrecoVenda,
		&item.EmPromocao,
		&item.PrecoPromocional,
		&item.Destaque,
		&item.Lancamento,
		&item.Ativo,
		&item.Ordem,
		&capaBase64,
		&capaMime,
		&capaWidth,
		&capaHeight,
		&capaSize,
		&capaHash,
		&fotosJSON,
		&criadoEm,
		&atualizadoEm,
	)
	if err != nil {
		return entity.AdminProductDetail{}, err
	}

	item.Authors = unmarshalProductAuthors(authorsJSON)
	item.Fotos = unmarshalProductPhotos(fotosJSON)
	item.Capa = preferredProductImage(item.Fotos, buildImageAsset(capaBase64, capaMime, capaWidth, capaHeight, capaSize, capaHash))
	item.Categorias = categorias
	item.CriadoEm = formatRepositoryOptionalTime(criadoEm)
	item.AtualizadoEm = formatRepositoryOptionalTime(atualizadoEm)
	return item, nil
}

func formatRepositoryOptionalTime(value time.Time) string {
	if value.IsZero() || value.Unix() <= 0 {
		return ""
	}
	return value.Format(time.RFC3339)
}

func syncProductCategories(ctx context.Context, tx pgx.Tx, productID string, categories []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM loja_produto_categorias WHERE produto_id = $1`, productID); err != nil {
		return err
	}

	if len(categories) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(categories))
	for _, item := range categories {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, strings.ToLower(trimmed))
	}
	if len(normalized) == 0 {
		return nil
	}

	rows, err := tx.Query(
		ctx,
		`SELECT id FROM loja_categorias WHERE lower(nome) = ANY($1::text[])`,
		normalized,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	categoryIDs := make([]string, 0, len(normalized))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		categoryIDs = append(categoryIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, categoryID := range categoryIDs {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO loja_produto_categorias (produto_id, categoria_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			productID,
			categoryID,
		); err != nil {
			return err
		}
	}

	return nil
}

func buildImageAsset(base64 string, mime string, width int, height int, size int, hash string) *entity.ImageAsset {
	if strings.TrimSpace(base64) == "" {
		return nil
	}
	return &entity.ImageAsset{
		Base64:       base64,
		Mime:         strings.TrimSpace(mime),
		Largura:      width,
		Altura:       height,
		TamanhoBytes: size,
		HashSHA256:   strings.TrimSpace(hash),
	}
}

func nullableString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableUUID(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableMoney(value float64) any {
	if value <= 0 {
		return nil
	}
	return value
}

func nullableInt(value *int) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func nullableFloat(value *float64) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func nullableDate(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func readBase64(asset *entity.ImageAsset) any {
	if asset == nil || strings.TrimSpace(asset.Base64) == "" {
		return nil
	}
	return asset.Base64
}

func readMime(asset *entity.ImageAsset) any {
	if asset == nil || strings.TrimSpace(asset.Base64) == "" {
		return nil
	}
	return strings.TrimSpace(asset.Mime)
}

func readWidth(asset *entity.ImageAsset) any {
	if asset == nil || strings.TrimSpace(asset.Base64) == "" {
		return nil
	}
	return asset.Largura
}

func readHeight(asset *entity.ImageAsset) any {
	if asset == nil || strings.TrimSpace(asset.Base64) == "" {
		return nil
	}
	return asset.Altura
}

func readSize(asset *entity.ImageAsset) any {
	if asset == nil || strings.TrimSpace(asset.Base64) == "" {
		return nil
	}
	return asset.TamanhoBytes
}

func readHash(asset *entity.ImageAsset) any {
	if asset == nil || strings.TrimSpace(asset.Base64) == "" {
		return nil
	}
	return strings.TrimSpace(asset.HashSHA256)
}

func marshalProductPhotos(items []entity.ProductPhoto) any {
	normalized := normalizeProductPhotos(items)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func marshalProductAuthors(items []entity.ProductAuthor) any {
	normalized := normalizeProductAuthors(items)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func marshalNavigationLinks(items []entity.NavigationLink) any {
	normalized := make([]entity.NavigationLink, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		url := strings.TrimSpace(item.URL)
		if label == "" || url == "" {
			continue
		}
		normalized = append(normalized, entity.NavigationLink{
			Label:   label,
			URL:     url,
			Visible: item.Visible,
			Kind:    strings.TrimSpace(item.Kind),
		})
	}
	if len(normalized) == 0 {
		return "[]"
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func marshalFeatureHighlights(items []entity.FeatureHighlight) any {
	normalized := make([]entity.FeatureHighlight, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		text := strings.TrimSpace(item.Text)
		icon := strings.TrimSpace(item.Icon)
		if title == "" || text == "" {
			continue
		}
		normalized = append(normalized, entity.FeatureHighlight{
			Title:      title,
			Text:       text,
			Icon:       icon,
			TextAlign:  strings.TrimSpace(item.TextAlign),
			IconSize:   strings.TrimSpace(item.IconSize),
			FontFamily: strings.TrimSpace(item.FontFamily),
		})
	}
	if len(normalized) == 0 {
		return "[]"
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func marshalProductSectionConfig(item entity.ProductSectionConfig) any {
	payload, err := json.Marshal(item)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func marshalInstitutionalSectionConfig(item entity.InstitutionalSectionConfig) any {
	payload, err := json.Marshal(item)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func marshalProductListingConfig(item entity.ProductListingConfig) any {
	payload, err := json.Marshal(item)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func marshalSectionOrder(items []entity.StorefrontSection) any {
	normalized := normalizeSectionOrder(items)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func marshalSectionVisibility(items []entity.StorefrontSection) any {
	normalized := normalizeSectionVisibility(items)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func marshalIntegrationSettings(item entity.IntegrationSettings) any {
	payload, err := json.Marshal(item)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func unmarshalNavigationLinks(value string) []entity.NavigationLink {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []entity.NavigationLink{}
	}
	var items []entity.NavigationLink
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return []entity.NavigationLink{}
	}
	return items
}

func unmarshalFeatureHighlights(value string) []entity.FeatureHighlight {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []entity.FeatureHighlight{}
	}
	var items []entity.FeatureHighlight
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return []entity.FeatureHighlight{}
	}
	return items
}

func unmarshalProductSectionConfig(value string) entity.ProductSectionConfig {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return entity.ProductSectionConfig{}
	}
	var item entity.ProductSectionConfig
	if err := json.Unmarshal([]byte(trimmed), &item); err != nil {
		return entity.ProductSectionConfig{}
	}
	return item
}

func unmarshalInstitutionalSectionConfig(value string) entity.InstitutionalSectionConfig {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return entity.InstitutionalSectionConfig{}
	}
	var item entity.InstitutionalSectionConfig
	if err := json.Unmarshal([]byte(trimmed), &item); err != nil {
		return entity.InstitutionalSectionConfig{}
	}
	return item
}

func unmarshalProductListingConfig(value string) entity.ProductListingConfig {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return entity.ProductListingConfig{}
	}
	var item entity.ProductListingConfig
	if err := json.Unmarshal([]byte(trimmed), &item); err != nil {
		return entity.ProductListingConfig{}
	}
	return item
}

func unmarshalSectionOrder(value string) []entity.StorefrontSection {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultSectionOrder()
	}
	var items []entity.StorefrontSection
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return defaultSectionOrder()
	}
	return normalizeSectionOrder(items)
}

func unmarshalSectionVisibility(value string) []entity.StorefrontSection {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultSectionVisibility()
	}
	var items []entity.StorefrontSection
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return defaultSectionVisibility()
	}
	return normalizeSectionVisibility(items)
}

func unmarshalIntegrationSettings(value string) entity.IntegrationSettings {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return entity.IntegrationSettings{}
	}
	var item entity.IntegrationSettings
	if err := json.Unmarshal([]byte(trimmed), &item); err != nil {
		return entity.IntegrationSettings{}
	}
	return item
}

func unmarshalProductPhotos(value string) []entity.ProductPhoto {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []entity.ProductPhoto{}
	}
	var items []entity.ProductPhoto
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return []entity.ProductPhoto{}
	}
	return normalizeProductPhotos(items)
}

func unmarshalProductAuthors(value string) []entity.ProductAuthor {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []entity.ProductAuthor{}
	}
	var items []entity.ProductAuthor
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return []entity.ProductAuthor{}
	}
	return normalizeProductAuthors(items)
}

func normalizeProductAuthors(items []entity.ProductAuthor) []entity.ProductAuthor {
	if len(items) == 0 {
		return []entity.ProductAuthor{}
	}

	normalized := make([]entity.ProductAuthor, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		nome := strings.TrimSpace(item.Nome)
		if nome == "" {
			continue
		}
		key := strings.ToLower(id + "::" + nome)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, entity.ProductAuthor{
			ID:   id,
			Nome: nome,
		})
	}

	return normalized
}

func normalizeProductPhotos(items []entity.ProductPhoto) []entity.ProductPhoto {
	if len(items) == 0 {
		return []entity.ProductPhoto{}
	}

	normalized := make([]entity.ProductPhoto, 0, len(items))
	for index, item := range items {
		if len(normalized) >= 5 {
			break
		}

		image := sanitizeRepositoryImage(item.Image)
		if image == nil {
			continue
		}

		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = image.HashSHA256
		}
		if id == "" {
			id = strconv.Itoa(index + 1)
		}

		normalized = append(normalized, entity.ProductPhoto{
			ID:        id,
			Order:     len(normalized),
			IsPrimary: item.IsPrimary,
			Image:     image,
		})
	}

	if len(normalized) == 0 {
		return []entity.ProductPhoto{}
	}

	primaryIndex := 0
	for index, item := range normalized {
		if item.IsPrimary {
			primaryIndex = index
			break
		}
	}

	for index := range normalized {
		normalized[index].Order = index
		normalized[index].IsPrimary = index == primaryIndex
	}

	return normalized
}

func sanitizeRepositoryImage(asset *entity.ImageAsset) *entity.ImageAsset {
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

func preferredProductImage(photos []entity.ProductPhoto, fallback *entity.ImageAsset) *entity.ImageAsset {
	for _, item := range photos {
		if item.IsPrimary && item.Image != nil {
			return item.Image
		}
	}
	for _, item := range photos {
		if item.Image != nil {
			return item.Image
		}
	}
	return fallback
}

func normalizeSectionOrder(items []entity.StorefrontSection) []entity.StorefrontSection {
	valid := defaultSectionOrder()
	allowed := make(map[entity.StorefrontSection]struct{}, len(valid))
	for _, item := range valid {
		allowed[item] = struct{}{}
	}

	used := make(map[entity.StorefrontSection]struct{}, len(valid))
	normalized := make([]entity.StorefrontSection, 0, len(valid))
	for _, item := range items {
		if _, ok := allowed[item]; !ok {
			continue
		}
		if _, exists := used[item]; exists {
			continue
		}
		used[item] = struct{}{}
		normalized = append(normalized, item)
	}

	for _, item := range valid {
		if _, exists := used[item]; exists {
			continue
		}
		normalized = append(normalized, item)
	}

	return normalized
}

func normalizeSectionVisibility(items []entity.StorefrontSection) []entity.StorefrontSection {
	if items == nil {
		return defaultSectionVisibility()
	}

	valid := defaultSectionVisibility()
	allowed := make(map[entity.StorefrontSection]struct{}, len(valid))
	for _, item := range valid {
		allowed[item] = struct{}{}
	}

	used := make(map[entity.StorefrontSection]struct{}, len(items))
	normalized := make([]entity.StorefrontSection, 0, len(items))
	for _, item := range items {
		if _, ok := allowed[item]; !ok {
			continue
		}
		if _, exists := used[item]; exists {
			continue
		}
		used[item] = struct{}{}
		normalized = append(normalized, item)
	}

	return normalized
}

func defaultSectionOrder() []entity.StorefrontSection {
	return []entity.StorefrontSection{
		entity.StorefrontSectionBanner,
		entity.StorefrontSectionCustomBlock,
		entity.StorefrontSectionLaunches,
		entity.StorefrontSectionFeatured,
		entity.StorefrontSectionInstitutional,
		entity.StorefrontSectionPromotions,
	}
}

func defaultSectionVisibility() []entity.StorefrontSection {
	return []entity.StorefrontSection{
		entity.StorefrontSectionBanner,
		entity.StorefrontSectionCustomBlock,
		entity.StorefrontSectionLaunches,
		entity.StorefrontSectionFeatured,
		entity.StorefrontSectionInstitutional,
		entity.StorefrontSectionPromotions,
	}
}

var _ interface {
	GetSettings(ctx context.Context) (entity.SettingsRecord, error)
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
} = (*Repository)(nil)

var _ = pgx.ErrNoRows
