package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")

	return &DB{db}, nil
}

func (db *DB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		token_balance INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS trees (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		location VARCHAR(255),
		sensor_id VARCHAR(255) UNIQUE NOT NULL,
		user_id INTEGER REFERENCES users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS measurements (
		id SERIAL PRIMARY KEY,
		tree_id INTEGER REFERENCES trees(id),
		soil_moisture FLOAT NOT NULL,
		temperature FLOAT NOT NULL,
		tilt FLOAT NOT NULL,
		battery_level FLOAT NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		tokens_awarded INTEGER DEFAULT 0,
		UNIQUE(tree_id, timestamp)
	);

	CREATE TABLE IF NOT EXISTS token_transactions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		measurement_id INTEGER REFERENCES measurements(id),
		amount INTEGER NOT NULL,
		type VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_measurements_tree_id ON measurements(tree_id);
	CREATE INDEX IF NOT EXISTS idx_measurements_timestamp ON measurements(timestamp);
	CREATE INDEX IF NOT EXISTS idx_token_transactions_user_id ON token_transactions(user_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Database schema initialized")
	return nil
}