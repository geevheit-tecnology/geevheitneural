package app

import (
	"context"
)

type ListConnectorsUseCase struct {
	connectorRepo ConnectorRepository
}

func NewListConnectorsUseCase(connectorRepo ConnectorRepository) *ListConnectorsUseCase {
	return &ListConnectorsUseCase{connectorRepo: connectorRepo}
}

func (uc *ListConnectorsUseCase) Execute(ctx context.Context) ([]ConnectorDTO, error) {
	return uc.connectorRepo.List(ctx)
}
