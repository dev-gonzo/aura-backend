package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"sistema-editorial/editora/backend/src/auth/entity"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindByLogin(ctx context.Context, login string) (entity.LoginUserRecord, error) {
	const query = `
		SELECT
			u.id,
			u.cpf,
			u.email,
			u.nome_completo,
			u.senha_hash,
			COALESCE(array_agg(DISTINCT p.codigo) FILTER (WHERE p.codigo IS NOT NULL), '{}') AS papeis,
			u.precisa_trocar_senha,
			u.ativo,
			COALESCE(u.status_acesso, CASE WHEN u.ativo THEN 'ATIVO' ELSE 'BLOQUEADO' END),
			COALESCE(u.cliente_ativo, u.ativo)
		FROM usuarios u
		LEFT JOIN usuarios_papeis up ON up.usuario_id = u.id
		LEFT JOIN papeis p ON p.id = up.papel_id
		WHERE u.cpf = $1 OR lower(u.email) = lower($2)
		GROUP BY u.id
	`

	var record entity.LoginUserRecord
	err := r.pool.QueryRow(ctx, query, login, login).Scan(
		&record.ID,
		&record.CPF,
		&record.Email,
		&record.NomeCompleto,
		&record.SenhaHash,
		&record.Papeis,
		&record.PrecisaTrocarSenha,
		&record.Ativo,
		&record.StatusAcesso,
		&record.ClienteAtivo,
	)

	return record, err
}

func (r *PostgresRepository) FindByID(ctx context.Context, userID string) (entity.AuthenticatedUser, error) {
	const query = `
		SELECT
			u.id,
			u.cpf,
			u.email,
			u.nome_completo,
			COALESCE(array_agg(DISTINCT p.codigo) FILTER (WHERE p.codigo IS NOT NULL), '{}') AS papeis,
			u.precisa_trocar_senha,
			COALESCE(u.status_acesso, CASE WHEN u.ativo THEN 'ATIVO' ELSE 'BLOQUEADO' END),
			COALESCE(u.cliente_ativo, u.ativo)
		FROM usuarios u
		LEFT JOIN usuarios_papeis up ON up.usuario_id = u.id
		LEFT JOIN papeis p ON p.id = up.papel_id
		WHERE u.id = $1
		GROUP BY u.id
	`

	var currentUser entity.AuthenticatedUser
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&currentUser.ID,
		&currentUser.CPF,
		&currentUser.Email,
		&currentUser.NomeCompleto,
		&currentUser.Papeis,
		&currentUser.PrecisaTrocarSenha,
		&currentUser.StatusAcesso,
		&currentUser.ClienteAtivo,
	)

	return currentUser, err
}

func (r *PostgresRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	const query = `
		UPDATE usuarios
		SET senha_hash = $2,
			precisa_trocar_senha = FALSE,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID, passwordHash)
	return err
}
