package database_actions

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"database/sql"
)

func ImportBreeds(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Cannot open file %s: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("Cannot read CSV file: %w", err)
	}

	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) != 6 {
			return fmt.Errorf("invalid format line %d", i+1)
		}
		species := strings.TrimSpace(row[1])
		petSize := strings.TrimSpace(row[2])
		name := strings.TrimSpace(row[3])
		weightMin, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if err != nil {
			return fmt.Errorf("invalid weight_min at line %d: %w", i+1, err)
		}
		weightMax, err := strconv.ParseFloat(strings.TrimSpace(row[5]), 64)
		if err != nil {
			return fmt.Errorf("invalid weight_max at line %d: %w", i+1, err)
		}
		_, err = db.Exec(
			"INSERT INTO breeds (species, pet_size, name, weight_min, weight_max) VALUES (?, ?, ?, ?, ?)",
			species, petSize, name, weightMin, weightMax,
		)
		if err != nil {
			return fmt.Errorf("failed to insert record at line %d: %w", i+1, err)
		}
	}
	return nil
}
