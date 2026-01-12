package token

import (
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/models"
)

// Calculator handles token calculation logic
type Calculator struct {
	TokensPerReading int
}

func NewCalculator(tokensPerReading int) *Calculator {
	return &Calculator{
		TokensPerReading: tokensPerReading,
	}
}

// CalculateTokens determines how many tokens to award for a measurement
func (c *Calculator) CalculateTokens(measurement *models.Measurement) int {
	// Basic logic: award base tokens for valid measurement
	tokens := c.TokensPerReading

	// Bonus tokens based on data quality (optional logic)
	// Example: bonus for good battery level
	if measurement.BatteryLevel > 80 {
		tokens += 2
	}

	// Example: bonus for healthy soil moisture range
	if measurement.SoilMoisture >= 20 && measurement.SoilMoisture <= 60 {
		tokens += 3
	}

	return tokens
}

// ValidateMeasurement checks if measurement data is within acceptable ranges
func (c *Calculator) ValidateMeasurement(measurement *models.Measurement) bool {
	// Soil moisture: 0-100%
	if measurement.SoilMoisture < 0 || measurement.SoilMoisture > 100 {
		return false
	}

	// Temperature: -40 to 85°C (typical sensor range)
	if measurement.Temperature < -40 || measurement.Temperature > 85 {
		return false
	}

	// Tilt: 0-360 degrees
	if measurement.Tilt < 0 || measurement.Tilt > 360 {
		return false
	}

	// Battery: 0-100%
	if measurement.BatteryLevel < 0 || measurement.BatteryLevel > 100 {
		return false
	}

	return true
}