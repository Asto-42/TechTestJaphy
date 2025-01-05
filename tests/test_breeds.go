package tests

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"bytes"
	"time"
)

type Breed struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Species       string  `json:"species"`
	AverageWeight float64 `json:"average_weight"`
}

func WaitForServer(apiURL string) {
	const maxRetries = 10
	const retryDelay = time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/health", apiURL))
		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Println("✅ Serveur prêt.")
			return
		}
		fmt.Printf("⏳ En attente du serveur... Tentative %d/%d\n", i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	fmt.Println("❌ Le serveur n'est pas prêt après plusieurs tentatives.")
	os.Exit(1)
}

func StartTests() {
	fmt.Println("\n\n=== Début des tests de l'API Breeds ===")

	csvFile := "./breeds.csv"
	apiURL := "http://127.0.0.1:5000/v1/breeds"
	
	fmt.Println("🔍 Lecture des données du fichier CSV...")
	file, err := os.Open(csvFile)
	if err != nil {
		fmt.Printf("❌ Erreur lors de l'ouverture du fichier CSV : %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	expectedBreeds := []Breed{}
	_, err = csvReader.Read()
	if err != nil {
		fmt.Printf("❌ Erreur lors de la lecture de l'en-tête CSV : %s\n", err)
		os.Exit(1)
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("❌ Erreur lors de la lecture du fichier CSV : %s\n", err)
			os.Exit(1)
		}

		weightMin, _ := strconv.Atoi(record[4])
		weightMax, _ := strconv.Atoi(record[5])
		averageWeight := float64(weightMin+weightMax) / 2

		expectedBreeds = append(expectedBreeds, Breed{
			Name:          record[3],
			Species:       record[1],
			AverageWeight: averageWeight,
		})
	}
	fmt.Println("✅ Lecture du fichier CSV réussie.")
	fmt.Println("🔍 Envoi de la requête à l'API...")
	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("❌ Erreur lors de la requête à l'API : %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ L'API a retourné un code HTTP inattendu : %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var apiBreeds []Breed
	if err := json.NewDecoder(resp.Body).Decode(&apiBreeds); err != nil {
		fmt.Printf("❌ Erreur lors du décodage JSON de la réponse API : %s\n", err)
		os.Exit(1)
	}
	fmt.Println("🔍 Comparaison des données entre CSV et API...")
	if len(expectedBreeds) != len(apiBreeds) {
		fmt.Printf("❌ Nombre de races différentes. CSV : %d, API : %d\n", len(expectedBreeds), len(apiBreeds))
		os.Exit(1)
	}

	for i, expected := range expectedBreeds {
		api := apiBreeds[i]
		if expected.Name != api.Name || expected.Species != api.Species || expected.AverageWeight != api.AverageWeight {
			fmt.Printf("❌ Mismatch à l'index %d : attendu %+v, reçu %+v\n", i, expected, api)
			os.Exit(1)
		}
	}
	fmt.Println("✅ Toutes les données correspondent entre le CSV et l'API.")
	fmt.Println("=== Tests terminés avec succès ===")
	testOtherEndpoint(apiURL)
}

func testOtherEndpoint(apiURL string) {

	fmt.Printf("🔍 URL de base pour l'API : %s\n", apiURL)

	fmt.Println("🔍 Test de POST /v1/breeds...")
	newBreed := Breed{
		Name:          "Test Breed",
		Species:       "Test Species",
		AverageWeight: 15.0,
	}
	newBreedID := testPost(apiURL, newBreed)
	
	fmt.Println("🔍 Validation de la création avec GET...")
	fmt.Printf("🔍 URL utilisée pour GET : %s/%d\n", apiURL, newBreedID)
	testGet(apiURL, newBreedID, newBreed)

	fmt.Println("🔍 Test de PUT /v1/breeds/{id}...")
	updatedBreed := Breed{
		Name:          "Updated Test Breed",
		Species:       "Updated Test Species",
		AverageWeight: 20.0,
	}
	testPut(apiURL, newBreedID, updatedBreed)

	fmt.Println("🔍 Validation de la mise à jour avec GET...")
	testGet(apiURL, newBreedID, updatedBreed)

	fmt.Println("🔍 Test de DELETE /v1/breeds/{id}...")
	testDelete(apiURL, newBreedID)

	fmt.Println("🔍 Validation de la suppression avec GET...")
	testGetDeleted(apiURL, newBreedID)

	fmt.Println("✅ Tous les tests CRUD ont réussi.")
}

func testPost(apiURL string, breed Breed) int {
    body, _ := json.Marshal(breed)
    resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
    if err != nil {
        fmt.Printf("❌ Erreur lors de POST : %s\n", err)
        os.Exit(1)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusCreated {
        fmt.Printf("❌ POST a retourné un code inattendu : %d\n", resp.StatusCode)
        os.Exit(1)
    }
    var createdBreed Breed
    err = json.NewDecoder(resp.Body).Decode(&createdBreed)
    if err != nil {
        fmt.Printf("❌ Erreur lors du décodage de la réponse POST : %s\n", err)
        os.Exit(1)
    }
    fmt.Printf("✅ POST réussi. ID créé : %d\n", createdBreed.ID)
    return createdBreed.ID
}


func testGet(apiURL string, id int, expected Breed) {
	finalURL := fmt.Sprintf("%s/%d", apiURL, id)
	fmt.Printf("🔍 Envoi de la requête GET à l'URL : %s\n", finalURL)

	resp, err := http.Get(finalURL)
	if err != nil {
		fmt.Printf("❌ Erreur lors de GET : %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ GET a retourné un code inattendu : %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var breed Breed
	json.NewDecoder(resp.Body).Decode(&breed)

	if breed.Name != expected.Name || breed.Species != expected.Species || breed.AverageWeight != expected.AverageWeight {
		fmt.Printf("❌ GET : Données incorrectes. Attendu %+v, reçu %+v\n", expected, breed)
		os.Exit(1)
	}
	fmt.Println("✅ GET réussi.")
}


func testPut(apiURL string, id int, updated Breed) {
	body, _ := json.Marshal(updated)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", apiURL, id), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Erreur lors de PUT : %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ PUT a retourné un code inattendu : %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("✅ PUT réussi.")
}

func testDelete(apiURL string, id int) {
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", apiURL, id), nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Erreur lors de DELETE : %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		fmt.Printf("❌ DELETE a retourné un code inattendu : %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("✅ DELETE réussi.")
}

func testGetDeleted(apiURL string, id int) {
	resp, err := http.Get(fmt.Sprintf("%s/%d", apiURL, id))
	if err != nil {
		fmt.Printf("❌ Erreur lors de GET : %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		fmt.Printf("❌ GET après suppression a retourné un code inattendu : %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("✅ Validation de la suppression réussie.")
}
