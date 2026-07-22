package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"sistema-editorial/editora/backend/src/logistica/entity"
)

type CatalogRepository struct {
	pool *pgxpool.Pool
}

func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

func (r *CatalogRepository) LookupCatalogItems(
	ctx context.Context,
	ids []string,
) (map[string]entity.CatalogItem, error) {
	const sql = `
		SELECT
			id,
			titulo,
			coalesce(peso_gramas, 0),
			coalesce(largura_cm, 0),
			coalesce(altura_cm, 0),
			coalesce(profundidade_cm, 0),
			coalesce(preco_venda, 0),
			coalesce(preco_venda_fisico, 0),
			possui_formato_fisico,
			ativo
		FROM livros
		WHERE id = ANY($1)
	`

	rows, err := r.pool.Query(ctx, sql, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[string]entity.CatalogItem, len(ids))
	for rows.Next() {
		var item entity.CatalogItem
		if err := rows.Scan(
			&item.LivroID,
			&item.Titulo,
			&item.PesoGramas,
			&item.LarguraCM,
			&item.AlturaCM,
			&item.ProfundidadeCM,
			&item.PrecoVenda,
			&item.PrecoVendaFisico,
			&item.PossuiFormatoFisico,
			&item.Ativo,
		); err != nil {
			return nil, err
		}
		items[item.LivroID] = item
	}

	return items, rows.Err()
}
