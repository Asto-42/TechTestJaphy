package internal

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	charmLog "github.com/charmbracelet/log"
	"github.com/gorilla/mux"
)

type App struct {
	logger *charmLog.Logger
	DB     *sql.DB
}

func NewApp(logger *charmLog.Logger) *App {
	return &App{
		logger: logger,
	}
}

func (a *App) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/breeds", a.GetBreeds).Methods("GET")
	r.HandleFunc("/breeds", a.CreateBreed).Methods("POST")
	r.HandleFunc("/breeds/{id:[0-9]+}", a.UpdateBreed).Methods("PUT")
	r.HandleFunc("/breeds/{id:[0-9]+}", a.DeleteBreed).Methods("DELETE")
	r.HandleFunc("/breeds/search", a.SearchBreeds).Methods("GET")
}

type Breed struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Species       string  `json:"species"`
	AverageWeight float64 `json:"average_weight"`
}

func (a *App) GetBreeds(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(`
    SELECT 
        id, 
        name, 
        species, 
        (weight_min + weight_max) / 2 AS average_weight 
    FROM breeds
`)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "Failed to fetch breeds", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var breeds []Breed
	for rows.Next() {
		var breed Breed
		rows.Scan(&breed.ID, &breed.Name, &breed.Species, &breed.AverageWeight)
		breeds = append(breeds, breed)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(breeds)
}

func (a *App) CreateBreed(w http.ResponseWriter, r *http.Request) {
	var breed Breed
	if err := json.NewDecoder(r.Body).Decode(&breed); err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := a.DB.Exec(`
    INSERT INTO breeds (name, species, pet_size, weight_min, weight_max) 
    VALUES (?, ?, ?, ?, ?)`,
    breed.Name, breed.Species, "Unknown", breed.AverageWeight-1, breed.AverageWeight+1)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "Failed to create breed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *App) UpdateBreed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var breed Breed
	if err := json.NewDecoder(r.Body).Decode(&breed); err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := a.DB.Exec(`
    UPDATE breeds 
    SET name = ?, species = ?, weight_min = ?, weight_max = ? 
    WHERE id = ?`,
    breed.Name, breed.Species, breed.AverageWeight-1, breed.AverageWeight+1, id)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "Failed to update breed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *App) DeleteBreed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := a.DB.Exec("DELETE FROM breeds WHERE id = ?", id)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "Failed to delete breed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) SearchBreeds(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	species := queryParams.Get("species")
	weight := queryParams.Get("weight")

	query := `
    SELECT 
        id, 
        name, 
        species, 
        (weight_min + weight_max) / 2 AS average_weight 
    FROM breeds 
    WHERE 1=1
`
	args := []interface{}{}

	if species != "" {
		query += " AND species = ?"
		args = append(args, species)
	}

	if weight != "" {
		weightVal, err := strconv.ParseFloat(weight, 64)
		if err == nil {
			query += " AND average_weight <= ?"
			args = append(args, weightVal)
		}
	}

	rows, err := a.DB.Query(query, args...)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "Failed to search breeds", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var breeds []Breed
	for rows.Next() {
		var breed Breed
		rows.Scan(&breed.ID, &breed.Name, &breed.Species, &breed.AverageWeight)
		breeds = append(breeds, breed)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(breeds)
}