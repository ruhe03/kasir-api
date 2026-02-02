package database

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func InitDB(connectionString string) (*sql.DB, error) {
	if connectionString == "" {
		log.Println("ERROR: Connection string is empty!")
		return nil, sql.ErrConnDone
	}
	
	log.Println("Attempting to connect to database...")
	log.Println("Connection string (partial):", connectionString[:30]+"...")
	
	// Replace .pooler.supabase.com dengan direct connection jika ada
	// Supabase direct connection lebih reliable
	if strings.Contains(connectionString, ".pooler.supabase.com") {
		log.Println("Note: Using pooler connection. If this fails, try direct connection instead.")
	}
	
	// Open database
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Println("Error opening database:", err)
		return nil, err
	}

	// Test connection
	log.Println("Testing database connection...")
	err = db.Ping()
	if err != nil {
		log.Println("Ping failed:", err)
		log.Println("Troubleshooting tips:")
		log.Println("1. Check if password is correct")
		log.Println("2. Verify database is not paused in Supabase")
		log.Println("3. Try using direct connection string (not pooler)")
		log.Println("4. Check firewall/network settings")
		return nil, err
	}

	// Set connection pool settings (optional tapi recommended)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Database connected successfully")
	return db, nil
}