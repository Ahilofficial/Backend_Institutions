//go:build wireinject
// +build wireinject

package wire

import (
	"backend_institutions/EmailSender/repository"
	"backend_institutions/EmailSender/service"

	"github.com/google/wire"
)

func InitializeNotificationService() (*service.NotificationService, error) {
	wire.Build(
		repository.NewEmailRepository,
		service.NewNotificationService,
	)
	return nil, nil
}
