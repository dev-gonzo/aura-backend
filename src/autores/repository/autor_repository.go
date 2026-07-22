package repository

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"sistema-editorial/editora/backend/src/autores/entity"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, input entity.PersistInput) (string, error) {
	const query = `
		INSERT INTO autores (
			usuario_id,
			nome_completo,
			nome_publico,
			email,
			email_privado,
			whatsapp,
			whatsapp_privado,
			instagram,
			instagram_privado,
			wattpad,
			wattpad_privado,
			facebook,
			facebook_privado,
			x_twitter,
			x_twitter_privado,
			tiktok,
			tiktok_privado,
			youtube,
			youtube_privado,
			linkedin,
			linkedin_privado,
			nacionalidade,
			biografia,
			foto_base64,
			foto_mime,
			foto_largura,
			foto_altura,
			foto_tamanho_bytes,
			foto_hash_sha256,
			status
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30
		)
		RETURNING id
	`

	var id string
	err := r.pool.QueryRow(
		ctx,
		query,
		input.UsuarioID,
		input.NomeCompleto,
		input.NomePublico,
		input.Email,
		input.EmailPrivado,
		input.Whatsapp,
		input.WhatsappPrivado,
		input.Instagram,
		input.InstagramPrivado,
		input.Wattpad,
		input.WattpadPrivado,
		input.Facebook,
		input.FacebookPrivado,
		input.XTwitter,
		input.XTwitterPrivado,
		input.Tiktok,
		input.TiktokPrivado,
		input.Youtube,
		input.YoutubePrivado,
		input.Linkedin,
		input.LinkedinPrivado,
		input.Nacionalidade,
		input.Biografia,
		readPhotoBase64(input.Foto),
		readPhotoMime(input.Foto),
		readPhotoWidth(input.Foto),
		readPhotoHeight(input.Foto),
		readPhotoSize(input.Foto),
		readPhotoHash(input.Foto),
		input.Status,
	).Scan(&id)
	return id, err
}

func (r *PostgresRepository) Update(ctx context.Context, input entity.PersistInput) error {
	const query = `
		UPDATE autores
		SET
			usuario_id = $2,
			nome_completo = $3,
			nome_publico = $4,
			email = $5,
			email_privado = $6,
			whatsapp = $7,
			whatsapp_privado = $8,
			instagram = $9,
			instagram_privado = $10,
			wattpad = $11,
			wattpad_privado = $12,
			facebook = $13,
			facebook_privado = $14,
			x_twitter = $15,
			x_twitter_privado = $16,
			tiktok = $17,
			tiktok_privado = $18,
			youtube = $19,
			youtube_privado = $20,
			linkedin = $21,
			linkedin_privado = $22,
			nacionalidade = $23,
			biografia = $24,
			foto_base64 = $25,
			foto_mime = $26,
			foto_largura = $27,
			foto_altura = $28,
			foto_tamanho_bytes = $29,
			foto_hash_sha256 = $30,
			status = $31,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		input.ID,
		input.UsuarioID,
		input.NomeCompleto,
		input.NomePublico,
		input.Email,
		input.EmailPrivado,
		input.Whatsapp,
		input.WhatsappPrivado,
		input.Instagram,
		input.InstagramPrivado,
		input.Wattpad,
		input.WattpadPrivado,
		input.Facebook,
		input.FacebookPrivado,
		input.XTwitter,
		input.XTwitterPrivado,
		input.Tiktok,
		input.TiktokPrivado,
		input.Youtube,
		input.YoutubePrivado,
		input.Linkedin,
		input.LinkedinPrivado,
		input.Nacionalidade,
		input.Biografia,
		readPhotoBase64(input.Foto),
		readPhotoMime(input.Foto),
		readPhotoWidth(input.Foto),
		readPhotoHeight(input.Foto),
		readPhotoSize(input.Foto),
		readPhotoHash(input.Foto),
		input.Status,
	)
	return err
}

func (r *PostgresRepository) List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error) {
	const sql = `
		SELECT
			a.id,
			a.usuario_id,
			a.nome_completo,
			a.nome_publico,
			COALESCE(NULLIF(a.nome_publico, ''), a.nome_completo) AS nome_exibicao,
			a.email,
			a.email_privado,
			a.whatsapp,
			a.whatsapp_privado,
			a.instagram,
			a.instagram_privado,
			a.wattpad,
			a.wattpad_privado,
			a.facebook,
			a.facebook_privado,
			a.x_twitter,
			a.x_twitter_privado,
			a.tiktok,
			a.tiktok_privado,
			a.youtube,
			a.youtube_privado,
			a.linkedin,
			a.linkedin_privado,
			a.nacionalidade,
			a.status,
			u.nome_completo,
			a.foto_base64,
			a.foto_mime,
			a.foto_largura,
			a.foto_altura,
			a.foto_tamanho_bytes,
			a.foto_hash_sha256,
			to_char(a.criado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(a.atualizado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM autores a
		LEFT JOIN usuarios u ON u.id = a.usuario_id
		WHERE
			($1 = '' OR a.status = $1)
			AND (
				$2 = ''
				OR a.nome_completo ILIKE '%' || $2 || '%'
				OR COALESCE(a.nome_publico, '') ILIKE '%' || $2 || '%'
				OR COALESCE(a.email, '') ILIKE '%' || $2 || '%'
			)
		ORDER BY a.nome_completo ASC
	`

	rows, err := r.pool.Query(ctx, sql, strings.TrimSpace(query.Status), strings.TrimSpace(query.Search))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.ListItem, 0)
	for rows.Next() {
		var item entity.ListItem
		var fotoBase64, fotoMime, fotoHash *string
		var fotoLargura, fotoAltura, fotoTamanho *int

		if err := rows.Scan(
			&item.ID,
			&item.UsuarioID,
			&item.NomeCompleto,
			&item.NomePublico,
			&item.NomeExibicao,
			&item.Email,
			&item.EmailPrivado,
			&item.Whatsapp,
			&item.WhatsappPrivado,
			&item.Instagram,
			&item.InstagramPrivado,
			&item.Wattpad,
			&item.WattpadPrivado,
			&item.Facebook,
			&item.FacebookPrivado,
			&item.XTwitter,
			&item.XTwitterPrivado,
			&item.Tiktok,
			&item.TiktokPrivado,
			&item.Youtube,
			&item.YoutubePrivado,
			&item.Linkedin,
			&item.LinkedinPrivado,
			&item.Nacionalidade,
			&item.Status,
			&item.UsuarioNome,
			&fotoBase64,
			&fotoMime,
			&fotoLargura,
			&fotoAltura,
			&fotoTamanho,
			&fotoHash,
			&item.CriadoEm,
			&item.AtualizadoEm,
		); err != nil {
			return nil, err
		}

		item.Foto = buildPhoto(fotoBase64, fotoMime, fotoLargura, fotoAltura, fotoTamanho, fotoHash)
		item.PossuiFoto = item.Foto != nil
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (entity.DetailResponse, error) {
	const query = `
		SELECT
			a.id,
			a.usuario_id,
			a.nome_completo,
			a.nome_publico,
			a.email,
			a.email_privado,
			a.whatsapp,
			a.whatsapp_privado,
			a.instagram,
			a.instagram_privado,
			a.wattpad,
			a.wattpad_privado,
			a.facebook,
			a.facebook_privado,
			a.x_twitter,
			a.x_twitter_privado,
			a.tiktok,
			a.tiktok_privado,
			a.youtube,
			a.youtube_privado,
			a.linkedin,
			a.linkedin_privado,
			a.nacionalidade,
			a.biografia,
			a.status,
			u.nome_completo,
			a.foto_base64,
			a.foto_mime,
			a.foto_largura,
			a.foto_altura,
			a.foto_tamanho_bytes,
			a.foto_hash_sha256,
			to_char(a.criado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(a.atualizado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM autores a
		LEFT JOIN usuarios u ON u.id = a.usuario_id
		WHERE a.id = $1
	`

	var item entity.DetailResponse
	var fotoBase64, fotoMime, fotoHash *string
	var fotoLargura, fotoAltura, fotoTamanho *int

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.UsuarioID,
		&item.NomeCompleto,
		&item.NomePublico,
		&item.Email,
		&item.EmailPrivado,
		&item.Whatsapp,
		&item.WhatsappPrivado,
		&item.Instagram,
		&item.InstagramPrivado,
		&item.Wattpad,
		&item.WattpadPrivado,
		&item.Facebook,
		&item.FacebookPrivado,
		&item.XTwitter,
		&item.XTwitterPrivado,
		&item.Tiktok,
		&item.TiktokPrivado,
		&item.Youtube,
		&item.YoutubePrivado,
		&item.Linkedin,
		&item.LinkedinPrivado,
		&item.Nacionalidade,
		&item.Biografia,
		&item.Status,
		&item.UsuarioNome,
		&fotoBase64,
		&fotoMime,
		&fotoLargura,
		&fotoAltura,
		&fotoTamanho,
		&fotoHash,
		&item.CriadoEm,
		&item.AtualizadoEm,
	)
	if err != nil {
		return entity.DetailResponse{}, err
	}

	item.Foto = buildPhoto(fotoBase64, fotoMime, fotoLargura, fotoAltura, fotoTamanho, fotoHash)
	return item, nil
}

func (r *PostgresRepository) ExistsByEmail(ctx context.Context, email string, excludeID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM autores
			WHERE lower(coalesce(email, '')) = lower($1)
			  AND (NULLIF($2, '')::uuid IS NULL OR id <> NULLIF($2, '')::uuid)
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(email), strings.TrimSpace(excludeID)).Scan(&exists)
	return exists, err
}

func buildPhoto(
	base64 *string,
	mime *string,
	largura *int,
	altura *int,
	tamanho *int,
	hash *string,
) *entity.PhotoInput {
	if base64 == nil || strings.TrimSpace(*base64) == "" {
		return nil
	}

	return &entity.PhotoInput{
		Base64:       strings.TrimSpace(*base64),
		Mime:         readStringPointer(mime),
		Largura:      readIntPointer(largura),
		Altura:       readIntPointer(altura),
		TamanhoBytes: readIntPointer(tamanho),
		HashSHA256:   readStringPointer(hash),
	}
}

func readPhotoBase64(foto *entity.PhotoInput) *string {
	if foto == nil || strings.TrimSpace(foto.Base64) == "" {
		return nil
	}
	value := strings.TrimSpace(foto.Base64)
	return &value
}

func readPhotoMime(foto *entity.PhotoInput) *string {
	if foto == nil || strings.TrimSpace(foto.Mime) == "" {
		return nil
	}
	value := strings.TrimSpace(foto.Mime)
	return &value
}

func readPhotoWidth(foto *entity.PhotoInput) *int {
	if foto == nil || foto.Largura <= 0 {
		return nil
	}
	value := foto.Largura
	return &value
}

func readPhotoHeight(foto *entity.PhotoInput) *int {
	if foto == nil || foto.Altura <= 0 {
		return nil
	}
	value := foto.Altura
	return &value
}

func readPhotoSize(foto *entity.PhotoInput) *int {
	if foto == nil || foto.TamanhoBytes <= 0 {
		return nil
	}
	value := foto.TamanhoBytes
	return &value
}

func readPhotoHash(foto *entity.PhotoInput) *string {
	if foto == nil || strings.TrimSpace(foto.HashSHA256) == "" {
		return nil
	}
	value := strings.TrimSpace(foto.HashSHA256)
	return &value
}

func readStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func readIntPointer(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
