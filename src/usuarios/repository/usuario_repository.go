package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sistema-editorial/editora/backend/src/usuarios/entity"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ExistsByCPFOrEmail(ctx context.Context, cpf string, email string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM usuarios
			WHERE cpf = $1 OR lower(email) = lower($2)
		)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, cpf, email).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PostgresRepository) ExistsEmailByDifferentID(ctx context.Context, id string, email string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM usuarios
			WHERE id <> $1
			  AND lower(email) = lower($2)
		)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, id, email).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input entity.PersistInput) (entity.CreateResponse, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return entity.CreateResponse{}, err
	}
	defer tx.Rollback(ctx)

	const insertUserQuery = `
		INSERT INTO usuarios (
			cpf,
			email,
			nome_completo,
			foto_base64,
			foto_mime,
			foto_largura,
			foto_altura,
			foto_tamanho_bytes,
			foto_hash_sha256,
			descricao,
			pseudonimo,
			whatsapp,
			data_nascimento,
			nacionalidade,
			senha_hash,
			precisa_trocar_senha,
			origem_cadastro,
			ativo,
			status_acesso,
			cliente_ativo
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
		RETURNING id
	`

	var userID string
	err = tx.QueryRow(
		ctx,
		insertUserQuery,
		input.CPF,
		strings.ToLower(input.Email),
		input.NomeCompleto,
		input.Foto.Base64,
		input.Foto.Mime,
		input.Foto.Largura,
		input.Foto.Altura,
		input.Foto.TamanhoBytes,
		input.Foto.HashSHA256,
		nullableString(input.Descricao),
		nullableString(input.Pseudonimo),
		input.WhatsApp,
		input.DataNascimento,
		nullableString(input.Nacionalidade),
		input.SenhaHash,
		input.PrecisaTrocarSenha,
		input.OrigemCadastro,
		input.ClienteAtivo,
		input.StatusAcesso,
		input.ClienteAtivo,
	).Scan(&userID)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	roleIDs, err := r.getRoleIDs(ctx, tx, input.Papeis)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	for roleCode, roleID := range roleIDs {
		const insertRoleQuery = `
			INSERT INTO usuarios_papeis (usuario_id, papel_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`

		if _, err := tx.Exec(ctx, insertRoleQuery, userID, roleID); err != nil {
			return entity.CreateResponse{}, fmt.Errorf("erro ao vincular papel %s: %w", roleCode, err)
		}
	}

	if input.EnderecoPrincipal != nil {
		if err := r.insertPrimaryAddress(ctx, tx, userID, *input.EnderecoPrincipal); err != nil {
			return entity.CreateResponse{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return entity.CreateResponse{}, err
	}

	return entity.CreateResponse{
		ID:                 userID,
		CPF:                input.CPF,
		Email:              strings.ToLower(input.Email),
		NomeCompleto:       input.NomeCompleto,
		Pseudonimo:         input.Pseudonimo,
		Papeis:             input.Papeis,
		PrecisaTrocarSenha: input.PrecisaTrocarSenha,
	}, nil
}

func (r *PostgresRepository) Update(ctx context.Context, input entity.PersistInput) (entity.CreateResponse, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return entity.CreateResponse{}, err
	}
	defer tx.Rollback(ctx)

	updateUserQuery := `
		UPDATE usuarios
		SET
			email = $2,
			nome_completo = $3,
			foto_base64 = $4,
			foto_mime = $5,
			foto_largura = $6,
			foto_altura = $7,
			foto_tamanho_bytes = $8,
			foto_hash_sha256 = $9,
			descricao = $10,
			pseudonimo = $11,
			whatsapp = $12,
			data_nascimento = $13,
			nacionalidade = $14,
			precisa_trocar_senha = $15,
			ativo = $16,
			status_acesso = $17,
			cliente_ativo = $18,
			atualizado_em = CURRENT_TIMESTAMP
	`

	args := []any{
		input.ID,
		strings.ToLower(input.Email),
		input.NomeCompleto,
		input.Foto.Base64,
		input.Foto.Mime,
		input.Foto.Largura,
		input.Foto.Altura,
		input.Foto.TamanhoBytes,
		input.Foto.HashSHA256,
		nullableString(input.Descricao),
		nullableString(input.Pseudonimo),
		input.WhatsApp,
		input.DataNascimento,
		nullableString(input.Nacionalidade),
		input.PrecisaTrocarSenha,
		input.ClienteAtivo,
		input.StatusAcesso,
		input.ClienteAtivo,
	}

	if input.UpdatePassword {
		updateUserQuery += `,
			senha_hash = $19
		`
		args = append(args, input.SenhaHash)
	}

	updateUserQuery += `
		WHERE id = $1
	`

	commandTag, err := tx.Exec(ctx, updateUserQuery, args...)
	if err != nil {
		return entity.CreateResponse{}, err
	}
	if commandTag.RowsAffected() == 0 {
		return entity.CreateResponse{}, pgx.ErrNoRows
	}

	if _, err := tx.Exec(ctx, `DELETE FROM usuarios_papeis WHERE usuario_id = $1`, input.ID); err != nil {
		return entity.CreateResponse{}, err
	}

	roleIDs, err := r.getRoleIDs(ctx, tx, input.Papeis)
	if err != nil {
		return entity.CreateResponse{}, err
	}

	for roleCode, roleID := range roleIDs {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO usuarios_papeis (usuario_id, papel_id) VALUES ($1, $2)`,
			input.ID,
			roleID,
		); err != nil {
			return entity.CreateResponse{}, fmt.Errorf("erro ao vincular papel %s: %w", roleCode, err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM usuarios_enderecos WHERE usuario_id = $1 AND principal = TRUE`, input.ID); err != nil {
		return entity.CreateResponse{}, err
	}

	if input.EnderecoPrincipal != nil {
		if err := r.insertPrimaryAddress(ctx, tx, input.ID, *input.EnderecoPrincipal); err != nil {
			return entity.CreateResponse{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return entity.CreateResponse{}, err
	}

	return entity.CreateResponse{
		ID:                 input.ID,
		CPF:                input.CPF,
		Email:              strings.ToLower(input.Email),
		NomeCompleto:       input.NomeCompleto,
		Pseudonimo:         input.Pseudonimo,
		Papeis:             input.Papeis,
		PrecisaTrocarSenha: input.PrecisaTrocarSenha,
	}, nil
}

func (r *PostgresRepository) UpdateAccessStatus(ctx context.Context, id string, status string, clienteAtivo bool) error {
	const query = `
		UPDATE usuarios
		SET
			ativo = $2,
			status_acesso = $3,
			cliente_ativo = $2,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	commandTag, err := r.pool.Exec(ctx, query, id, clienteAtivo, status)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) ResetPassword(ctx context.Context, id string, passwordHash string) error {
	const query = `
		UPDATE usuarios
		SET
			senha_hash = $2,
			precisa_trocar_senha = TRUE,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	commandTag, err := r.pool.Exec(ctx, query, id, passwordHash)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (entity.ListItem, error) {
	items, _, err := r.listUsers(ctx, filterOptions{ID: id, Page: 1, PageSize: 1})
	if err != nil {
		return entity.ListItem{}, err
	}

	if len(items) == 0 {
		return entity.ListItem{}, pgx.ErrNoRows
	}

	return items[0], nil
}

func (r *PostgresRepository) List(ctx context.Context, query entity.ListQuery) (entity.ListResponse, error) {
	items, total, err := r.listUsers(ctx, filterOptions{
		Search:   query.Search,
		Role:     query.Role,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return entity.ListResponse{}, err
	}

	totalPages := 0
	if query.PageSize > 0 && total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}

	return entity.ListResponse{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (r *PostgresRepository) insertPrimaryAddress(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	address entity.EnderecoInput,
) error {
	const insertAddressQuery = `
		INSERT INTO enderecos (
			cep,
			logradouro,
			numero,
			complemento,
			bairro,
			cidade,
			uf,
			pais
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var addressID string
	err := tx.QueryRow(
		ctx,
		insertAddressQuery,
		address.CEP,
		address.Logradouro,
		address.Numero,
		nullableString(address.Complemento),
		address.Bairro,
		address.Cidade,
		address.UF,
		coalesce(address.Pais, "BRASIL"),
	).Scan(&addressID)
	if err != nil {
		return err
	}

	const linkQuery = `
		INSERT INTO usuarios_enderecos (usuario_id, endereco_id, principal)
		VALUES ($1, $2, TRUE)
	`

	_, err = tx.Exec(ctx, linkQuery, userID, addressID)
	return err
}

func (r *PostgresRepository) getRoleIDs(ctx context.Context, tx pgx.Tx, roleCodes []string) (map[string]string, error) {
	rows, err := tx.Query(ctx, `SELECT codigo, id FROM papeis WHERE codigo = ANY($1)`, roleCodes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var code string
		var id string
		if err := rows.Scan(&code, &id); err != nil {
			return nil, err
		}
		result[code] = id
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) != len(roleCodes) {
		return nil, fmt.Errorf("nem todos os papeis informados existem")
	}

	return result, nil
}

type filterOptions struct {
	ID       string
	Search   string
	Role     string
	Page     int
	PageSize int
}

func (r *PostgresRepository) listUsers(
	ctx context.Context,
	options filterOptions,
) ([]entity.ListItem, int, error) {
	search := strings.TrimSpace(options.Search)
	searchLike := "%" + strings.ToLower(search) + "%"
	searchDigits := digitsOnly(search)
	searchDigitsLike := "%" + searchDigits + "%"
	role := strings.ToUpper(strings.TrimSpace(options.Role))

	const countQuery = `
		SELECT COUNT(*)
		FROM usuarios u
		WHERE
			($1 = '' OR u.id::text = $1)
			AND (
				$2 = ''
				OR lower(u.nome_completo) LIKE $3
				OR lower(COALESCE(u.pseudonimo, '')) LIKE $3
				OR ($4 <> '' AND (u.cpf LIKE $5 OR regexp_replace(COALESCE(u.whatsapp, ''), '[^0-9]', '', 'g') LIKE $5))
			)
			AND (
				$6 = ''
				OR EXISTS (
					SELECT 1
					FROM usuarios_papeis up
					INNER JOIN papeis p ON p.id = up.papel_id
					WHERE up.usuario_id = u.id
					  AND p.codigo = $6
				)
			)
	`

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, options.ID, search, searchLike, searchDigits, searchDigitsLike, role).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []entity.ListItem{}, 0, nil
	}

	limit := options.PageSize
	offset := (options.Page - 1) * options.PageSize
	if options.ID != "" {
		limit = 1
		offset = 0
	}

	const listQuery = `
		WITH filtered_users AS (
			SELECT u.id
			FROM usuarios u
			WHERE
				($1 = '' OR u.id::text = $1)
				AND (
					$2 = ''
					OR lower(u.nome_completo) LIKE $3
					OR lower(COALESCE(u.pseudonimo, '')) LIKE $3
					OR ($4 <> '' AND (u.cpf LIKE $5 OR regexp_replace(COALESCE(u.whatsapp, ''), '[^0-9]', '', 'g') LIKE $5))
				)
				AND (
					$6 = ''
					OR EXISTS (
						SELECT 1
						FROM usuarios_papeis up
						INNER JOIN papeis p ON p.id = up.papel_id
						WHERE up.usuario_id = u.id
						  AND p.codigo = $6
					)
				)
			ORDER BY u.nome_completo ASC
			LIMIT $7 OFFSET $8
		)
		SELECT
			u.id::text,
			u.cpf,
			u.email,
			u.nome_completo,
			COALESCE(u.pseudonimo, ''),
			u.whatsapp,
			COALESCE(u.origem_cadastro, 'EDITORA'),
			COALESCE(u.status_acesso, CASE WHEN u.ativo THEN 'ATIVO' ELSE 'BLOQUEADO' END),
			COALESCE(u.cliente_ativo, u.ativo),
			COALESCE(u.descricao, ''),
			TO_CHAR(u.data_nascimento, 'YYYY-MM-DD'),
			COALESCE(u.nacionalidade, ''),
			COALESCE(u.foto_base64, ''),
			COALESCE(u.foto_mime, ''),
			COALESCE(u.foto_largura, 0),
			COALESCE(u.foto_altura, 0),
			COALESCE(u.foto_tamanho_bytes, 0),
			COALESCE(u.foto_hash_sha256, ''),
			COALESCE(array_agg(DISTINCT p.codigo) FILTER (WHERE p.codigo IS NOT NULL), '{}') AS papeis,
			COALESCE(e.cep, ''),
			COALESCE(e.logradouro, ''),
			COALESCE(e.numero, ''),
			COALESCE(e.complemento, ''),
			COALESCE(e.bairro, ''),
			COALESCE(e.cidade, ''),
			COALESCE(e.uf, ''),
			COALESCE(e.pais, '')
		FROM filtered_users fu
		INNER JOIN usuarios u ON u.id = fu.id
		LEFT JOIN usuarios_papeis up ON up.usuario_id = u.id
		LEFT JOIN papeis p ON p.id = up.papel_id
		LEFT JOIN usuarios_enderecos ue ON ue.usuario_id = u.id AND ue.principal = TRUE
		LEFT JOIN enderecos e ON e.id = ue.endereco_id
		GROUP BY
			u.id, u.cpf, u.email, u.nome_completo, u.pseudonimo, u.whatsapp, u.origem_cadastro,
			u.status_acesso, u.cliente_ativo, u.descricao, u.data_nascimento, u.nacionalidade, u.foto_base64, u.foto_mime, u.foto_largura,
			u.foto_altura, u.foto_tamanho_bytes, u.foto_hash_sha256,
			e.cep, e.logradouro, e.numero, e.complemento, e.bairro, e.cidade, e.uf, e.pais
		ORDER BY u.nome_completo ASC
	`

	rows, err := r.pool.Query(
		ctx,
		listQuery,
		options.ID,
		search,
		searchLike,
		searchDigits,
		searchDigitsLike,
		role,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]entity.ListItem, 0, limit)
	for rows.Next() {
		var item entity.ListItem
		var statusAcesso string
		var clienteAtivo bool
		var nacionalidade string
		var fotoBase64 string
		var fotoMime string
		var fotoLargura int
		var fotoAltura int
		var fotoTamanhoBytes int
		var fotoHash string
		var papeis []string
		var cep, logradouro, numero, complemento, bairro, cidade, uf, pais string

		if err := rows.Scan(
			&item.ID,
			&item.CPF,
			&item.Email,
			&item.NomeCompleto,
			&item.Pseudonimo,
			&item.WhatsApp,
			&item.OrigemCadastro,
			&statusAcesso,
			&clienteAtivo,
			&item.Descricao,
			&item.DataNascimento,
			&nacionalidade,
			&fotoBase64,
			&fotoMime,
			&fotoLargura,
			&fotoAltura,
			&fotoTamanhoBytes,
			&fotoHash,
			&papeis,
			&cep,
			&logradouro,
			&numero,
			&complemento,
			&bairro,
			&cidade,
			&uf,
			&pais,
		); err != nil {
			return nil, 0, err
		}

		item.Papeis = papeis
		item.StatusCodigo = normalizeAccessStatus(statusAcesso)
		item.ClienteAtivo = clienteAtivo
		item.Status = buildStatusLabel(item.StatusCodigo, item.ClienteAtivo)
		item.Nacionalidade = strings.TrimSpace(nacionalidade)

		if fotoBase64 != "" {
			item.Foto = &entity.FotoInput{
				Base64:       fotoBase64,
				Mime:         coalesce(fotoMime, "image/webp"),
				Largura:      fotoLargura,
				Altura:       fotoAltura,
				TamanhoBytes: fotoTamanhoBytes,
				HashSHA256:   fotoHash,
			}
		}

		if cep != "" || logradouro != "" || numero != "" || bairro != "" || cidade != "" || uf != "" {
			item.EnderecoPrincipal = &entity.EnderecoInput{
				CEP:         cep,
				Logradouro:  logradouro,
				Numero:      numero,
				Complemento: complemento,
				Bairro:      bairro,
				Cidade:      cidade,
				UF:          uf,
				Pais:        coalesce(pais, "BRASIL"),
			}
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return strings.TrimSpace(value)
}

func coalesce(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return strings.TrimSpace(value)
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

func buildStatusLabel(status string, clienteAtivo bool) string {
	switch normalizeAccessStatus(status) {
	case entity.StatusAcessoBloqueado:
		return "Bloqueado"
	case entity.StatusAcessoPendenteAprovacao:
		if clienteAtivo {
			return "Cliente ativo"
		}
		return "Pendente"
	default:
		return "Ativo"
	}
}

func digitsOnly(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}

	return builder.String()
}
