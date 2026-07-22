package repository

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"sistema-editorial/editora/backend/src/livros/entity"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) IsCatalogIdentifierAvailable(
	ctx context.Context,
	normalizedISBN string,
	codigoBarra string,
	excludeID string,
) (bool, error) {
	const query = `
		SELECT NOT EXISTS (
			SELECT 1
			FROM livros
			WHERE (
				($1 <> '' AND regexp_replace(coalesce(isbn, ''), '[^0-9]', '', 'g') = $1)
				OR ($2 <> '' AND codigo_barra = $2)
			)
			AND (NULLIF($3, '')::uuid IS NULL OR id <> NULLIF($3, '')::uuid)
		)
	`

	var available bool
	err := r.pool.QueryRow(ctx, query, normalizedISBN, codigoBarra, strings.TrimSpace(excludeID)).Scan(&available)
	if err != nil {
		return false, err
	}

	return available, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input entity.PersistInput) (string, error) {
	const query = `
		INSERT INTO livros (
			autor_id,
			titulo,
			subtitulo,
			sinopse,
			isbn,
			codigo_barra,
			status,
			formato,
			possui_formato_fisico,
			possui_formato_digital,
			edicao,
			idioma,
			numero_paginas,
			genero,
			preco_venda,
			preco_venda_fisico,
			preco_venda_digital,
			canal_venda_digital,
			url_compra_digital,
			custo_impressao,
			venda_infinita,
			controlar_estoque,
			estoque_disponivel,
			estoque_minimo,
			peso_gramas,
			largura_cm,
			altura_cm,
			profundidade_cm,
			tipo_capa,
			possui_box,
			detalhes_edicao,
			data_publicacao_prevista,
			publicado_em,
			capa_base64,
			capa_mime,
			capa_largura,
			capa_altura,
			capa_tamanho_bytes,
			capa_hash_sha256,
			ativo
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
			$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,
			$31,$32,$33,$34,$35,$36,$37,$38,$39,$40
		)
		RETURNING id
	`

	var id string
	err := r.pool.QueryRow(
		ctx,
		query,
		input.AutorID,
		input.Titulo,
		input.Subtitulo,
		input.Sinopse,
		input.ISBN,
		input.CodigoBarra,
		input.Status,
		input.Formato,
		input.PossuiFormatoFisico,
		input.PossuiFormatoDigital,
		input.Edicao,
		input.Idioma,
		input.NumeroPaginas,
		input.Genero,
		input.PrecoVenda,
		input.PrecoVendaFisico,
		input.PrecoVendaDigital,
		input.CanalVendaDigital,
		input.URLCompraDigital,
		input.CustoImpressao,
		input.VendaInfinita,
		input.ControlarEstoque,
		input.EstoqueDisponivel,
		input.EstoqueMinimo,
		input.PesoGramas,
		input.LarguraCm,
		input.AlturaCm,
		input.ProfundidadeCm,
		input.TipoCapa,
		input.PossuiBox,
		input.DetalhesEdicao,
		input.DataPublicacaoPrevista,
		input.DataPublicacao,
		readCoverBase64(input.Capa),
		readCoverMime(input.Capa),
		readCoverWidth(input.Capa),
		readCoverHeight(input.Capa),
		readCoverSize(input.Capa),
		readCoverHash(input.Capa),
		input.Ativo,
	).Scan(&id)
	return id, err
}

func (r *PostgresRepository) Update(ctx context.Context, input entity.PersistInput) error {
	const query = `
		UPDATE livros
		SET
			autor_id = $2,
			titulo = $3,
			subtitulo = $4,
			sinopse = $5,
			isbn = $6,
			codigo_barra = $7,
			status = $8,
			formato = $9,
			possui_formato_fisico = $10,
			possui_formato_digital = $11,
			edicao = $12,
			idioma = $13,
			numero_paginas = $14,
			genero = $15,
			preco_venda = $16,
			preco_venda_fisico = $17,
			preco_venda_digital = $18,
			canal_venda_digital = $19,
			url_compra_digital = $20,
			custo_impressao = $21,
			venda_infinita = $22,
			controlar_estoque = $23,
			estoque_disponivel = $24,
			estoque_minimo = $25,
			peso_gramas = $26,
			largura_cm = $27,
			altura_cm = $28,
			profundidade_cm = $29,
			tipo_capa = $30,
			possui_box = $31,
			detalhes_edicao = $32,
			data_publicacao_prevista = $33,
			publicado_em = $34,
			capa_base64 = $35,
			capa_mime = $36,
			capa_largura = $37,
			capa_altura = $38,
			capa_tamanho_bytes = $39,
			capa_hash_sha256 = $40,
			ativo = $41,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		input.ID,
		input.AutorID,
		input.Titulo,
		input.Subtitulo,
		input.Sinopse,
		input.ISBN,
		input.CodigoBarra,
		input.Status,
		input.Formato,
		input.PossuiFormatoFisico,
		input.PossuiFormatoDigital,
		input.Edicao,
		input.Idioma,
		input.NumeroPaginas,
		input.Genero,
		input.PrecoVenda,
		input.PrecoVendaFisico,
		input.PrecoVendaDigital,
		input.CanalVendaDigital,
		input.URLCompraDigital,
		input.CustoImpressao,
		input.VendaInfinita,
		input.ControlarEstoque,
		input.EstoqueDisponivel,
		input.EstoqueMinimo,
		input.PesoGramas,
		input.LarguraCm,
		input.AlturaCm,
		input.ProfundidadeCm,
		input.TipoCapa,
		input.PossuiBox,
		input.DetalhesEdicao,
		input.DataPublicacaoPrevista,
		input.DataPublicacao,
		readCoverBase64(input.Capa),
		readCoverMime(input.Capa),
		readCoverWidth(input.Capa),
		readCoverHeight(input.Capa),
		readCoverSize(input.Capa),
		readCoverHash(input.Capa),
		input.Ativo,
	)
	return err
}

func (r *PostgresRepository) List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error) {
	const sql = `
		SELECT
			l.id,
			l.autor_id,
			a.nome_completo,
			l.titulo,
			l.subtitulo,
			l.isbn,
			l.codigo_barra,
			l.status,
			l.formato,
			l.possui_formato_fisico,
			l.possui_formato_digital,
			l.genero,
			l.preco_venda,
			l.preco_venda_fisico,
			l.preco_venda_digital,
			l.canal_venda_digital,
			l.url_compra_digital,
			l.venda_infinita,
			l.controlar_estoque,
			l.estoque_disponivel,
			l.estoque_minimo,
			l.ativo,
			l.capa_base64,
			l.capa_mime,
			l.capa_largura,
			l.capa_altura,
			l.capa_tamanho_bytes,
			l.capa_hash_sha256,
			to_char(l.publicado_em, 'YYYY-MM-DD'),
			to_char(l.data_publicacao_prevista, 'YYYY-MM-DD'),
			to_char(l.criado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(l.atualizado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM livros l
		INNER JOIN autores a ON a.id = l.autor_id
		WHERE
			($1 = '' OR l.status = $1)
			AND (NULLIF($2, '')::uuid IS NULL OR l.autor_id = NULLIF($2, '')::uuid)
			AND (
				$3 = ''
				OR l.titulo ILIKE '%' || $3 || '%'
				OR COALESCE(l.subtitulo, '') ILIKE '%' || $3 || '%'
				OR a.nome_completo ILIKE '%' || $3 || '%'
			)
		ORDER BY l.titulo ASC
	`

	rows, err := r.pool.Query(
		ctx,
		sql,
		strings.TrimSpace(query.Status),
		strings.TrimSpace(query.AutorID),
		strings.TrimSpace(query.Search),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.ListItem, 0)
	for rows.Next() {
		var item entity.ListItem
		var capaBase64, capaMime, capaHash, dataPublicacao, dataPrevista *string
		var capaLargura, capaAltura, capaTamanho *int
		if err := rows.Scan(
			&item.ID,
			&item.AutorID,
			&item.AutorNome,
			&item.Titulo,
			&item.Subtitulo,
			&item.ISBN,
			&item.CodigoBarra,
			&item.Status,
			&item.Formato,
			&item.PossuiFormatoFisico,
			&item.PossuiFormatoDigital,
			&item.Genero,
			&item.PrecoVenda,
			&item.PrecoVendaFisico,
			&item.PrecoVendaDigital,
			&item.CanalVendaDigital,
			&item.URLCompraDigital,
			&item.VendaInfinita,
			&item.ControlarEstoque,
			&item.EstoqueDisponivel,
			&item.EstoqueMinimo,
			&item.Ativo,
			&capaBase64,
			&capaMime,
			&capaLargura,
			&capaAltura,
			&capaTamanho,
			&capaHash,
			&dataPublicacao,
			&dataPrevista,
			&item.CriadoEm,
			&item.AtualizadoEm,
		); err != nil {
			return nil, err
		}
		item.Capa = buildCover(capaBase64, capaMime, capaLargura, capaAltura, capaTamanho, capaHash)
		item.PossuiCapa = item.Capa != nil
		item.DataPublicacao = dataPublicacao
		item.DataPublicacaoPrevista = dataPrevista
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (entity.DetailResponse, error) {
	const sql = `
		SELECT
			l.id,
			l.autor_id,
			a.nome_completo,
			l.titulo,
			l.subtitulo,
			l.sinopse,
			l.isbn,
			l.codigo_barra,
			l.status,
			l.formato,
			l.possui_formato_fisico,
			l.possui_formato_digital,
			l.edicao,
			l.idioma,
			l.numero_paginas,
			l.genero,
			l.preco_venda,
			l.preco_venda_fisico,
			l.preco_venda_digital,
			l.canal_venda_digital,
			l.url_compra_digital,
			l.custo_impressao,
			l.venda_infinita,
			l.controlar_estoque,
			l.estoque_disponivel,
			l.estoque_reservado,
			l.estoque_minimo,
			l.peso_gramas,
			l.largura_cm,
			l.altura_cm,
			l.profundidade_cm,
			l.tipo_capa,
			l.possui_box,
			l.detalhes_edicao,
			to_char(l.publicado_em, 'YYYY-MM-DD'),
			to_char(l.data_publicacao_prevista, 'YYYY-MM-DD'),
			l.capa_base64,
			l.capa_mime,
			l.capa_largura,
			l.capa_altura,
			l.capa_tamanho_bytes,
			l.capa_hash_sha256,
			l.ativo,
			to_char(l.criado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(l.atualizado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM livros l
		INNER JOIN autores a ON a.id = l.autor_id
		WHERE l.id = $1
	`

	var item entity.DetailResponse
	var capaBase64, capaMime, capaHash *string
	var capaLargura, capaAltura, capaTamanho *int
	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&item.ID,
		&item.AutorID,
		&item.AutorNome,
		&item.Titulo,
		&item.Subtitulo,
		&item.Sinopse,
		&item.ISBN,
		&item.CodigoBarra,
		&item.Status,
		&item.Formato,
		&item.PossuiFormatoFisico,
		&item.PossuiFormatoDigital,
		&item.Edicao,
		&item.Idioma,
		&item.NumeroPaginas,
		&item.Genero,
		&item.PrecoVenda,
		&item.PrecoVendaFisico,
		&item.PrecoVendaDigital,
		&item.CanalVendaDigital,
		&item.URLCompraDigital,
		&item.CustoImpressao,
		&item.VendaInfinita,
		&item.ControlarEstoque,
		&item.EstoqueDisponivel,
		&item.EstoqueReservado,
		&item.EstoqueMinimo,
		&item.PesoGramas,
		&item.LarguraCm,
		&item.AlturaCm,
		&item.ProfundidadeCm,
		&item.TipoCapa,
		&item.PossuiBox,
		&item.DetalhesEdicao,
		&item.DataPublicacao,
		&item.DataPublicacaoPrevista,
		&capaBase64,
		&capaMime,
		&capaLargura,
		&capaAltura,
		&capaTamanho,
		&capaHash,
		&item.Ativo,
		&item.CriadoEm,
		&item.AtualizadoEm,
	)
	if err != nil {
		return entity.DetailResponse{}, err
	}

	item.Capa = buildCover(capaBase64, capaMime, capaLargura, capaAltura, capaTamanho, capaHash)
	return item, nil
}

func (r *PostgresRepository) RegisterStockMovement(
	ctx context.Context,
	livroID string,
	request entity.StockMovementRequest,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var estoqueAtual int
	var vendaInfinita bool
	if err := tx.QueryRow(
		ctx,
		`SELECT estoque_disponivel, venda_infinita FROM livros WHERE id = $1`,
		livroID,
	).Scan(&estoqueAtual, &vendaInfinita); err != nil {
		return err
	}

	if !vendaInfinita {
		switch request.Tipo {
		case "ENTRADA":
			estoqueAtual += request.Quantidade
		case "SAIDA":
			estoqueAtual -= request.Quantidade
		case "AJUSTE":
			estoqueAtual = request.Quantidade
		}

		if estoqueAtual < 0 {
			estoqueAtual = 0
		}

		if _, err := tx.Exec(
			ctx,
			`UPDATE livros SET estoque_disponivel = $2, atualizado_em = CURRENT_TIMESTAMP WHERE id = $1`,
			livroID,
			estoqueAtual,
		); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO livro_estoque_movimentos (livro_id, tipo, quantidade, motivo, observacao) VALUES ($1, $2, $3, $4, $5)`,
		livroID,
		request.Tipo,
		request.Quantidade,
		request.Motivo,
		strings.TrimSpace(request.Observacao),
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func buildCover(
	base64 *string,
	mime *string,
	largura *int,
	altura *int,
	tamanho *int,
	hash *string,
) *entity.CoverInput {
	if base64 == nil || strings.TrimSpace(*base64) == "" {
		return nil
	}

	return &entity.CoverInput{
		Base64:       strings.TrimSpace(*base64),
		Mime:         readStringPointer(mime),
		Largura:      readIntPointer(largura),
		Altura:       readIntPointer(altura),
		TamanhoBytes: readIntPointer(tamanho),
		HashSHA256:   readStringPointer(hash),
	}
}

func readCoverBase64(capa *entity.CoverInput) *string {
	if capa == nil || strings.TrimSpace(capa.Base64) == "" {
		return nil
	}
	value := strings.TrimSpace(capa.Base64)
	return &value
}

func readCoverMime(capa *entity.CoverInput) *string {
	if capa == nil || strings.TrimSpace(capa.Mime) == "" {
		return nil
	}
	value := strings.TrimSpace(capa.Mime)
	return &value
}

func readCoverWidth(capa *entity.CoverInput) *int {
	if capa == nil || capa.Largura <= 0 {
		return nil
	}
	value := capa.Largura
	return &value
}

func readCoverHeight(capa *entity.CoverInput) *int {
	if capa == nil || capa.Altura <= 0 {
		return nil
	}
	value := capa.Altura
	return &value
}

func readCoverSize(capa *entity.CoverInput) *int {
	if capa == nil || capa.TamanhoBytes <= 0 {
		return nil
	}
	value := capa.TamanhoBytes
	return &value
}

func readCoverHash(capa *entity.CoverInput) *string {
	if capa == nil || strings.TrimSpace(capa.HashSHA256) == "" {
		return nil
	}
	value := strings.TrimSpace(capa.HashSHA256)
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
