package mock

import (
	"aviation-weather/internal/domain"

	"github.com/stretchr/testify/mock"
)

// Fake service that won't call any API or functionalities
type ServiceMock struct {
	mock.Mock
}

// SyncAirportQueued implements service.ServiceInterface.
func (m *ServiceMock) SyncAirportQueued(faa string) (*domain.Airport, error) {
	args := m.Called(faa)
	return args.Get(0).(*domain.Airport), args.Error(1)
}

// SyncAllAirportsQueued implements service.ServiceInterface.
func (m *ServiceMock) SyncAllAirportsQueued() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *ServiceMock) CreateAirport(a *domain.Airport) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *ServiceMock) UpdateAirport(a *domain.Airport) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *ServiceMock) DeleteAirportByFAA(faa string) error {
	args := m.Called(faa)
	return args.Error(0)
}

func (m *ServiceMock) GetAirportByFAA(faa string) (*domain.Airport, error) {
	args := m.Called(faa)
	return args.Get(0).(*domain.Airport), args.Error(1)
}

func (m *ServiceMock) GetAllAirports() ([]domain.Airport, error) {
	args := m.Called()
	return args.Get(0).([]domain.Airport), args.Error(1)
}

func (m *ServiceMock) SyncAirportByFAA(faa string) (*domain.Airport, error) {
	args := m.Called(faa)
	return args.Get(0).(*domain.Airport), args.Error(1)
}

func (m *ServiceMock) SyncAllAirports() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}
