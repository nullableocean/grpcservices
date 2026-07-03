package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type listQuery struct {
	query  string
	args   []interface{}
	argIdx int
}

func newListQuery(userUUID string) *listQuery {
	return &listQuery{
		query: `
			SELECT o.uuid, o.user_uuid, o.market_uuid,
				   s.code AS order_side,
				   t.code AS order_type,
				   st.code AS order_status,
				   o.price, o.quantity, o.created_at, o.updated_at
			FROM orders o
			JOIN order_sides s ON o.side_id = s.id
			JOIN order_types t ON o.order_type_id = t.id
			JOIN order_statuses st ON o.order_status_id = st.id
			WHERE o.user_uuid = $1
		`,
		args:   []interface{}{userUUID},
		argIdx: 1,
	}
}

func (q *listQuery) addPlaceholders(count int) string {
	if count == 0 {
		return ""
	}

	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		q.argIdx++
		placeholders[i] = fmt.Sprintf("$%d", q.argIdx)
	}

	return strings.Join(placeholders, ", ")
}

func (r *OrderRepository) List(ctx context.Context, userUUID string, filters model.OrderListFilter) (model.OrderList, error) {
	if err := filters.Validate(); err != nil {
		return model.OrderList{}, err
	}

	q := newListQuery(userUUID)

	if err := r.addStatusFilter(q, filters.Statuses); err != nil {
		return model.OrderList{}, err
	}

	r.addMarketFilter(q, filters.MarketUUIDs)

	if err := r.addTypeFilter(q, filters.Type); err != nil {
		return model.OrderList{}, err
	}

	r.addDateRangeFilter(q, filters.CreatedFrom, filters.CreatedTo)
	r.addCursorFilter(q, filters.PageCursor)

	limit := filters.Limit
	r.addOrderByAndLimit(q, limit+1)

	query, args := q.query, q.args
	rows, err := r.pgpool.Query(ctx, query, args...)
	if err != nil {
		return model.OrderList{}, r.mapDBError(err, "failed to query list orders")
	}
	defer rows.Close()

	orders, err := r.scanOrders(rows)
	if err != nil {
		return model.OrderList{}, err
	}

	hasNext, nextCursor := r.computeNextPage(orders, limit)
	return model.OrderList{
		List:           orders,
		HasNextPage:    hasNext,
		NextPageCursor: nextCursor,
	}, nil
}

func (r *OrderRepository) addStatusFilter(q *listQuery, statuses []model.OrderStatus) error {
	if len(statuses) == 0 {
		return nil
	}

	ids, err := r.getStatusesIDs(context.Background(), statuses)
	if err != nil {
		return err
	}

	for _, id := range ids {
		q.args = append(q.args, id)
	}

	q.query += fmt.Sprintf(" AND o.order_status_id IN (%s)", q.addPlaceholders(len(ids)))
	return nil
}

func (r *OrderRepository) addMarketFilter(q *listQuery, marketUUIDs []string) {
	if len(marketUUIDs) == 0 {
		return
	}

	for _, uuid := range marketUUIDs {
		q.args = append(q.args, uuid)
	}

	q.query += fmt.Sprintf(" AND o.market_uuid IN (%s)", q.addPlaceholders(len(marketUUIDs)))
}

func (r *OrderRepository) addTypeFilter(q *listQuery, orderType *model.OrderType) error {
	if orderType == nil {
		return nil
	}

	id, err := r.getTypeIDByCode(context.Background(), *orderType)
	if err != nil {
		return err
	}

	q.argIdx++
	q.args = append(q.args, id)
	q.query += fmt.Sprintf(" AND o.order_type_id = $%d", q.argIdx)

	return nil
}

func (r *OrderRepository) addDateRangeFilter(q *listQuery, from, to *time.Time) {
	if from != nil {
		q.argIdx++
		q.args = append(q.args, *from)
		q.query += fmt.Sprintf(" AND o.created_at >= $%d", q.argIdx)
	}

	if to != nil {
		q.argIdx++
		q.args = append(q.args, *to)
		q.query += fmt.Sprintf(" AND o.created_at <= $%d", q.argIdx)
	}
}

func (r *OrderRepository) addCursorFilter(q *listQuery, cursor *model.PaginationCursor) {
	if cursor == nil || cursor.CreatedAt.IsZero() || cursor.OrderUUID == "" {
		return
	}

	q.argIdx += 2
	q.args = append(q.args, cursor.CreatedAt, cursor.OrderUUID)
	q.query += fmt.Sprintf(" AND (o.created_at, o.uuid) < ($%d, $%d)", q.argIdx-1, q.argIdx)
}

func (r *OrderRepository) addOrderByAndLimit(q *listQuery, limit int) {
	q.argIdx++
	q.args = append(q.args, limit)
	q.query += fmt.Sprintf(" ORDER BY o.created_at DESC, o.uuid ASC LIMIT $%d", q.argIdx)
}

func (r *OrderRepository) getStatusesIDs(ctx context.Context, statuses []model.OrderStatus) ([]int, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	rows, err := r.pgpool.Query(ctx, `SELECT id FROM order_statuses WHERE code = ANY($1)`, statuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *OrderRepository) getTypeIDByCode(ctx context.Context, orderType model.OrderType) (int, error) {
	var id int
	err := r.pgpool.QueryRow(ctx, `SELECT id FROM order_types WHERE code = $1`, orderType).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *OrderRepository) scanOrders(rows pgx.Rows) ([]*model.Order, error) {
	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		var sideCode, typeCode, statusCode string
		err := rows.Scan(
			&o.UUID,
			&o.UserUUID,
			&o.MarketUUID,
			&sideCode,
			&typeCode,
			&statusCode,
			&o.Price,
			&o.Quantity,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, r.mapDBError(err, "failed to scan order")
		}

		o.Side = model.OrderSide(sideCode)
		o.Type = model.OrderType(typeCode)
		o.Status = model.OrderStatus(statusCode)

		orders = append(orders, &o)
	}

	if err := rows.Err(); err != nil {
		return nil, r.mapDBError(err, "rows iteration error")
	}

	return orders, nil
}

func (r *OrderRepository) computeNextPage(orders []*model.Order, limit int) (bool, *model.PaginationCursor) {
	hasNext := len(orders) > limit
	if !hasNext {
		return false, nil
	}

	orders = orders[:limit]

	last := orders[len(orders)-1]
	cursor := &model.PaginationCursor{
		CreatedAt: last.CreatedAt,
		OrderUUID: last.UUID,
	}

	return true, cursor
}
