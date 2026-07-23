package entity

import "time"

type ImageAsset struct {
	Base64       string `json:"base64"`
	Mime         string `json:"mime"`
	Largura      int    `json:"largura"`
	Altura       int    `json:"altura"`
	TamanhoBytes int    `json:"tamanho_bytes"`
	HashSHA256   string `json:"hash_sha256"`
}

type BannerPayload struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	Subtitle         string      `json:"subtitle"`
	Link             string      `json:"link"`
	Order            int         `json:"order"`
	Active           bool        `json:"active"`
	ShowContent      bool        `json:"show_content"`
	LinkMode         string      `json:"link_mode"`
	ContentPosition  string      `json:"content_position"`
	ContentPositionX string      `json:"content_position_x"`
	ContentPositionY string      `json:"content_position_y"`
	UseButton        bool        `json:"use_button"`
	ButtonLabel      string      `json:"button_label"`
	Desktop          *ImageAsset `json:"desktop"`
	Mobile           *ImageAsset `json:"mobile"`
	CreatedAt        time.Time   `json:"-"`
	UpdatedAt        time.Time   `json:"-"`
}

type NavigationLink struct {
	Label   string `json:"label"`
	URL     string `json:"url"`
	Visible bool   `json:"visible"`
	Kind    string `json:"kind"`
}

type FeatureHighlight struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	Icon       string `json:"icon"`
	TextAlign  string `json:"text_align"`
	IconSize   string `json:"icon_size"`
	FontFamily string `json:"font_family"`
}

type ProductSectionDisplayMode string

const (
	ProductSectionDisplayModeWrap             ProductSectionDisplayMode = "wrap"
	ProductSectionDisplayModeHorizontalScroll ProductSectionDisplayMode = "horizontal_scroll"
)

type ProductSectionConfig struct {
	Eyebrow     string                    `json:"eyebrow"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	DisplayMode ProductSectionDisplayMode `json:"display_mode"`
}

type InstitutionalSectionConfig struct {
	Eyebrow     string `json:"eyebrow"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DisplayMode string `json:"display_mode"`
	WidthMode   string `json:"width_mode"`
	BackgroundColor string `json:"background_color"`
}

type ProductListingConfig struct {
	ShowBuyButton          bool   `json:"show_buy_button"`
	BuyButtonLabel         string `json:"buy_button_label"`
	BuyButtonUppercase     bool   `json:"buy_button_uppercase"`
	ShowAddToCartButton    bool   `json:"show_add_to_cart_button"`
	AddToCartButtonLabel   string `json:"add_to_cart_button_label"`
	AddToCartButtonUppercase bool `json:"add_to_cart_button_uppercase"`
	ShowPrice              bool   `json:"show_price"`
	ShowComparePrice       bool   `json:"show_compare_price"`
	ShowTags               bool   `json:"show_tags"`
	CardBackgroundColor    string `json:"card_background_color"`
	ShowBorder             bool   `json:"show_border"`
	BorderColor            string `json:"border_color"`
	BorderWidth            int    `json:"border_width"`
	ShowShadow             bool   `json:"show_shadow"`
	ShadowDirection        string `json:"shadow_direction"`
	ShadowIntensity        string `json:"shadow_intensity"`
	UseDefaultTitleColor   bool   `json:"use_default_title_color"`
	TitleTextColor         string `json:"title_text_color"`
	UseDefaultBodyColor    bool   `json:"use_default_body_color"`
	BodyTextColor          string `json:"body_text_color"`
	SecondaryButtonColor   string `json:"secondary_button_color"`
	SecondaryButtonTextColor string `json:"secondary_button_text_color"`
	ButtonShape            string `json:"button_shape"`
}

type StoreCategoryPayload struct {
	Nome      string `json:"nome"`
	Slug      string `json:"slug"`
	Descricao string `json:"descricao"`
	Ordem     int    `json:"ordem"`
	Ativa     bool   `json:"ativa"`
}

type StoreCategoryRecord struct {
	ID string
	StoreCategoryPayload
}

type AdminStoreCategoryListItem struct {
	ID            string `json:"id"`
	Nome          string `json:"nome"`
	Slug          string `json:"slug"`
	Descricao     string `json:"descricao"`
	Ordem         int    `json:"ordem"`
	Ativa         bool   `json:"ativa"`
	ProdutosCount int    `json:"produtos_count"`
	CriadoEm      string `json:"criado_em"`
	AtualizadoEm  string `json:"atualizado_em"`
}

type ProductPhoto struct {
	ID        string      `json:"id"`
	Order     int         `json:"order"`
	IsPrimary bool        `json:"is_primary"`
	Image     *ImageAsset `json:"image,omitempty"`
}

type StorefrontSection string

const (
	StorefrontSectionBanner        StorefrontSection = "banner"
	StorefrontSectionCustomBlock   StorefrontSection = "custom_block"
	StorefrontSectionLaunches      StorefrontSection = "launches"
	StorefrontSectionFeatured      StorefrontSection = "featured"
	StorefrontSectionInstitutional StorefrontSection = "institutional"
	StorefrontSectionPromotions    StorefrontSection = "promotions"
)

type SettingsRecord struct {
	StoreName             string               `json:"store_name"`
	StoreSubtitle         string               `json:"store_subtitle"`
	StoreDescription      string               `json:"store_description"`
	StoreTags             string               `json:"store_tags"`
	StoreInMaintenance    bool                 `json:"store_in_maintenance"`
	MaintenanceTitle      string               `json:"maintenance_title"`
	MaintenanceMessage    string               `json:"maintenance_message"`
	PrimaryColor          string               `json:"primary_color"`
	AccentColor           string               `json:"accent_color"`
	SecondaryColor        string               `json:"secondary_color"`
	BackgroundColor       string               `json:"background_color"`
	HeaderColor           string               `json:"header_color"`
	MenuColor             string               `json:"menu_color"`
	FooterColor           string               `json:"footer_color"`
	HeaderTextColor       string               `json:"header_text_color"`
	MenuTextColor         string               `json:"menu_text_color"`
	FooterTextColor       string               `json:"footer_text_color"`
	PrimaryTextColor      string               `json:"primary_text_color"`
	HighlightColor        string               `json:"highlight_color"`
	DiscountColor         string               `json:"discount_color"`
	LaunchBadgeColor      string               `json:"launch_badge_color"`
	FeaturedBadgeColor    string               `json:"featured_badge_color"`
	PromotionBadgeColor   string               `json:"promotion_badge_color"`
	ButtonColor           string               `json:"button_color"`
	ButtonTextColor       string               `json:"button_text_color"`
	ButtonStyle           string               `json:"button_style"`
	FontFamily            string               `json:"font_family"`
	FontSecondaryFamily   string               `json:"font_secondary_family"`
	FontMenuFamily        string               `json:"font_menu_family"`
	FontButtonFamily      string               `json:"font_button_family"`
	FontHighlightFamily   string               `json:"font_highlight_family"`
	FontImportURL         string               `json:"font_import_url"`
	FontEmbedCode         string               `json:"font_embed_code"`
	Integrations          IntegrationSettings  `json:"integrations"`
	CornerStyle           string               `json:"corner_style"`
	ContentWidthMode      string               `json:"content_width_mode"`
	MenuBackgroundMode    string               `json:"menu_background_mode"`
	BannerWidthMode       string               `json:"banner_width_mode"`
	BannerEffectMode      string               `json:"banner_effect_mode"`
	BannerRotationSeconds int                  `json:"banner_rotation_seconds"`
	BrandDisplayMode      string               `json:"brand_display_mode"`
	BrandLogoSize         string               `json:"brand_logo_size"`
	ProductsPerRowDesktop int                  `json:"products_per_row_desktop"`
	ProductsPerRowMobile  int                  `json:"products_per_row_mobile"`
	HeroTitle             string               `json:"hero_title"`
	HeroSubtitle          string               `json:"hero_subtitle"`
	Logo                  *ImageAsset          `json:"logo,omitempty"`
	Favicon               *ImageAsset          `json:"favicon,omitempty"`
	BannerDesktop         *ImageAsset          `json:"banner_desktop,omitempty"`
	BannerMobile          *ImageAsset          `json:"banner_mobile,omitempty"`
	Banners               []BannerPayload      `json:"banners"`
	MenuLinks             []NavigationLink     `json:"menu_links"`
	SectionOrder          []StorefrontSection  `json:"section_order"`
	VisibleSections       []StorefrontSection  `json:"visible_sections"`
	CustomBlockEyebrow    string               `json:"custom_block_eyebrow"`
	CustomBlockTitle      string               `json:"custom_block_title"`
	CustomBlockDescription string              `json:"custom_block_description"`
	CustomBlockHTML       string               `json:"custom_block_html"`
	CustomBlockCSS        string               `json:"custom_block_css"`
	CustomBlockJS         string               `json:"custom_block_js"`
	FeatureHighlights     []FeatureHighlight   `json:"feature_highlights"`
	InstitutionalSection  InstitutionalSectionConfig `json:"institutional_section"`
	ProductListingConfig  ProductListingConfig `json:"product_listing_config"`
	FooterLinks           []NavigationLink     `json:"footer_links"`
	FooterContactTitle    string               `json:"footer_contact_title"`
	FooterContactText     string               `json:"footer_contact_text"`
	LaunchesSection       ProductSectionConfig `json:"launches_section"`
	FeaturedSection       ProductSectionConfig `json:"featured_section"`
	PromotionsSection     ProductSectionConfig `json:"promotions_section"`
	CreatedAt             time.Time            `json:"-"`
	UpdatedAt             time.Time            `json:"-"`
	DraftUpdatedAt        time.Time            `json:"-"`
	PublishedAt           time.Time            `json:"-"`
	HasUnpublishedChanges bool                 `json:"-"`
}

type UpdateSettingsRequest struct {
	StoreName             string               `json:"store_name"`
	StoreSubtitle         string               `json:"store_subtitle"`
	StoreDescription      string               `json:"store_description"`
	StoreTags             string               `json:"store_tags"`
	StoreInMaintenance    bool                 `json:"store_in_maintenance"`
	MaintenanceTitle      string               `json:"maintenance_title"`
	MaintenanceMessage    string               `json:"maintenance_message"`
	PrimaryColor          string               `json:"primary_color"`
	AccentColor           string               `json:"accent_color"`
	SecondaryColor        string               `json:"secondary_color"`
	BackgroundColor       string               `json:"background_color"`
	HeaderColor           string               `json:"header_color"`
	MenuColor             string               `json:"menu_color"`
	FooterColor           string               `json:"footer_color"`
	HeaderTextColor       string               `json:"header_text_color"`
	MenuTextColor         string               `json:"menu_text_color"`
	FooterTextColor       string               `json:"footer_text_color"`
	PrimaryTextColor      string               `json:"primary_text_color"`
	HighlightColor        string               `json:"highlight_color"`
	DiscountColor         string               `json:"discount_color"`
	LaunchBadgeColor      string               `json:"launch_badge_color"`
	FeaturedBadgeColor    string               `json:"featured_badge_color"`
	PromotionBadgeColor   string               `json:"promotion_badge_color"`
	ButtonColor           string               `json:"button_color"`
	ButtonTextColor       string               `json:"button_text_color"`
	ButtonStyle           string               `json:"button_style"`
	FontFamily            string               `json:"font_family"`
	FontSecondaryFamily   string               `json:"font_secondary_family"`
	FontMenuFamily        string               `json:"font_menu_family"`
	FontButtonFamily      string               `json:"font_button_family"`
	FontHighlightFamily   string               `json:"font_highlight_family"`
	FontImportURL         string               `json:"font_import_url"`
	FontEmbedCode         string               `json:"font_embed_code"`
	Integrations          IntegrationSettings  `json:"integrations"`
	CornerStyle           string               `json:"corner_style"`
	ContentWidthMode      string               `json:"content_width_mode"`
	MenuBackgroundMode    string               `json:"menu_background_mode"`
	BannerWidthMode       string               `json:"banner_width_mode"`
	BannerEffectMode      string               `json:"banner_effect_mode"`
	BannerRotationSeconds int                  `json:"banner_rotation_seconds"`
	BrandDisplayMode      string               `json:"brand_display_mode"`
	BrandLogoSize         string               `json:"brand_logo_size"`
	ProductsPerRowDesktop int                  `json:"products_per_row_desktop"`
	ProductsPerRowMobile  int                  `json:"products_per_row_mobile"`
	HeroTitle             string               `json:"hero_title"`
	HeroSubtitle          string               `json:"hero_subtitle"`
	Logo                  *ImageAsset          `json:"logo"`
	Favicon               *ImageAsset          `json:"favicon"`
	BannerDesktop         *ImageAsset          `json:"banner_desktop"`
	BannerMobile          *ImageAsset          `json:"banner_mobile"`
	Banners               []BannerPayload      `json:"banners"`
	MenuLinks             []NavigationLink     `json:"menu_links"`
	SectionOrder          []StorefrontSection  `json:"section_order"`
	VisibleSections       []StorefrontSection  `json:"visible_sections"`
	CustomBlockEyebrow    string               `json:"custom_block_eyebrow"`
	CustomBlockTitle      string               `json:"custom_block_title"`
	CustomBlockDescription string              `json:"custom_block_description"`
	CustomBlockHTML       string               `json:"custom_block_html"`
	CustomBlockCSS        string               `json:"custom_block_css"`
	CustomBlockJS         string               `json:"custom_block_js"`
	FeatureHighlights     []FeatureHighlight   `json:"feature_highlights"`
	InstitutionalSection  InstitutionalSectionConfig `json:"institutional_section"`
	ProductListingConfig  ProductListingConfig `json:"product_listing_config"`
	FooterLinks           []NavigationLink     `json:"footer_links"`
	FooterContactTitle    string               `json:"footer_contact_title"`
	FooterContactText     string               `json:"footer_contact_text"`
	LaunchesSection       ProductSectionConfig `json:"launches_section"`
	FeaturedSection       ProductSectionConfig `json:"featured_section"`
	PromotionsSection     ProductSectionConfig `json:"promotions_section"`
}

type SettingsResponse struct {
	StoreName             string               `json:"store_name"`
	StoreSubtitle         string               `json:"store_subtitle"`
	StoreDescription      string               `json:"store_description"`
	StoreTags             string               `json:"store_tags"`
	StoreInMaintenance    bool                 `json:"store_in_maintenance"`
	MaintenanceTitle      string               `json:"maintenance_title"`
	MaintenanceMessage    string               `json:"maintenance_message"`
	PrimaryColor          string               `json:"primary_color"`
	AccentColor           string               `json:"accent_color"`
	SecondaryColor        string               `json:"secondary_color"`
	BackgroundColor       string               `json:"background_color"`
	HeaderColor           string               `json:"header_color"`
	MenuColor             string               `json:"menu_color"`
	FooterColor           string               `json:"footer_color"`
	HeaderTextColor       string               `json:"header_text_color"`
	MenuTextColor         string               `json:"menu_text_color"`
	FooterTextColor       string               `json:"footer_text_color"`
	PrimaryTextColor      string               `json:"primary_text_color"`
	HighlightColor        string               `json:"highlight_color"`
	DiscountColor         string               `json:"discount_color"`
	LaunchBadgeColor      string               `json:"launch_badge_color"`
	FeaturedBadgeColor    string               `json:"featured_badge_color"`
	PromotionBadgeColor   string               `json:"promotion_badge_color"`
	ButtonColor           string               `json:"button_color"`
	ButtonTextColor       string               `json:"button_text_color"`
	ButtonStyle           string               `json:"button_style"`
	FontFamily            string               `json:"font_family"`
	FontSecondaryFamily   string               `json:"font_secondary_family"`
	FontMenuFamily        string               `json:"font_menu_family"`
	FontButtonFamily      string               `json:"font_button_family"`
	FontHighlightFamily   string               `json:"font_highlight_family"`
	FontImportURL         string               `json:"font_import_url"`
	FontEmbedCode         string               `json:"font_embed_code"`
	Integrations          IntegrationSettings  `json:"integrations"`
	CornerStyle           string               `json:"corner_style"`
	ContentWidthMode      string               `json:"content_width_mode"`
	MenuBackgroundMode    string               `json:"menu_background_mode"`
	BannerWidthMode       string               `json:"banner_width_mode"`
	BannerEffectMode      string               `json:"banner_effect_mode"`
	BannerRotationSeconds int                  `json:"banner_rotation_seconds"`
	BrandDisplayMode      string               `json:"brand_display_mode"`
	BrandLogoSize         string               `json:"brand_logo_size"`
	ProductsPerRowDesktop int                  `json:"products_per_row_desktop"`
	ProductsPerRowMobile  int                  `json:"products_per_row_mobile"`
	HeroTitle             string               `json:"hero_title"`
	HeroSubtitle          string               `json:"hero_subtitle"`
	Logo                  *ImageAsset          `json:"logo,omitempty"`
	Favicon               *ImageAsset          `json:"favicon,omitempty"`
	BannerDesktop         *ImageAsset          `json:"banner_desktop,omitempty"`
	BannerMobile          *ImageAsset          `json:"banner_mobile,omitempty"`
	Banners               []BannerPayload      `json:"banners"`
	MenuLinks             []NavigationLink     `json:"menu_links"`
	SectionOrder          []StorefrontSection  `json:"section_order"`
	VisibleSections       []StorefrontSection  `json:"visible_sections"`
	CustomBlockEyebrow    string               `json:"custom_block_eyebrow"`
	CustomBlockTitle      string               `json:"custom_block_title"`
	CustomBlockDescription string              `json:"custom_block_description"`
	CustomBlockHTML       string               `json:"custom_block_html"`
	CustomBlockCSS        string               `json:"custom_block_css"`
	CustomBlockJS         string               `json:"custom_block_js"`
	FeatureHighlights     []FeatureHighlight   `json:"feature_highlights"`
	InstitutionalSection  InstitutionalSectionConfig `json:"institutional_section"`
	ProductListingConfig  ProductListingConfig `json:"product_listing_config"`
	FooterLinks           []NavigationLink     `json:"footer_links"`
	FooterContactTitle    string               `json:"footer_contact_title"`
	FooterContactText     string               `json:"footer_contact_text"`
	LaunchesSection       ProductSectionConfig `json:"launches_section"`
	FeaturedSection       ProductSectionConfig `json:"featured_section"`
	PromotionsSection     ProductSectionConfig `json:"promotions_section"`
	CreatedAt             string               `json:"created_at"`
	UpdatedAt             string               `json:"updated_at"`
	DraftUpdatedAt        string               `json:"draft_updated_at"`
	PublishedAt           string               `json:"published_at"`
	HasUnpublishedChanges bool                 `json:"has_unpublished_changes"`
}

type ProductPayload struct {
	LivroID          string          `json:"livro_id"`
	IsBook           bool            `json:"is_book"`
	Authors          []ProductAuthor `json:"authors"`
	AutorNome        string          `json:"autor_nome"`
	Marca            string          `json:"marca"`
	Editora          string          `json:"editora"`
	Subtitulo        string          `json:"subtitulo"`
	Sinopse          string          `json:"sinopse"`
	ISBN             string          `json:"isbn"`
	CodigoBarra      string          `json:"codigo_barra"`
	Edicao           string          `json:"edicao"`
	Idioma           string          `json:"idioma"`
	NumeroPaginas    *int            `json:"numero_paginas,omitempty"`
	Genero           string          `json:"genero"`
	DataPublicacao   string          `json:"data_publicacao"`
	TipoCapa         string          `json:"tipo_capa"`
	PesoGramas       *int            `json:"peso_gramas,omitempty"`
	LarguraCm        *float64        `json:"largura_cm,omitempty"`
	AlturaCm         *float64        `json:"altura_cm,omitempty"`
	ProfundidadeCm   *float64        `json:"profundidade_cm,omitempty"`
	Slug             string          `json:"slug"`
	NomeExibicao     string          `json:"nome_exibicao"`
	DescricaoCurta   string          `json:"descricao_curta"`
	Categoria        string          `json:"categoria"`
	Categorias       []string        `json:"categorias"`
	PrecoVenda       float64         `json:"preco_venda"`
	EmPromocao       bool            `json:"em_promocao"`
	PrecoPromocional float64         `json:"preco_promocional"`
	Destaque         bool            `json:"destaque"`
	Lancamento       bool            `json:"lancamento"`
	Ativo            bool            `json:"ativo"`
	Ordem            int             `json:"ordem"`
	Fotos            []ProductPhoto  `json:"fotos"`
}

type ProductRecord struct {
	ID string
	ProductPayload
}

type ProductAuthor struct {
	ID   string `json:"id,omitempty"`
	Nome string `json:"nome"`
}

type AdminProductListItem struct {
	ID               string          `json:"id"`
	LivroID          string          `json:"livro_id"`
	IsBook           bool            `json:"is_book"`
	Authors          []ProductAuthor `json:"authors"`
	LivroTitulo      string          `json:"livro_titulo"`
	LivroSubtitulo   *string         `json:"livro_subtitulo,omitempty"`
	AutorNome        *string         `json:"autor_nome,omitempty"`
	Marca            *string         `json:"marca,omitempty"`
	Editora          *string         `json:"editora,omitempty"`
	Subtitulo        *string         `json:"subtitulo,omitempty"`
	Sinopse          *string         `json:"sinopse,omitempty"`
	ISBN             *string         `json:"isbn,omitempty"`
	CodigoBarra      *string         `json:"codigo_barra,omitempty"`
	Edicao           *string         `json:"edicao,omitempty"`
	Idioma           *string         `json:"idioma,omitempty"`
	NumeroPaginas    *int            `json:"numero_paginas,omitempty"`
	Genero           *string         `json:"genero,omitempty"`
	DataPublicacao   *string         `json:"data_publicacao,omitempty"`
	TipoCapa         *string         `json:"tipo_capa,omitempty"`
	PesoGramas       *int            `json:"peso_gramas,omitempty"`
	LarguraCm        *float64        `json:"largura_cm,omitempty"`
	AlturaCm         *float64        `json:"altura_cm,omitempty"`
	ProfundidadeCm   *float64        `json:"profundidade_cm,omitempty"`
	Slug             string          `json:"slug"`
	NomeExibicao     string          `json:"nome_exibicao"`
	DescricaoCurta   string          `json:"descricao_curta"`
	Categoria        string          `json:"categoria"`
	Categorias       []string        `json:"categorias"`
	PrecoVenda       float64         `json:"preco_venda"`
	EmPromocao       bool            `json:"em_promocao"`
	PrecoPromocional float64         `json:"preco_promocional"`
	Destaque         bool            `json:"destaque"`
	Lancamento       bool            `json:"lancamento"`
	Ativo            bool            `json:"ativo"`
	Ordem            int             `json:"ordem"`
	Capa             *ImageAsset     `json:"capa,omitempty"`
	Fotos            []ProductPhoto  `json:"fotos"`
	CriadoEm         string          `json:"criado_em"`
	AtualizadoEm     string          `json:"atualizado_em"`
}

type AdminProductDetail = AdminProductListItem

type PublicConfigResponse struct {
	StoreName             string               `json:"store_name"`
	StoreSubtitle         string               `json:"store_subtitle"`
	StoreDescription      string               `json:"store_description"`
	StoreTags             string               `json:"store_tags"`
	StoreInMaintenance    bool                 `json:"store_in_maintenance"`
	MaintenanceTitle      string               `json:"maintenance_title"`
	MaintenanceMessage    string               `json:"maintenance_message"`
	PrimaryColor          string               `json:"primary_color"`
	AccentColor           string               `json:"accent_color"`
	SecondaryColor        string               `json:"secondary_color"`
	BackgroundColor       string               `json:"background_color"`
	HeaderColor           string               `json:"header_color"`
	MenuColor             string               `json:"menu_color"`
	FooterColor           string               `json:"footer_color"`
	HeaderTextColor       string               `json:"header_text_color"`
	MenuTextColor         string               `json:"menu_text_color"`
	FooterTextColor       string               `json:"footer_text_color"`
	PrimaryTextColor      string               `json:"primary_text_color"`
	HighlightColor        string               `json:"highlight_color"`
	DiscountColor         string               `json:"discount_color"`
	LaunchBadgeColor      string               `json:"launch_badge_color"`
	FeaturedBadgeColor    string               `json:"featured_badge_color"`
	PromotionBadgeColor   string               `json:"promotion_badge_color"`
	ButtonColor           string               `json:"button_color"`
	ButtonTextColor       string               `json:"button_text_color"`
	ButtonStyle           string               `json:"button_style"`
	FontFamily            string               `json:"font_family"`
	FontSecondaryFamily   string               `json:"font_secondary_family"`
	FontMenuFamily        string               `json:"font_menu_family"`
	FontButtonFamily      string               `json:"font_button_family"`
	FontHighlightFamily   string               `json:"font_highlight_family"`
	FontImportURL         string               `json:"font_import_url"`
	FontEmbedCode         string               `json:"font_embed_code"`
	Integrations          IntegrationSettings  `json:"integrations"`
	CornerStyle           string               `json:"corner_style"`
	ContentWidthMode      string               `json:"content_width_mode"`
	MenuBackgroundMode    string               `json:"menu_background_mode"`
	BannerWidthMode       string               `json:"banner_width_mode"`
	BannerEffectMode      string               `json:"banner_effect_mode"`
	BannerRotationSeconds int                  `json:"banner_rotation_seconds"`
	BrandDisplayMode      string               `json:"brand_display_mode"`
	BrandLogoSize         string               `json:"brand_logo_size"`
	ProductsPerRowDesktop int                  `json:"products_per_row_desktop"`
	ProductsPerRowMobile  int                  `json:"products_per_row_mobile"`
	HeroTitle             string               `json:"hero_title"`
	HeroSubtitle          string               `json:"hero_subtitle"`
	Logo                  *ImageAsset          `json:"logo,omitempty"`
	Favicon               *ImageAsset          `json:"favicon,omitempty"`
	BannerDesktop         *ImageAsset          `json:"banner_desktop,omitempty"`
	BannerMobile          *ImageAsset          `json:"banner_mobile,omitempty"`
	Banners               []BannerPayload      `json:"banners"`
	MenuLinks             []NavigationLink     `json:"menu_links"`
	SectionOrder          []StorefrontSection  `json:"section_order"`
	VisibleSections       []StorefrontSection  `json:"visible_sections"`
	CustomBlockEyebrow    string               `json:"custom_block_eyebrow"`
	CustomBlockTitle      string               `json:"custom_block_title"`
	CustomBlockDescription string              `json:"custom_block_description"`
	CustomBlockHTML       string               `json:"custom_block_html"`
	CustomBlockCSS        string               `json:"custom_block_css"`
	CustomBlockJS         string               `json:"custom_block_js"`
	FeatureHighlights     []FeatureHighlight   `json:"feature_highlights"`
	InstitutionalSection  InstitutionalSectionConfig `json:"institutional_section"`
	ProductListingConfig  ProductListingConfig `json:"product_listing_config"`
	FooterLinks           []NavigationLink     `json:"footer_links"`
	FooterContactTitle    string               `json:"footer_contact_title"`
	FooterContactText     string               `json:"footer_contact_text"`
	LaunchesSection       ProductSectionConfig `json:"launches_section"`
	FeaturedSection       ProductSectionConfig `json:"featured_section"`
	PromotionsSection     ProductSectionConfig `json:"promotions_section"`
}

type PublicProductListItem struct {
	ID               string          `json:"id"`
	Slug             string          `json:"slug"`
	NomeExibicao     string          `json:"nome_exibicao"`
	DescricaoCurta   string          `json:"descricao_curta"`
	Categoria        string          `json:"categoria"`
	Categorias       []string        `json:"categorias"`
	Ordem            int             `json:"ordem"`
	PrecoVenda       float64         `json:"preco_venda"`
	EmPromocao       bool            `json:"em_promocao"`
	PrecoPromocional float64         `json:"preco_promocional"`
	Destaque         bool            `json:"destaque"`
	Lancamento       bool            `json:"lancamento"`
	IsBook           bool            `json:"is_book"`
	Authors          []ProductAuthor `json:"authors"`
	AutorNome        string          `json:"autor_nome"`
	Marca            string          `json:"marca"`
	Editora          string          `json:"editora"`
	LivroTitulo      string          `json:"livro_titulo"`
	LivroSubtitulo   *string         `json:"livro_subtitulo,omitempty"`
	PossuiDigital    bool            `json:"possui_formato_digital"`
	PossuiFisico     bool            `json:"possui_formato_fisico"`
	URLCompraDigital *string         `json:"url_compra_digital,omitempty"`
	Capa             *ImageAsset     `json:"capa,omitempty"`
	Fotos            []ProductPhoto  `json:"fotos"`
}

type IntegrationSettings struct {
	FacebookPixelID          string `json:"facebook_pixel_id"`
	GoogleAdsID              string `json:"google_ads_id"`
	GoogleAdsConversionLabel string `json:"google_ads_conversion_label"`
	GoogleAnalyticsID        string `json:"google_analytics_id"`
	GoogleTagManagerID       string `json:"google_tag_manager_id"`
	MicrosoftClarityID       string `json:"microsoft_clarity_id"`
	TikTokPixelID            string `json:"tiktok_pixel_id"`
}

type PublicProductDetail struct {
	ID               string          `json:"id"`
	Slug             string          `json:"slug"`
	NomeExibicao     string          `json:"nome_exibicao"`
	DescricaoCurta   string          `json:"descricao_curta"`
	Categoria        string          `json:"categoria"`
	Categorias       []string        `json:"categorias"`
	Ordem            int             `json:"ordem"`
	PrecoVenda       float64         `json:"preco_venda"`
	EmPromocao       bool            `json:"em_promocao"`
	PrecoPromocional float64         `json:"preco_promocional"`
	Destaque         bool            `json:"destaque"`
	Lancamento       bool            `json:"lancamento"`
	IsBook           bool            `json:"is_book"`
	Authors          []ProductAuthor `json:"authors"`
	AutorNome        string          `json:"autor_nome"`
	Marca            string          `json:"marca"`
	Editora          string          `json:"editora"`
	LivroTitulo      string          `json:"livro_titulo"`
	LivroSubtitulo   *string         `json:"livro_subtitulo,omitempty"`
	Sinopse          *string         `json:"sinopse,omitempty"`
	PossuiDigital    bool            `json:"possui_formato_digital"`
	PossuiFisico     bool            `json:"possui_formato_fisico"`
	URLCompraDigital *string         `json:"url_compra_digital,omitempty"`
	Capa             *ImageAsset     `json:"capa,omitempty"`
	Fotos            []ProductPhoto  `json:"fotos"`
}
