package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/google/uuid"
)

const factoryToolConnectorName = "Ferramentas de criação — provedor de IA"

type FactoryToolSettingsRepository struct{ q *sqlcgen.Queries }

func (r *FactoryToolSettingsRepository) Get(ctx context.Context) (*app.FactoryToolSelection, error) {
	rows, err := r.q.ListConnectors(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.Name == factoryToolConnectorName {
			return &app.FactoryToolSelection{Provider: row.Scope, Active: row.Active}, nil
		}
	}
	return nil, nil
}

func (r *FactoryToolSettingsRepository) Save(ctx context.Context, ownerID string, selected app.FactoryToolSelection) error {
	owner, err := uuid.Parse(ownerID)
	if err != nil {
		return fmt.Errorf("identidade do dono inválida")
	}
	now := time.Now()
	_, err = r.q.UpsertConnector(ctx, sqlcgen.UpsertConnectorParams{
		ID:      textToUUID(uuid.NewSHA1(owner, []byte("factory-tools")).String()),
		OwnerID: ownerUUID(ctx, ownerID), Name: factoryToolConnectorName,
		Active: selected.Active, Scope: selected.Provider, SecretRef: "",
		CreatedAt: timeToTimestamptz(now), UpdatedAt: timeToTimestamptz(now),
	})
	return err
}
