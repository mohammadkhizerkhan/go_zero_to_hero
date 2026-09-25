package testing_demo

import "fmt"

//go:generate mockgen -source=weather.go -destination=mocks/mock_weather.go -package=mocks

// Interface contract
type WeatherAPI interface {
	GetTemperature(city string) (float64, error)
}

// Service depending on the interface
type WeatherAdvisor struct {
	api WeatherAPI
}

func NewWeatherAdvisor(api WeatherAPI) *WeatherAdvisor {
	return &WeatherAdvisor{api: api}
}

func (w *WeatherAdvisor) RecommendOutfit(city string) (string, error) {
	temp, err := w.api.GetTemperature(city)
	if err != nil {
		return "", fmt.Errorf("could not fetch weather: %w", err)
	}

	if temp < 15.0 {
		return "Wear a heavy jacket", nil
	}
	return "Wear a t-shirt", nil
}
