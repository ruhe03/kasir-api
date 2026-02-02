package main

import (
	"encoding/json"
	"fmt"
	"kasir-api/database"
	// "kasir-api/models"
	"kasir-api/repositories"
	"kasir-api/services"
	"kasir-api/handlers"
	"net/http"
	"os"
	"strconv"
	"strings"
	"log"

	"github.com/spf13/viper"
)

// Category represents a product in the cashier system
type Category struct {
	ID    int          `json:"id"`
	Name  string 	   `json:"name"`
	Description string `json:"description"`
}

// In-memory storage (sementara, nanti ganti database)
var category = []Category{
	{ID: 1, Name: "Makanan Instan", Description: "Produk siap saji atau mudah dimasak"},
	{ID: 2, Name: "Minuman", Description: "Produk minuman kemasan"},
	{ID: 3, Name: "Bumbu Dapur", Description: "Bahan pelengkap dan penyedap masakan"},
}

func getCategoriesByID(w http.ResponseWriter, r *http.Request) {
	// Parse ID dari URL path
	// URL: /api/produk/123 -> ID = 123
	idStr := strings.TrimPrefix(r.URL.Path, "/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	// Cari category dengan ID tersebut
	for _, p := range category {
		if p.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p)
			return
		}
	}

	// Kalau tidak found
	http.Error(w, "Category belum ada", http.StatusNotFound)
}

func updateCategories(w http.ResponseWriter, r *http.Request) {
	// get id dari request
	idStr := strings.TrimPrefix(r.URL.Path, "/categories/")

	// ganti int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	// get data dari request
	var updateCategories Category
	err = json.NewDecoder(r.Body).Decode(&updateCategories)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// loop category, cari id, ganti sesuai data dari request
	for i := range category {
		if category[i].ID == id {
			updateCategories.ID = id
			category[i] = updateCategories

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updateCategories)
			return
		}
	}
	
	http.Error(w, "Category belum ada", http.StatusNotFound)
}

func deleteCategories(w http.ResponseWriter, r *http.Request) {
	// get id
	idStr := strings.TrimPrefix(r.URL.Path, "/categories/")
	
	// ganti id int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}
	
	// loop category cari ID, dapet index yang mau dihapus
	for i, p := range category {
		if p.ID == id {
			// bikin slice baru dengan data sebelum dan sesudah index
			category = append(category[:i], category[i+1:]...)
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "sukses delete",
			})
			return
		}
	}

	http.Error(w, "Category belum ada", http.StatusNotFound)
}

type Config struct {
	Port    string `mapstructure:"PORT"`
	DBConn  string `mapstructure:"DB_CONN"`
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	config := Config{
		Port: viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	// Setup database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// setup routes
	http.HandleFunc("/api/produk", productHandler.HandleProducts)
	http.HandleFunc("/api/produk/", productHandler.HandleProductsByID)

	// >>>>> CATEGORIES 

	// GET localhost:8080/categories{id}
	// PUT localhost:8080/categories{id}
	// DELETE localhost:8080/categories{id}
	http.HandleFunc("/categories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getCategoriesByID(w, r)
		} else if r.Method == "PUT" {
			updateCategories(w, r)
		} else if r.Method == "DELETE" {
			deleteCategories(w, r)
		}
	})

	// GET localhost:8080/categories
	// POST localhost:8080/categories
	http.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(category)
			
		} else if r.Method == "POST" {
			// baca data dari request
			var categoryBaru Category
			err := json.NewDecoder(r.Body).Decode(&categoryBaru)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			// masukkin data ke dalam variable produk
			categoryBaru.ID = len(category) + 1
			category = append(category, categoryBaru)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated) // 201
			json.NewEncoder(w).Encode(categoryBaru)
		}
	})

	// localhost:8080/health
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "API Running",
		})
	})

	fmt.Println("Server running di localhost:" + config.Port)
	
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
