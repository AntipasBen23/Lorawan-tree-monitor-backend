package models

import "time"

// Tree represents a tree being monitored
type Tree struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	SensorID  string    `json:"sensor_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Measurement represents sensor data from a tree
type Measurement struct {
	ID             int       `json:"id"`
	TreeID         int       `json:"tree_id"`
	SoilMoisture   float64   `json:"soil_moisture"`
	Temperature    float64   `json:"temperature"`
	Tilt           float64   `json:"tilt"`
	BatteryLevel   float64   `json:"battery_level"`
	Timestamp      time.Time `json:"timestamp"`
	ReceivedAt     time.Time `json:"received_at"`
	TokensAwarded  int       `json:"tokens_awarded"`
}

// User represents a system user
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	TokenBalance int       `json:"token_balance"`
	CreatedAt    time.Time `json:"created_at"`
}

// TokenTransaction represents token award history
type TokenTransaction struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	MeasurementID int       `json:"measurement_id"`
	Amount        int       `json:"amount"`
	Type          string    `json:"type"` // "earn", "claim", "exchange"
	CreatedAt     time.Time `json:"created_at"`
}

// LoRaWANPayload represents incoming webhook data from Network Server
type LoRaWANPayload struct {
	DeviceID     string                 `json:"deviceId"`
	Timestamp    string                 `json:"timestamp"`
	Data         map[string]interface{} `json:"data"`
	RawData      string                 `json:"rawData,omitempty"`
}