package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sistema-editorial/editora/backend/src/pedidos/entity"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) NextCode(ctx context.Context) (string, error) {
	const query = `SELECT COALESCE(MAX(CAST(regexp_replace(codigo, '[^0-9]', '', 'g') AS INTEGER)), 0) + 1 FROM pedidos`
	var next int
	if err := r.pool.QueryRow(ctx, query).Scan(&next); err != nil {
		return "", err
	}
	return fmt.Sprintf("PED-%06d", next), nil
}

func (r *PostgresRepository) LookupLivro(ctx context.Context, id string) (string, string, error) {
	const query = `
		SELECT l.titulo, a.nome_completo
		FROM livros l
		INNER JOIN autores a ON a.id = l.autor_id
		WHERE l.id = $1
	`
	var titulo, autor string
	err := r.pool.QueryRow(ctx, query, id).Scan(&titulo, &autor)
	return titulo, autor, err
}

func (r *PostgresRepository) Create(ctx context.Context, input entity.PersistInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	const pedidoSQL = `
		INSERT INTO pedidos (
			codigo, canal_venda, status, cliente_nome, cliente_email, cliente_whatsapp,
			subtotal, desconto, frete, total, observacao
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id
	`

	var id string
	if err := tx.QueryRow(
		ctx,
		pedidoSQL,
		input.Codigo,
		input.CanalVenda,
		input.Status,
		input.ClienteNome,
		input.ClienteEmail,
		input.ClienteWhatsapp,
		input.Subtotal,
		input.Desconto,
		input.Frete,
		input.Total,
		input.Observacao,
	).Scan(&id); err != nil {
		return "", err
	}

	if err := r.replaceItems(ctx, tx, id, input.Itens); err != nil {
		return "", err
	}
	if err := r.replaceEntrega(ctx, tx, id, input.Entrega); err != nil {
		return "", err
	}

	return id, tx.Commit(ctx)
}

func (r *PostgresRepository) Update(ctx context.Context, input entity.PersistInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const pedidoSQL = `
		UPDATE pedidos
		SET
			canal_venda = $2,
			status = $3,
			cliente_nome = $4,
			cliente_email = $5,
			cliente_whatsapp = $6,
			subtotal = $7,
			desconto = $8,
			frete = $9,
			total = $10,
			observacao = $11,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	if _, err := tx.Exec(
		ctx,
		pedidoSQL,
		input.ID,
		input.CanalVenda,
		input.Status,
		input.ClienteNome,
		input.ClienteEmail,
		input.ClienteWhatsapp,
		input.Subtotal,
		input.Desconto,
		input.Frete,
		input.Total,
		input.Observacao,
	); err != nil {
		return err
	}

	if err := r.replaceItems(ctx, tx, input.ID, input.Itens); err != nil {
		return err
	}
	if err := r.replaceEntrega(ctx, tx, input.ID, input.Entrega); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error) {
	const sql = `
		SELECT
			p.id,
			p.codigo,
			p.canal_venda,
			p.status,
			p.cliente_nome,
			p.cliente_email,
			p.subtotal,
			p.desconto,
			p.frete,
			p.total,
			COALESCE(SUM(pi.quantidade), 0),
			to_char(p.criado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(p.atualizado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM pedidos p
		LEFT JOIN pedido_itens pi ON pi.pedido_id = p.id
		WHERE
			($1 = '' OR p.status = $1)
			AND (
				$2 = ''
				OR p.codigo ILIKE '%' || $2 || '%'
				OR p.cliente_nome ILIKE '%' || $2 || '%'
				OR COALESCE(p.cliente_email, '') ILIKE '%' || $2 || '%'
			)
		GROUP BY p.id
		ORDER BY p.criado_em DESC
	`

	rows, err := r.pool.Query(ctx, sql, strings.TrimSpace(query.Status), strings.TrimSpace(query.Search))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.ListItem, 0)
	for rows.Next() {
		var item entity.ListItem
		if err := rows.Scan(
			&item.ID,
			&item.Codigo,
			&item.CanalVenda,
			&item.Status,
			&item.ClienteNome,
			&item.ClienteEmail,
			&item.Subtotal,
			&item.Desconto,
			&item.Frete,
			&item.Total,
			&item.ItensQuantidade,
			&item.CriadoEm,
			&item.AtualizadoEm,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (entity.DetailResponse, error) {
	const pedidoSQL = `
		SELECT
			id, codigo, canal_venda, status, cliente_nome, cliente_email, cliente_whatsapp,
			subtotal, desconto, frete, total, observacao,
			to_char(criado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(atualizado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM pedidos
		WHERE id = $1
	`

	var item entity.DetailResponse
	if err := r.pool.QueryRow(ctx, pedidoSQL, id).Scan(
		&item.ID,
		&item.Codigo,
		&item.CanalVenda,
		&item.Status,
		&item.ClienteNome,
		&item.ClienteEmail,
		&item.ClienteWhatsapp,
		&item.Subtotal,
		&item.Desconto,
		&item.Frete,
		&item.Total,
		&item.Observacao,
		&item.CriadoEm,
		&item.AtualizadoEm,
	); err != nil {
		return entity.DetailResponse{}, err
	}

	itensRows, err := r.pool.Query(
		ctx,
		`SELECT livro_id, titulo_livro, autor_nome, quantidade, preco_unitario, subtotal FROM pedido_itens WHERE pedido_id = $1 ORDER BY titulo_livro`,
		id,
	)
	if err != nil {
		return entity.DetailResponse{}, err
	}
	defer itensRows.Close()

	item.Itens = make([]entity.PersistItem, 0)
	for itensRows.Next() {
		var persisted entity.PersistItem
		if err := itensRows.Scan(
			&persisted.LivroID,
			&persisted.TituloLivro,
			&persisted.AutorNome,
			&persisted.Quantidade,
			&persisted.PrecoUnitario,
			&persisted.Subtotal,
		); err != nil {
			return entity.DetailResponse{}, err
		}
		item.Itens = append(item.Itens, persisted)
	}

	const entregaSQL = `
		SELECT
			tipo_entrega, status_entrega, transportadora, codigo_rastreio, destinatario_nome, destinatario_documento,
			cep, logradouro, numero, complemento, bairro, cidade, uf,
			to_char(prazo_previsto_em, 'YYYY-MM-DD'),
			to_char(postado_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			to_char(entregue_em AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			observacao
		FROM pedido_entregas
		WHERE pedido_id = $1
	`

	var entrega entity.PersistEntrega
	err = r.pool.QueryRow(ctx, entregaSQL, id).Scan(
		&entrega.TipoEntrega,
		&entrega.StatusEntrega,
		&entrega.Transportadora,
		&entrega.CodigoRastreio,
		&entrega.DestinatarioNome,
		&entrega.DestinatarioDocumento,
		&entrega.CEP,
		&entrega.Logradouro,
		&entrega.Numero,
		&entrega.Complemento,
		&entrega.Bairro,
		&entrega.Cidade,
		&entrega.UF,
		&entrega.PrazoPrevistoEm,
		&entrega.PostadoEm,
		&entrega.EntregueEm,
		&entrega.Observacao,
	)
	if err == nil {
		item.Entrega = &entrega
	}

	return item, nil
}

func (r *PostgresRepository) replaceItems(ctx context.Context, tx pgx.Tx, pedidoID string, itens []entity.PersistItem) error {
	if _, err := tx.Exec(ctx, `DELETE FROM pedido_itens WHERE pedido_id = $1`, pedidoID); err != nil {
		return err
	}
	for _, item := range itens {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO pedido_itens (pedido_id, livro_id, titulo_livro, autor_nome, quantidade, preco_unitario, subtotal) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			pedidoID,
			item.LivroID,
			item.TituloLivro,
			item.AutorNome,
			item.Quantidade,
			item.PrecoUnitario,
			item.Subtotal,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) replaceEntrega(ctx context.Context, tx pgx.Tx, pedidoID string, entrega *entity.PersistEntrega) error {
	if _, err := tx.Exec(ctx, `DELETE FROM pedido_entregas WHERE pedido_id = $1`, pedidoID); err != nil {
		return err
	}
	if entrega == nil {
		return nil
	}
	_, err := tx.Exec(
		ctx,
		`INSERT INTO pedido_entregas (
			pedido_id, tipo_entrega, status_entrega, transportadora, codigo_rastreio, destinatario_nome,
			destinatario_documento, cep, logradouro, numero, complemento, bairro, cidade, uf,
			prazo_previsto_em, postado_em, entregue_em, observacao
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
		)`,
		pedidoID,
		entrega.TipoEntrega,
		entrega.StatusEntrega,
		entrega.Transportadora,
		entrega.CodigoRastreio,
		entrega.DestinatarioNome,
		entrega.DestinatarioDocumento,
		entrega.CEP,
		entrega.Logradouro,
		entrega.Numero,
		entrega.Complemento,
		entrega.Bairro,
		entrega.Cidade,
		entrega.UF,
		entrega.PrazoPrevistoEm,
		entrega.PostadoEm,
		entrega.EntregueEm,
		entrega.Observacao,
	)
	return err
}
