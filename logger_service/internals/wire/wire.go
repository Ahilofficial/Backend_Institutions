//go:build wireinject
// +build wireinject

package wire

import (
	"backend_institutions/logger_service/internals/repository"
	"backend_institutions/logger_service/internals/services"

	"github.com/google/wire"
)

func InitializeLoggerService() (*services.LoggerService, error) {
	wire.Build(
		repository.NewLoggerRepo,
		services.NewLoggerService,
	)
	return nil, nil
}
