package repositories

import (
	"database/sql"
	"kasir-api/models"
	"errors"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (repo *CategoryRepository) GetAll() ([]models.Category, error) {
	query := "SELECT id, name, description FROM category"
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Category, 0)
	for rows.Next() {
		var p models.Category
		err := rows.Scan(&p.ID, &p.Name, &p.Description)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (repo *CategoryRepository) Create(category *models.Category) error {
	query := "INSERT INTO category (name, description) VALUES ($1, $2) RETURNING id"
	err := repo.db.QueryRow(query, category.Name, category.Description).Scan(&category.ID)
	return err
}

// GetByID - ambil category by ID
func (repo *CategoryRepository) GetByID(id int) (*models.Category, error) {
	query := "SELECT id, name, description FROM category WHERE id = $1"

	var c models.Category
	err := repo.db.QueryRow(query, id).Scan(&c.ID, &c.Name, &c.Description)
	if err == sql.ErrNoRows {
		return nil, errors.New("category tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (repo *CategoryRepository) Update(category *models.Category) error {
	query := "UPDATE category SET name = $1, description = $2 WHERE id = $3"
	result, err := repo.db.Exec(query, category.Name, category.Description, category.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("category tidak ditemukan")
	}

	return nil
}

func (repo *CategoryRepository) Delete(id int) error {
	query := "DELETE FROM category WHERE id = $1"
	result, err := repo.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("category tidak ditemukan")
	}

	return err
}