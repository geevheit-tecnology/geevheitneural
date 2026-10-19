package app

import (
	"context"
	"fmt"
)

type ListAppsUseCase struct {
	appRepo AppRepository
}

func NewListAppsUseCase(appRepo AppRepository) *ListAppsUseCase {
	return &ListAppsUseCase{appRepo: appRepo}
}

func (uc *ListAppsUseCase) Execute(ctx context.Context) ([]AppDTO, error) {
	return uc.appRepo.ListWithFinancials(ctx)
}

type CreateAppUseCase struct {
	appRepo   AppRepository
	publisher EventPublisher
}

func NewCreateAppUseCase(appRepo AppRepository, publisher EventPublisher) *CreateAppUseCase {
	return &CreateAppUseCase{appRepo: appRepo, publisher: publisher}
}

func (uc *CreateAppUseCase) Execute(ctx context.Context, cmd CreateAppCommand) (*AppDTO, error) {
	if cmd.Name == "" {
		return nil, fmt.Errorf("nome do app é obrigatório")
	}
	app, err := uc.appRepo.Create(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicAppsChanged, app)
	}
	return app, nil
}

type UpdateAppUseCase struct {
	appRepo   AppRepository
	publisher EventPublisher
}

func NewUpdateAppUseCase(appRepo AppRepository, publisher EventPublisher) *UpdateAppUseCase {
	return &UpdateAppUseCase{appRepo: appRepo, publisher: publisher}
}

func (uc *UpdateAppUseCase) Execute(ctx context.Context, cmd UpdateAppCommand) error {
	if cmd.ID == "" || cmd.Name == "" {
		return fmt.Errorf("id e nome são obrigatórios")
	}
	if err := uc.appRepo.Update(ctx, cmd); err != nil {
		return err
	}
	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicAppsChanged, cmd.ID)
	}
	return nil
}

type DeleteAppUseCase struct {
	appRepo   AppRepository
	publisher EventPublisher
}

func NewDeleteAppUseCase(appRepo AppRepository, publisher EventPublisher) *DeleteAppUseCase {
	return &DeleteAppUseCase{appRepo: appRepo, publisher: publisher}
}

func (uc *DeleteAppUseCase) Execute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id é obrigatório")
	}
	if err := uc.appRepo.Delete(ctx, id); err != nil {
		return err
	}
	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicAppsChanged, id)
	}
	return nil
}
