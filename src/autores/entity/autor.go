package entity

type PhotoInput struct {
	Base64       string `json:"base64"`
	Mime         string `json:"mime"`
	Largura      int    `json:"largura"`
	Altura       int    `json:"altura"`
	TamanhoBytes int    `json:"tamanho_bytes"`
	HashSHA256   string `json:"hash_sha256"`
}

type CreateRequest struct {
	UsuarioID            string      `json:"usuario_id"`
	NomeCompleto         string      `json:"nome_completo"`
	NomePublico          string      `json:"nome_publico"`
	Email                string      `json:"email"`
	EmailPrivado         bool        `json:"email_privado"`
	Whatsapp             string      `json:"whatsapp"`
	WhatsappPrivado      bool        `json:"whatsapp_privado"`
	Instagram            string      `json:"instagram"`
	InstagramPrivado     bool        `json:"instagram_privado"`
	Wattpad              string      `json:"wattpad"`
	WattpadPrivado       bool        `json:"wattpad_privado"`
	Facebook             string      `json:"facebook"`
	FacebookPrivado      bool        `json:"facebook_privado"`
	XTwitter             string      `json:"x_twitter"`
	XTwitterPrivado      bool        `json:"x_twitter_privado"`
	Tiktok               string      `json:"tiktok"`
	TiktokPrivado        bool        `json:"tiktok_privado"`
	Youtube              string      `json:"youtube"`
	YoutubePrivado       bool        `json:"youtube_privado"`
	Linkedin             string      `json:"linkedin"`
	LinkedinPrivado      bool        `json:"linkedin_privado"`
	Nacionalidade        string      `json:"nacionalidade"`
	Biografia            string      `json:"biografia"`
	Foto                 *PhotoInput `json:"foto"`
	Status               string      `json:"status"`
}

type UpdateRequest = CreateRequest

type PersistInput struct {
	ID                  string
	UsuarioID           *string
	NomeCompleto        string
	NomePublico         *string
	Email               *string
	EmailPrivado        bool
	Whatsapp            *string
	WhatsappPrivado     bool
	Instagram           *string
	InstagramPrivado    bool
	Wattpad             *string
	WattpadPrivado      bool
	Facebook            *string
	FacebookPrivado     bool
	XTwitter            *string
	XTwitterPrivado     bool
	Tiktok              *string
	TiktokPrivado       bool
	Youtube             *string
	YoutubePrivado      bool
	Linkedin            *string
	LinkedinPrivado     bool
	Nacionalidade       *string
	Biografia           *string
	Foto                *PhotoInput
	Status              string
}

type ListItem struct {
	ID                 string      `json:"id"`
	UsuarioID          *string     `json:"usuario_id,omitempty"`
	NomeCompleto       string      `json:"nome_completo"`
	NomePublico        *string     `json:"nome_publico,omitempty"`
	NomeExibicao       string      `json:"nome_exibicao"`
	Email              *string     `json:"email,omitempty"`
	EmailPrivado       bool        `json:"email_privado"`
	Whatsapp           *string     `json:"whatsapp,omitempty"`
	WhatsappPrivado    bool        `json:"whatsapp_privado"`
	Instagram          *string     `json:"instagram,omitempty"`
	InstagramPrivado   bool        `json:"instagram_privado"`
	Wattpad            *string     `json:"wattpad,omitempty"`
	WattpadPrivado     bool        `json:"wattpad_privado"`
	Facebook           *string     `json:"facebook,omitempty"`
	FacebookPrivado    bool        `json:"facebook_privado"`
	XTwitter           *string     `json:"x_twitter,omitempty"`
	XTwitterPrivado    bool        `json:"x_twitter_privado"`
	Tiktok             *string     `json:"tiktok,omitempty"`
	TiktokPrivado      bool        `json:"tiktok_privado"`
	Youtube            *string     `json:"youtube,omitempty"`
	YoutubePrivado     bool        `json:"youtube_privado"`
	Linkedin           *string     `json:"linkedin,omitempty"`
	LinkedinPrivado    bool        `json:"linkedin_privado"`
	Nacionalidade      *string     `json:"nacionalidade,omitempty"`
	Status             string      `json:"status"`
	UsuarioNome        *string     `json:"usuario_nome,omitempty"`
	PossuiFoto         bool        `json:"possui_foto"`
	Foto               *PhotoInput `json:"foto,omitempty"`
	CriadoEm           string      `json:"criado_em"`
	AtualizadoEm       string      `json:"atualizado_em"`
}

type DetailResponse struct {
	ID                 string      `json:"id"`
	UsuarioID          *string     `json:"usuario_id,omitempty"`
	NomeCompleto       string      `json:"nome_completo"`
	NomePublico        *string     `json:"nome_publico,omitempty"`
	Email              *string     `json:"email,omitempty"`
	EmailPrivado       bool        `json:"email_privado"`
	Whatsapp           *string     `json:"whatsapp,omitempty"`
	WhatsappPrivado    bool        `json:"whatsapp_privado"`
	Instagram          *string     `json:"instagram,omitempty"`
	InstagramPrivado   bool        `json:"instagram_privado"`
	Wattpad            *string     `json:"wattpad,omitempty"`
	WattpadPrivado     bool        `json:"wattpad_privado"`
	Facebook           *string     `json:"facebook,omitempty"`
	FacebookPrivado    bool        `json:"facebook_privado"`
	XTwitter           *string     `json:"x_twitter,omitempty"`
	XTwitterPrivado    bool        `json:"x_twitter_privado"`
	Tiktok             *string     `json:"tiktok,omitempty"`
	TiktokPrivado      bool        `json:"tiktok_privado"`
	Youtube            *string     `json:"youtube,omitempty"`
	YoutubePrivado     bool        `json:"youtube_privado"`
	Linkedin           *string     `json:"linkedin,omitempty"`
	LinkedinPrivado    bool        `json:"linkedin_privado"`
	Nacionalidade      *string     `json:"nacionalidade,omitempty"`
	Biografia          *string     `json:"biografia,omitempty"`
	Status             string      `json:"status"`
	UsuarioNome        *string     `json:"usuario_nome,omitempty"`
	Foto               *PhotoInput `json:"foto,omitempty"`
	CriadoEm           string      `json:"criado_em"`
	AtualizadoEm       string      `json:"atualizado_em"`
}

type ListQuery struct {
	Search string
	Status string
}
