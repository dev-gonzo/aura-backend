package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"sistema-editorial/editora/backend/src/pagamentos/entity"
)

type SettingsRepository struct {
	pool *pgxpool.Pool
}

func NewSettingsRepository(pool *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{pool: pool}
}

func (r *SettingsRepository) Get(ctx context.Context) (entity.SettingsRecord, error) {
	const sql = `
		SELECT
			provider_padrao,
			timeout_segundos,
			coalesce(email_contato, ''),
			mercado_pago_habilitado,
			mercado_pago_sandbox,
			coalesce(mercado_pago_base_url, ''),
			coalesce(mercado_pago_public_key, ''),
			coalesce(mercado_pago_access_token, ''),
			coalesce(mercado_pago_statement_descriptor, ''),
			coalesce(mercado_pago_success_url, ''),
			coalesce(mercado_pago_failure_url, ''),
			coalesce(mercado_pago_pending_url, ''),
			coalesce(mercado_pago_webhook_url, ''),
			mercado_pago_binary_mode,
			mercado_pago_wallet_purchase,
			mercado_pago_installments,
			criado_em,
			atualizado_em
		FROM pagamentos_configuracoes
		WHERE id = 1
	`

	var record entity.SettingsRecord
	err := r.pool.QueryRow(ctx, sql).Scan(
		&record.DefaultProvider,
		&record.TimeoutSeconds,
		&record.ContactEmail,
		&record.MercadoPagoEnabled,
		&record.MercadoPago.Sandbox,
		&record.MercadoPago.BaseURL,
		&record.MercadoPago.PublicKey,
		&record.MercadoPago.AccessToken,
		&record.MercadoPago.StatementDescriptor,
		&record.MercadoPago.SuccessURL,
		&record.MercadoPago.FailureURL,
		&record.MercadoPago.PendingURL,
		&record.MercadoPago.WebhookURL,
		&record.MercadoPago.BinaryMode,
		&record.MercadoPago.WalletPurchase,
		&record.MercadoPago.Installments,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	return record, err
}

func (r *SettingsRepository) Upsert(ctx context.Context, input entity.SettingsRecord) error {
	const sql = `
		INSERT INTO pagamentos_configuracoes (
			id,
			provider_padrao,
			timeout_segundos,
			email_contato,
			mercado_pago_habilitado,
			mercado_pago_sandbox,
			mercado_pago_base_url,
			mercado_pago_public_key,
			mercado_pago_access_token,
			mercado_pago_statement_descriptor,
			mercado_pago_success_url,
			mercado_pago_failure_url,
			mercado_pago_pending_url,
			mercado_pago_webhook_url,
			mercado_pago_binary_mode,
			mercado_pago_wallet_purchase,
			mercado_pago_installments,
			atualizado_em
		)
		VALUES (
			1,
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,
			NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			provider_padrao = EXCLUDED.provider_padrao,
			timeout_segundos = EXCLUDED.timeout_segundos,
			email_contato = EXCLUDED.email_contato,
			mercado_pago_habilitado = EXCLUDED.mercado_pago_habilitado,
			mercado_pago_sandbox = EXCLUDED.mercado_pago_sandbox,
			mercado_pago_base_url = EXCLUDED.mercado_pago_base_url,
			mercado_pago_public_key = EXCLUDED.mercado_pago_public_key,
			mercado_pago_access_token = EXCLUDED.mercado_pago_access_token,
			mercado_pago_statement_descriptor = EXCLUDED.mercado_pago_statement_descriptor,
			mercado_pago_success_url = EXCLUDED.mercado_pago_success_url,
			mercado_pago_failure_url = EXCLUDED.mercado_pago_failure_url,
			mercado_pago_pending_url = EXCLUDED.mercado_pago_pending_url,
			mercado_pago_webhook_url = EXCLUDED.mercado_pago_webhook_url,
			mercado_pago_binary_mode = EXCLUDED.mercado_pago_binary_mode,
			mercado_pago_wallet_purchase = EXCLUDED.mercado_pago_wallet_purchase,
			mercado_pago_installments = EXCLUDED.mercado_pago_installments,
			atualizado_em = NOW()
	`

	_, err := r.pool.Exec(
		ctx,
		sql,
		input.DefaultProvider,
		input.TimeoutSeconds,
		input.ContactEmail,
		input.MercadoPagoEnabled,
		input.MercadoPago.Sandbox,
		input.MercadoPago.BaseURL,
		input.MercadoPago.PublicKey,
		input.MercadoPago.AccessToken,
		input.MercadoPago.StatementDescriptor,
		input.MercadoPago.SuccessURL,
		input.MercadoPago.FailureURL,
		input.MercadoPago.PendingURL,
		input.MercadoPago.WebhookURL,
		input.MercadoPago.BinaryMode,
		input.MercadoPago.WalletPurchase,
		input.MercadoPago.Installments,
	)
	return err
}
