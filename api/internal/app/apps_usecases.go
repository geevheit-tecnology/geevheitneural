package app

import (
	"context"
	"fmt"
)

type ListAppsUseCase struct {
	appRepo AppRepository
	uow     UnitOfWork
}

func NewListAppsUseCase(appRepo AppRepository, uow ...UnitOfWork) *ListAppsUseCase {
	uc := &ListAppsUseCase{appRepo: appRepo}
	if len(uow) > 0 {
		uc.uow = uow[0]
	}
	return uc
}

func (uc *ListAppsUseCase) Execute(ctx context.Context) ([]AppDTO, error) {
	if uc.uow != nil {
		var apps []AppDTO
		err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.Apps == nil {
				return fmt.Errorf("repositório de aplicativos indisponível")
			}
			var err error
			apps, err = tx.Apps.ListWithFinancials(ctx)
			return err
		})
		return apps, err
	}
	return uc.appRepo.ListWithFinancials(ctx)
}

type CreateAppUseCase struct {
	appRepo   AppRepository
	publisher EventPublisher
	uow       UnitOfWork
}

func NewCreateAppUseCase(appRepo AppRepository, publisher EventPublisher, uow ...UnitOfWork) *CreateAppUseCase {
	uc := &CreateAppUseCase{appRepo: appRepo, publisher: publisher}
	if len(uow) > 0 {
		uc.uow = uow[0]
	}
	return uc
}

func (uc *CreateAppUseCase) Execute(ctx context.Context, cmd CreateAppCommand) (*AppDTO, error) {
	if cmd.Name == "" {
		return nil, fmt.Errorf("nome do app é obrigatório")
	}
	var app *AppDTO
	var err error
	if uc.uow != nil {
		err = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.Apps == nil {
				return fmt.Errorf("repositório de aplicativos indisponível")
			}
			app, err = tx.Apps.Create(ctx, cmd)
			return err
		})
	} else {
		app, err = uc.appRepo.Create(ctx, cmd)
	}
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
	uow       UnitOfWork
}

func NewUpdateAppUseCase(appRepo AppRepository, publisher EventPublisher, uow ...UnitOfWork) *UpdateAppUseCase {
	uc := &UpdateAppUseCase{appRepo: appRepo, publisher: publisher}
	if len(uow) > 0 {
		uc.uow = uow[0]
	}
	return uc
}

func (uc *UpdateAppUseCase) Execute(ctx context.Context, cmd UpdateAppCommand) error {
	if cmd.ID == "" || cmd.Name == "" {
		return fmt.Errorf("id e nome são obrigatórios")
	}
	var err error
	if uc.uow != nil {
		err = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.Apps == nil {
				return fmt.Errorf("repositório de aplicativos indisponível")
			}
			return tx.Apps.Update(ctx, cmd)
		})
	} else {
		err = uc.appRepo.Update(ctx, cmd)
	}
	if err != nil {
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
	uow       UnitOfWork
}

func NewDeleteAppUseCase(appRepo AppRepository, publisher EventPublisher, uow ...UnitOfWork) *DeleteAppUseCase {
	uc := &DeleteAppUseCase{appRepo: appRepo, publisher: publisher}
	if len(uow) > 0 {
		uc.uow = uow[0]
	}
	return uc
}

func (uc *DeleteAppUseCase) Execute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id é obrigatório")
	}
	var err error
	if uc.uow != nil {
		err = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.Apps == nil {
				return fmt.Errorf("repositório de aplicativos indisponível")
			}
			return tx.Apps.Delete(ctx, id)
		})
	} else {
		err = uc.appRepo.Delete(ctx, id)
	}
	if err != nil {
		return err
	}
	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicAppsChanged, id)
	}
	return nil
}
