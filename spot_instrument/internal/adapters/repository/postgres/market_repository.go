package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/errs"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
	"go.uber.org/zap"
)

type MarketRepository struct {
	db        *pgxpool.Pool
	roleMap   map[string]int
	roleMapMu sync.RWMutex

	logger *zap.Logger
}

func NewMarketRepository(logger *zap.Logger, db *pgxpool.Pool) (*MarketRepository, error) {
	r := &MarketRepository{
		db:     db,
		logger: logger,
	}

	if err := r.loadRoleMap(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to load roles: %w", err)
	}
	return r, nil
}

func (r *MarketRepository) loadRoleMap(ctx context.Context) error {
	rows, err := r.db.Query(ctx, `SELECT id, code FROM roles`)
	if err != nil {
		return err
	}
	defer rows.Close()

	m := make(map[string]int)
	for rows.Next() {
		var id int
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return err
		}
		m[code] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}

	r.roleMapMu.Lock()
	r.roleMap = m
	r.roleMapMu.Unlock()

	return nil
}

func (r *MarketRepository) roleCodeToID(code model.UserRole) (int, bool) {
	r.roleMapMu.RLock()
	defer r.roleMapMu.RUnlock()

	id, ok := r.roleMap[string(code)]

	return id, ok
}

func (r *MarketRepository) FindEnabledByRolesPaginated(ctx context.Context, roles []model.UserRole, pageToken model.PageToken, limit int32) (*model.PaginationData, error) {
	roleIDs := make([]int, 0, len(roles))
	for _, rl := range roles {
		if id, ok := r.roleCodeToID(rl); ok {
			roleIDs = append(roleIDs, id)
		}
	}

	query := `
        SELECT m.uuid, m.name, m.is_enabled, m.deleted_at, m.created_at, m.updated_at,
               COALESCE(array_agg(r.code) FILTER (WHERE r.code IS NOT NULL), '{}') AS role_codes
        FROM markets m
        LEFT JOIN market_allowed_roles mar ON m.uuid = mar.market_uuid
        LEFT JOIN roles r ON mar.role_id = r.id
        WHERE m.is_enabled = true AND m.deleted_at IS NULL
    `

	cursor, err := pageToken.Decode()
	if err != nil {
		r.logger.Warn("failed decode pagination token", zap.Error(err))
	}

	if cursor.MarketName != "" && cursor.MarketUuid != "" {
		query += ` AND (m.name, m.uuid) > ($2, $3)`
	}

	query += `
        GROUP BY m.uuid
        HAVING COUNT(mar.role_id) = 0 OR array_agg(mar.role_id) && $1
        ORDER BY m.name, m.uuid
        LIMIT $4
    `

	rows, err := r.db.Query(ctx, query, roleIDs, cursor.MarketName, cursor.MarketUuid, limit+1)
	if err != nil {
		return nil, fmt.Errorf("failed to query markets: %w", err)
	}

	defer rows.Close()

	var markets []*model.Market
	for rows.Next() {
		var market model.Market
		var deletedAt sql.NullTime
		var roleCodes []string

		err := rows.Scan(
			&market.UUID,
			&market.Name,
			&market.IsEnabled,
			&deletedAt,
			&market.CreatedAt,
			&market.UpdatedAt,
			&roleCodes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan market: %w", err)
		}

		if deletedAt.Valid {
			market.DeletedAt = &deletedAt.Time
		}

		market.AllowedRoles = make([]model.UserRole, len(roleCodes))
		for i, code := range roleCodes {
			market.AllowedRoles[i] = model.UserRole(code)
		}

		markets = append(markets, &market)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed handle rows: %w", err)
	}

	var nextPageToken model.PageToken
	hasNext := len(markets) > int(limit)
	if hasNext {
		markets = markets[:limit]

		lastMarket := markets[len(markets)-1]
		cursor := model.PaginationCursor{MarketName: lastMarket.Name, MarketUuid: lastMarket.UUID}
		nextPageToken = cursor.Encode()
	}

	return &model.PaginationData{
		Markets:       markets,
		HasNext:       hasNext,
		NextPageToken: nextPageToken,
	}, nil
}

func (r *MarketRepository) FindEnabledByRoles(ctx context.Context, roles []model.UserRole) ([]*model.Market, error) {
	roleIDs := make([]int, 0, len(roles))
	for _, rl := range roles {
		if id, ok := r.roleCodeToID(rl); ok {
			roleIDs = append(roleIDs, id)
		}
	}

	query := `
            SELECT m.uuid, m.name, m.is_enabled, m.deleted_at, m.created_at, m.updated_at,
                   COALESCE(array_agg(r.code) FILTER (WHERE r.code IS NOT NULL), '{}') AS role_codes
            FROM markets m
            LEFT JOIN market_allowed_roles mar ON m.uuid = mar.market_uuid
            LEFT JOIN roles r ON mar.role_id = r.id
            WHERE m.is_enabled = true AND m.deleted_at IS NULL
            GROUP BY m.uuid
            HAVING COUNT(mar.role_id) = 0 OR array_agg(mar.role_id) && $1
            ORDER BY m.name
        `

	rows, err := r.db.Query(ctx, query, roleIDs)

	if err != nil {
		return nil, fmt.Errorf("failed to query markets: %w", err)
	}
	defer rows.Close()

	markets := make([]*model.Market, 0)
	for rows.Next() {
		var market model.Market
		var deletedAt sql.NullTime
		var roleCodes []string

		err := rows.Scan(
			&market.UUID,
			&market.Name,
			&market.IsEnabled,
			&deletedAt,
			&market.CreatedAt,
			&market.UpdatedAt,
			&roleCodes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan market: %w", err)
		}

		if deletedAt.Valid {
			market.DeletedAt = &deletedAt.Time
		}

		market.AllowedRoles = make([]model.UserRole, len(roleCodes))
		for i, code := range roleCodes {
			market.AllowedRoles[i] = model.UserRole(code)
		}

		markets = append(markets, &market)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return markets, nil
}

func (r *MarketRepository) FindByUUID(ctx context.Context, uuid string) (*model.Market, error) {
	query := `
        SELECT m.uuid, m.name, m.is_enabled, m.deleted_at, m.created_at, m.updated_at,
               COALESCE(array_agg(r.code) FILTER (WHERE r.code IS NOT NULL), '{}') AS role_codes
        FROM markets m
        LEFT JOIN market_allowed_roles mar ON m.uuid = mar.market_uuid
        LEFT JOIN roles r ON mar.role_id = r.id
        WHERE m.uuid = $1
        GROUP BY m.uuid
    `
	var m model.Market
	var deletedAt sql.NullTime
	var roleCodes []string

	err := r.db.QueryRow(ctx, query, uuid).Scan(
		&m.UUID,
		&m.Name,
		&m.IsEnabled,
		&deletedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
		&roleCodes,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("market not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find market: %w", err)
	}

	if deletedAt.Valid {
		m.DeletedAt = &deletedAt.Time
	}

	m.AllowedRoles = make([]model.UserRole, len(roleCodes))
	for i, code := range roleCodes {
		m.AllowedRoles[i] = model.UserRole(code)
	}

	return &m, nil
}

func (r *MarketRepository) Create(ctx context.Context, market *model.Market) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queryMarket := `
        INSERT INTO markets (uuid, name, is_enabled, deleted_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err = tx.Exec(ctx, queryMarket,
		market.UUID,
		market.Name,
		market.IsEnabled,
		market.DeletedAt,
		market.CreatedAt,
		market.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert market: %w", err)
	}

	if err := r.syncMarketRoles(ctx, tx, market.UUID, market.AllowedRoles); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *MarketRepository) Update(ctx context.Context, market *model.Market) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM markets WHERE uuid = $1)`, market.UUID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: market not found", errs.ErrNotFound)
	}

	queryMarket := `
        UPDATE markets
        SET name = $1, is_enabled = $2, deleted_at = $3, updated_at = $4
        WHERE uuid = $5
    `
	_, err = tx.Exec(ctx,
		queryMarket,
		market.Name,
		market.IsEnabled,
		market.DeletedAt,
		market.UpdatedAt,
		market.UUID,
	)
	if err != nil {
		return fmt.Errorf("failed to update market: %w", err)
	}

	if err := r.syncMarketRoles(ctx, tx, market.UUID, market.AllowedRoles); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *MarketRepository) Delete(ctx context.Context, uuid string) error {
	query := `UPDATE markets SET deleted_at = NOW(), updated_at = NOW() WHERE uuid = $1`

	_, err := r.db.Exec(ctx, query, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete market: %w", err)
	}

	return nil
}

func (r *MarketRepository) syncMarketRoles(ctx context.Context, tx pgx.Tx, marketUUID string, roles []model.UserRole) error {
	_, err := tx.Exec(ctx, `DELETE FROM market_allowed_roles WHERE market_uuid = $1`, marketUUID)
	if err != nil {
		return fmt.Errorf("failed to delete existing roles: %w", err)
	}

	if len(roles) == 0 {
		return nil
	}

	roleIDs := make([]int, 0, len(roles))
	for _, rl := range roles {
		id, ok := r.roleCodeToID(rl)
		if !ok {
			r.logger.Warn("skipping unknown role", zap.String("role", string(rl)))
			continue
		}
		roleIDs = append(roleIDs, id)
	}

	if len(roleIDs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, id := range roleIDs {
		batch.Queue(`INSERT INTO market_allowed_roles (market_uuid, role_id) VALUES ($1, $2)`, marketUUID, id)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(roleIDs); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to insert role id %d: %w", roleIDs[i], err)
		}
	}

	return nil
}
