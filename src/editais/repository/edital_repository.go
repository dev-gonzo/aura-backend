package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sistema-editorial/editora/backend/src/editais/entity"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, input entity.PersistInput) (string, error) {
	const query = `
		INSERT INTO editais (
			capa_base64,
			capa_mime,
			capa_largura,
			capa_altura,
			capa_tamanho_bytes,
			capa_hash_sha256,
			titulo,
			descricao,
			anexo_nome_arquivo,
			anexo_content_type,
			anexo_tamanho_bytes,
			anexo_bucket,
			anexo_key,
			anexo_url,
			taxa_inscricao,
			taxa_publicacao,
			status,
			data_inicio,
			data_fim,
			total_vagas,
			data_prevista_publicacao,
			atualizado_em
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, CURRENT_TIMESTAMP
		)
		RETURNING id
	`

	var id string
	err := r.pool.QueryRow(
		ctx,
		query,
		input.Capa.Base64,
		input.Capa.Mime,
		input.Capa.Largura,
		input.Capa.Altura,
		input.Capa.TamanhoBytes,
		input.Capa.HashSHA256,
		input.Titulo,
		input.Descricao,
		nullableAttachmentName(input.Anexo),
		nullableAttachmentContentType(input.Anexo),
		nullableAttachmentSize(input.Anexo),
		nullableAttachmentBucket(input.Anexo),
		nullableAttachmentKey(input.Anexo),
		nullableAttachmentURL(input.Anexo),
		nullableMoney(input.TaxaInscricao),
		nullableMoney(input.TaxaPublicacao),
		input.Status,
		nullableDate(input.DataInicio),
		nullableDate(input.DataFim),
		input.TotalVagas,
		nullableDate(input.DataPrevistaPublicacao),
	).Scan(&id)

	return id, err
}

func (r *PostgresRepository) Update(ctx context.Context, input entity.PersistInput) error {
	const query = `
		UPDATE editais
		SET
			capa_base64 = $2,
			capa_mime = $3,
			capa_largura = $4,
			capa_altura = $5,
			capa_tamanho_bytes = $6,
			capa_hash_sha256 = $7,
			titulo = $8,
			descricao = $9,
			anexo_nome_arquivo = $10,
			anexo_content_type = $11,
			anexo_tamanho_bytes = $12,
			anexo_bucket = $13,
			anexo_key = $14,
			anexo_url = $15,
			taxa_inscricao = $16,
			taxa_publicacao = $17,
			status = $18,
			data_inicio = $19,
			data_fim = $20,
			total_vagas = $21,
			data_prevista_publicacao = $22,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	commandTag, err := r.pool.Exec(
		ctx,
		query,
		input.ID,
		input.Capa.Base64,
		input.Capa.Mime,
		input.Capa.Largura,
		input.Capa.Altura,
		input.Capa.TamanhoBytes,
		input.Capa.HashSHA256,
		input.Titulo,
		input.Descricao,
		nullableAttachmentName(input.Anexo),
		nullableAttachmentContentType(input.Anexo),
		nullableAttachmentSize(input.Anexo),
		nullableAttachmentBucket(input.Anexo),
		nullableAttachmentKey(input.Anexo),
		nullableAttachmentURL(input.Anexo),
		nullableMoney(input.TaxaInscricao),
		nullableMoney(input.TaxaPublicacao),
		input.Status,
		nullableDate(input.DataInicio),
		nullableDate(input.DataFim),
		input.TotalVagas,
		nullableDate(input.DataPrevistaPublicacao),
	)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error) {
	sql := `
		SELECT
			id,
			titulo,
			descricao,
			status,
			COALESCE(to_char(data_inicio, 'YYYY-MM-DD'), ''),
			COALESCE(to_char(data_fim, 'YYYY-MM-DD'), ''),
			total_vagas,
			COALESCE(to_char(data_prevista_publicacao, 'YYYY-MM-DD'), ''),
			(capa_base64 IS NOT NULL AND capa_base64 <> ''),
			(anexo_key IS NOT NULL AND anexo_key <> ''),
			COALESCE(anexo_nome_arquivo, ''),
			to_char(criado_em, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
			to_char(atualizado_em, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM editais
		WHERE 1 = 1
	`

	args := make([]any, 0, 2)
	if search := strings.TrimSpace(query.Search); search != "" {
		sql += ` AND (titulo ILIKE $1 OR descricao ILIKE $1) `
		args = append(args, "%"+search+"%")
	}

	if status := strings.TrimSpace(strings.ToUpper(query.Status)); status != "" {
		placeholder := len(args) + 1
		sql += ` AND status = $` + intToString(placeholder)
		args = append(args, status)
	}

	sql += ` ORDER BY criado_em DESC, titulo ASC `

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.ListItem, 0)
	for rows.Next() {
		item := entity.ListItem{}
		var totalVagas *int
		if err := rows.Scan(
			&item.ID,
			&item.Titulo,
			&item.Descricao,
			&item.Status,
			&item.DataInicio,
			&item.DataFim,
			&totalVagas,
			&item.DataPrevistaPublicacao,
			&item.TemCapa,
			&item.TemAnexo,
			&item.AnexoNomeArquivo,
			&item.CriadoEm,
			&item.AtualizadoEm,
		); err != nil {
			return nil, err
		}

		item.TotalVagas = totalVagas
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (entity.DetailResponse, error) {
	const query = `
		SELECT
			id,
			capa_base64,
			capa_mime,
			COALESCE(capa_largura, 0),
			COALESCE(capa_altura, 0),
			COALESCE(capa_tamanho_bytes, 0),
			COALESCE(capa_hash_sha256, ''),
			titulo,
			descricao,
			COALESCE(anexo_nome_arquivo, ''),
			COALESCE(anexo_content_type, ''),
			COALESCE(anexo_tamanho_bytes, 0),
			COALESCE(anexo_bucket, ''),
			COALESCE(anexo_key, ''),
			COALESCE(anexo_url, ''),
			taxa_inscricao,
			taxa_publicacao,
			status,
			COALESCE(to_char(data_inicio, 'YYYY-MM-DD'), ''),
			COALESCE(to_char(data_fim, 'YYYY-MM-DD'), ''),
			total_vagas,
			COALESCE(to_char(data_prevista_publicacao, 'YYYY-MM-DD'), ''),
			to_char(criado_em, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
			to_char(atualizado_em, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM editais
		WHERE id = $1
	`

	var result entity.DetailResponse
	var totalVagas *int
	var anexoNomeArquivo string
	var anexoContentType string
	var anexoTamanhoBytes int64
	var anexoBucket string
	var anexoKey string
	var anexoURL string
	var taxaInscricao *float64
	var taxaPublicacao *float64

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.Capa.Base64,
		&result.Capa.Mime,
		&result.Capa.Largura,
		&result.Capa.Altura,
		&result.Capa.TamanhoBytes,
		&result.Capa.HashSHA256,
		&result.Titulo,
		&result.Descricao,
		&anexoNomeArquivo,
		&anexoContentType,
		&anexoTamanhoBytes,
		&anexoBucket,
		&anexoKey,
		&anexoURL,
		&taxaInscricao,
		&taxaPublicacao,
		&result.Status,
		&result.DataInicio,
		&result.DataFim,
		&totalVagas,
		&result.DataPrevistaPublicacao,
		&result.CriadoEm,
		&result.AtualizadoEm,
	)
	if err != nil {
		return entity.DetailResponse{}, err
	}

	result.TotalVagas = totalVagas
	result.TaxaInscricao = taxaInscricao
	result.TaxaPublicacao = taxaPublicacao
	if anexoKey != "" {
		result.Anexo = &entity.AnexoInput{
			NomeArquivo:  anexoNomeArquivo,
			ContentType:  anexoContentType,
			TamanhoBytes: anexoTamanhoBytes,
			Bucket:       anexoBucket,
			Key:          anexoKey,
			URL:          anexoURL,
		}
	}

	return result, nil
}

func nullableDate(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}

	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*value))
	if err != nil {
		return nil
	}

	return parsed.Format("2006-01-02")
}

func nullableAttachmentName(anexo *entity.AnexoInput) any {
	if anexo == nil || strings.TrimSpace(anexo.NomeArquivo) == "" {
		return nil
	}

	return strings.TrimSpace(anexo.NomeArquivo)
}

func nullableAttachmentContentType(anexo *entity.AnexoInput) any {
	if anexo == nil || strings.TrimSpace(anexo.ContentType) == "" {
		return nil
	}

	return strings.TrimSpace(anexo.ContentType)
}

func nullableAttachmentSize(anexo *entity.AnexoInput) any {
	if anexo == nil || anexo.TamanhoBytes <= 0 {
		return nil
	}

	return anexo.TamanhoBytes
}

func nullableAttachmentBucket(anexo *entity.AnexoInput) any {
	if anexo == nil || strings.TrimSpace(anexo.Bucket) == "" {
		return nil
	}

	return strings.TrimSpace(anexo.Bucket)
}

func nullableAttachmentKey(anexo *entity.AnexoInput) any {
	if anexo == nil || strings.TrimSpace(anexo.Key) == "" {
		return nil
	}

	return strings.TrimSpace(anexo.Key)
}

func nullableAttachmentURL(anexo *entity.AnexoInput) any {
	if anexo == nil || strings.TrimSpace(anexo.URL) == "" {
		return nil
	}

	return strings.TrimSpace(anexo.URL)
}

func nullableMoney(value *float64) any {
	if value == nil {
		return nil
	}

	return *value
}

func intToString(value int) string {
	return fmt.Sprintf("%d", value)
}
