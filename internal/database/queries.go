package database

import (
	
	"time"

	"github.com/AntipasBen23/lorawan-tree-monitor/internal/models"
)

// GetTreeBySensorID retrieves a tree by its sensor ID
func (db *DB) GetTreeBySensorID(sensorID string) (*models.Tree, error) {
	var tree models.Tree
	query := `SELECT id, name, location, sensor_id, user_id, created_at FROM trees WHERE sensor_id = $1`
	
	err := db.QueryRow(query, sensorID).Scan(
		&tree.ID,
		&tree.Name,
		&tree.Location,
		&tree.SensorID,
		&tree.UserID,
		&tree.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &tree, nil
}

// InsertMeasurement stores a new measurement
func (db *DB) InsertMeasurement(m *models.Measurement) error {
	query := `
		INSERT INTO measurements (tree_id, soil_moisture, temperature, tilt, battery_level, timestamp, tokens_awarded)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, received_at
	`
	
	err := db.QueryRow(
		query,
		m.TreeID,
		m.SoilMoisture,
		m.Temperature,
		m.Tilt,
		m.BatteryLevel,
		m.Timestamp,
		m.TokensAwarded,
	).Scan(&m.ID, &m.ReceivedAt)
	
	return err
}

// GetMeasurementsByTreeID retrieves measurement history for a tree
func (db *DB) GetMeasurementsByTreeID(treeID int, limit int) ([]models.Measurement, error) {
	query := `
		SELECT id, tree_id, soil_moisture, temperature, tilt, battery_level, timestamp, received_at, tokens_awarded
		FROM measurements
		WHERE tree_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	
	rows, err := db.Query(query, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var measurements []models.Measurement
	for rows.Next() {
		var m models.Measurement
		err := rows.Scan(
			&m.ID,
			&m.TreeID,
			&m.SoilMoisture,
			&m.Temperature,
			&m.Tilt,
			&m.BatteryLevel,
			&m.Timestamp,
			&m.ReceivedAt,
			&m.TokensAwarded,
		)
		if err != nil {
			return nil, err
		}
		measurements = append(measurements, m)
	}
	
	return measurements, nil
}

// UpdateUserTokenBalance adds tokens to a user's balance
func (db *DB) UpdateUserTokenBalance(userID, amount int) error {
	query := `UPDATE users SET token_balance = token_balance + $1 WHERE id = $2`
	_, err := db.Exec(query, amount, userID)
	return err
}

// InsertTokenTransaction records a token transaction
func (db *DB) InsertTokenTransaction(tx *models.TokenTransaction) error {
	query := `
		INSERT INTO token_transactions (user_id, measurement_id, amount, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	
	err := db.QueryRow(
		query,
		tx.UserID,
		tx.MeasurementID,
		tx.Amount,
		tx.Type,
	).Scan(&tx.ID, &tx.CreatedAt)
	
	return err
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, token_balance, created_at FROM users WHERE id = $1`
	
	err := db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.TokenBalance,
		&user.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// GetAllTrees retrieves all trees
func (db *DB) GetAllTrees() ([]models.Tree, error) {
	query := `SELECT id, name, location, sensor_id, user_id, created_at FROM trees ORDER BY id`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var trees []models.Tree
	for rows.Next() {
		var tree models.Tree
		err := rows.Scan(
			&tree.ID,
			&tree.Name,
			&tree.Location,
			&tree.SensorID,
			&tree.UserID,
			&tree.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}
	
	return trees, nil
}

// CheckDuplicateMeasurement checks if a measurement already exists
func (db *DB) CheckDuplicateMeasurement(treeID int, timestamp time.Time) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM measurements WHERE tree_id = $1 AND timestamp = $2)`
	
	err := db.QueryRow(query, treeID, timestamp).Scan(&exists)
	if err != nil {
		return false, err
	}
	
	return exists, nil
}