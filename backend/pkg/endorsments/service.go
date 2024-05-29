package endorsments

import (
	"github.com/go-kit/log"
)

type AppService struct {
	repo   Repository
	logger log.Logger
}

type Repository interface {
	GetEndorsmentRequest()
	ListEndorsmentRequests()
	CreateEndorsmentRequest()
	UpdateEndorsmentRequest()
	DeleteEndorsmentRequest()
}

func NewAppService(repository Repository, logger log.Logger) *AppService {
	return &AppService{
		repo:   repository,
		logger: logger,
	}
}
