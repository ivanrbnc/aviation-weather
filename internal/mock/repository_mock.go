package mock

import (
	"aviation-weather/internal/domain"

	"github.com/stretchr/testify/mock"
)

// Fake repository that won't call any API or functionalities
type RepositoryMock struct {
	mock.Mock
}

func (m *RepositoryMock) CreateAirport(airport *domain.Airport) error {
	args := m.Called(airport)
	return args.Error(0)
}

func (m *RepositoryMock) UpdateAirport(airport *domain.Airport) error {
	args := m.Called(airport)
	return args.Error(0)
}

func (m *RepositoryMock) DeleteByFAA(faa string) error {
	args := m.Called(faa)
	return args.Error(0)
}

func (m *RepositoryMock) GetAllAirports() ([]domain.Airport, error) {
	args := m.Called()
	return args.Get(0).([]domain.Airport), args.Error(1)
}

func (m *RepositoryMock) GetAirportByFAA(faaFilter string) (*domain.Airport, error) {
	args := m.Called(faaFilter)
	return args.Get(0).(*domain.Airport), args.Error(1)
}
