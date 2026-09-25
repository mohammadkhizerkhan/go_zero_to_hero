package testing_demo

import (
	"errors"
	"testing"

	"github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/testing/mocks"
	"go.uber.org/mock/gomock"
)

// Mock implementation satisfying WeatherAPI
type MockWeatherAPI struct {
	GetTemperatureFunc func(city string) (float64, error)
}

func (m *MockWeatherAPI) GetTemperature(city string) (float64, error) {
	return m.GetTemperatureFunc(city)
}

func TestRecommendOutfit(t *testing.T) {
	t.Run("Cold weather recommendation", func(t *testing.T) {
		mockAPI := &MockWeatherAPI{
			GetTemperatureFunc: func(city string) (float64, error) {
				return 10.0, nil // Return 10°C fake temperature
			},
		}

		advisor := NewWeatherAdvisor(mockAPI)
		recommendation, err := advisor.RecommendOutfit("London")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if recommendation != "Wear a heavy jacket" {
			t.Errorf("Expected 'Wear a heavy jacket', got %q", recommendation)
		}
	})

	t.Run("Warm weather recommendation", func(t *testing.T) {
		mockAPI := &MockWeatherAPI{
			GetTemperatureFunc: func(city string) (float64, error) {
				return 25.0, nil // Return 25°C fake temperature
			},
		}

		advisor := NewWeatherAdvisor(mockAPI)
		recommendation, err := advisor.RecommendOutfit("Miami")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if recommendation != "Wear a t-shirt" {
			t.Errorf("Expected 'Wear a t-shirt', got %q", recommendation)
		}
	})

	t.Run("API receives expected arguments", func(t *testing.T) {
		var receivedCity string
		mockAPI := &MockWeatherAPI{
			GetTemperatureFunc: func(city string) (float64, error) {
				receivedCity = city
				return 20.0, nil
			},
		}

		advisor := NewWeatherAdvisor(mockAPI)
		_, err := advisor.RecommendOutfit("Tokyo")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if receivedCity != "Tokyo" {
			t.Errorf("Expected city 'Tokyo' to be passed to API, got %q", receivedCity)
		}
	})

	t.Run("API failure handling", func(t *testing.T) {
		mockAPI := &MockWeatherAPI{
			GetTemperatureFunc: func(city string) (float64, error) {
				return 0, errors.New("network timeout")
			},
		}

		advisor := NewWeatherAdvisor(mockAPI)
		_, err := advisor.RecommendOutfit("London")

		if err == nil {
			t.Fatal("Expected error when API fails, got nil")
		}
	})
}

func TestRecommendOutfit_GoMock(t *testing.T) {
	t.Run("Cold weather recommendation with GoMock", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := mocks.NewMockWeatherAPI(ctrl)
		mockAPI.EXPECT().
			GetTemperature("London").
			Return(10.0, nil).
			Times(1)

		// Inject the generated mock via constructor DI
		advisor := NewWeatherAdvisor(mockAPI)
		recommendation, err := advisor.RecommendOutfit("London")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if recommendation != "Wear a heavy jacket" {
			t.Errorf("Expected 'Wear a heavy jacket', got %q", recommendation)
		}
	})

	t.Run("Warm weather recommendation with GoMock", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := mocks.NewMockWeatherAPI(ctrl)
		mockAPI.EXPECT().
			GetTemperature("Miami").
			Return(25.0, nil).
			Times(1)

		advisor := NewWeatherAdvisor(mockAPI)
		recommendation, err := advisor.RecommendOutfit("Miami")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if recommendation != "Wear a t-shirt" {
			t.Errorf("Expected 'Wear a t-shirt', got %q", recommendation)
		}
	})

	t.Run("API failure handling with GoMock matcher", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := mocks.NewMockWeatherAPI(ctrl)
		mockAPI.EXPECT().
			GetTemperature(gomock.Any()).
			Return(0.0, errors.New("network timeout")).
			Times(1)

		advisor := NewWeatherAdvisor(mockAPI)
		_, err := advisor.RecommendOutfit("Tokyo")

		if err == nil {
			t.Fatal("Expected error when API fails, got nil")
		}
	})
}
