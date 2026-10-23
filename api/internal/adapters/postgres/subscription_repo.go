package postgres

import (
	"context"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/domain/subscription"
	"github.com/jackc/pgx/v5/pgtype"
)

type SubscriptionRepository struct {
	q *sqlcgen.Queries
}

func NewSubscriptionRepository(q *sqlcgen.Queries) *SubscriptionRepository {
	return &SubscriptionRepository{q: q}
}

func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *subscription.Subscription) error {
	var appID pgtype.UUID
	if sub.AppID != nil {
		appID = textToUUID(*sub.AppID)
	}

	var endDate pgtype.Date
	if sub.EndDate != nil {
		endDate = pgtype.Date{Time: *sub.EndDate, Valid: true}
	}

	return r.q.CreateSubscription(ctx, sqlcgen.CreateSubscriptionParams{
		ID:          textToUUID(sub.ID),
		OwnerID:     textToUUID(sub.OwnerID),
		ClientID:    textToUUID(sub.ClientID),
		AppID:       appID,
		Plan:        sub.Plan,
		Description: sub.Description,
		AmountCents: sub.AmountCents,
		Currency:    sub.Currency,
		Interval:    string(sub.Interval),
		DueDay:      int32(sub.DueDay),
		Status:      string(sub.Status),
		StartDate:   pgtype.Date{Time: sub.StartDate, Valid: true},
		EndDate:     endDate,
		NextDueDate: pgtype.Date{Time: sub.NextDueDate, Valid: true},
		CreatedAt:   timeToTimestamptz(sub.CreatedAt),
		UpdatedAt:   timeToTimestamptz(sub.UpdatedAt),
	})
}

func (r *SubscriptionRepository) GetSubscriptionByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	row, err := r.q.GetSubscriptionByID(ctx, textToUUID(id))
	if err != nil {
		return nil, err
	}
	return mapSubscriptionRow(row), nil
}

func (r *SubscriptionRepository) ListSubscriptions(ctx context.Context, clientID *string, status *string) ([]subscription.Subscription, error) {
	var cID pgtype.UUID
	if clientID != nil {
		cID = textToUUID(*clientID)
	}

	var st string
	if status != nil {
		st = *status
	}

	rows, err := r.q.ListSubscriptions(ctx, sqlcgen.ListSubscriptionsParams{
		Column1: cID,
		Column2: st,
	})
	if err != nil {
		return nil, err
	}

	subs := make([]subscription.Subscription, len(rows))
	for i, row := range rows {
		subs[i] = *mapSubscriptionRow(row)
	}
	return subs, nil
}

func (r *SubscriptionRepository) UpdateSubscriptionStatus(ctx context.Context, id string, status subscription.Status, nextDue *time.Time) error {
	var nextDueDate pgtype.Date
	if nextDue != nil {
		nextDueDate = pgtype.Date{Time: *nextDue, Valid: true}
	}

	return r.q.UpdateSubscriptionStatus(ctx, sqlcgen.UpdateSubscriptionStatusParams{
		ID:          textToUUID(id),
		Status:      string(status),
		NextDueDate: nextDueDate,
	})
}

func (r *SubscriptionRepository) UpdateSubscriptionNextDueDate(ctx context.Context, id string, nextDue time.Time) error {
	return r.q.UpdateSubscriptionNextDueDate(ctx, sqlcgen.UpdateSubscriptionNextDueDateParams{
		ID:          textToUUID(id),
		NextDueDate: pgtype.Date{Time: nextDue, Valid: true},
	})
}

func (r *SubscriptionRepository) CreateCharge(ctx context.Context, charge *subscription.SubscriptionCharge) error {
	return r.q.CreateSubscriptionCharge(ctx, sqlcgen.CreateSubscriptionChargeParams{
		ID:             textToUUID(charge.ID),
		OwnerID:        textToUUID(charge.OwnerID),
		SubscriptionID: textToUUID(charge.SubscriptionID),
		AmountCents:    charge.AmountCents,
		Currency:       charge.Currency,
		DueDate:        pgtype.Date{Time: charge.DueDate, Valid: true},
		Status:         string(charge.Status),
		PaymentMethod:  charge.PaymentMethod,
		CreatedAt:      timeToTimestamptz(charge.CreatedAt),
		UpdatedAt:      timeToTimestamptz(charge.UpdatedAt),
	})
}

func (r *SubscriptionRepository) GetChargeByID(ctx context.Context, id string) (*subscription.SubscriptionCharge, error) {
	row, err := r.q.GetSubscriptionChargeByID(ctx, textToUUID(id))
	if err != nil {
		return nil, err
	}
	return mapSubscriptionChargeRow(row), nil
}

func (r *SubscriptionRepository) ListCharges(ctx context.Context, status *string, dueBefore *time.Time) ([]subscription.SubscriptionCharge, error) {
	var due pgtype.Date
	if dueBefore != nil {
		due = pgtype.Date{Time: *dueBefore, Valid: true}
	}

	var st string
	if status != nil {
		st = *status
	}

	rows, err := r.q.ListSubscriptionCharges(ctx, sqlcgen.ListSubscriptionChargesParams{
		Column1: st,
		Column2: due,
	})
	if err != nil {
		return nil, err
	}

	charges := make([]subscription.SubscriptionCharge, len(rows))
	for i, row := range rows {
		charges[i] = *mapSubscriptionChargeRow(row)
	}
	return charges, nil
}

func (r *SubscriptionRepository) UpdateChargeStatus(ctx context.Context, id string, status subscription.ChargeStatus, paidAt *time.Time, method string) error {
	var paid pgtype.Timestamptz
	if paidAt != nil {
		paid = timeToTimestamptz(*paidAt)
	}

	return r.q.UpdateSubscriptionChargeStatus(ctx, sqlcgen.UpdateSubscriptionChargeStatusParams{
		ID:            textToUUID(id),
		Status:        string(status),
		PaidAt:        paid,
		PaymentMethod: method,
	})
}

func mapSubscriptionRow(row sqlcgen.GeevSubscription) *subscription.Subscription {
	var appID *string
	if row.AppID.Valid {
		str := uuidToText(row.AppID)
		appID = &str
	}

	var endDate *time.Time
	if row.EndDate.Valid {
		endDate = &row.EndDate.Time
	}

	return &subscription.Subscription{
		ID:          uuidToText(row.ID),
		OwnerID:     uuidToText(row.OwnerID),
		ClientID:    uuidToText(row.ClientID),
		AppID:       appID,
		Plan:        row.Plan,
		Description: row.Description,
		AmountCents: row.AmountCents,
		Currency:    row.Currency,
		Interval:    subscription.Interval(row.Interval),
		DueDay:      int(row.DueDay),
		Status:      subscription.Status(row.Status),
		StartDate:   row.StartDate.Time,
		EndDate:     endDate,
		NextDueDate: row.NextDueDate.Time,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func mapSubscriptionChargeRow(row sqlcgen.GeevSubscriptionCharge) *subscription.SubscriptionCharge {
	var paidAt *time.Time
	if row.PaidAt.Valid {
		paidAt = &row.PaidAt.Time
	}

	return &subscription.SubscriptionCharge{
		ID:             uuidToText(row.ID),
		OwnerID:        uuidToText(row.OwnerID),
		SubscriptionID: uuidToText(row.SubscriptionID),
		AmountCents:    row.AmountCents,
		Currency:       row.Currency,
		DueDate:        row.DueDate.Time,
		Status:         subscription.ChargeStatus(row.Status),
		PaidAt:         paidAt,
		PaymentMethod:  row.PaymentMethod,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}
