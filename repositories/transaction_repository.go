package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0)

	for _, item := range items {
		var productPrice, stock int
		var productName string

		err := tx.QueryRow("SELECT name, price, stock FROM products WHERE id = $1", item.ProductID).Scan(&productName, &productPrice, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	for i, detail := range details {
		details[i].TransactionID = transactionID
		_, err = tx.Exec("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4)",
			transactionID, detail.ProductID, detail.Quantity, detail.Subtotal)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}, nil
}

func (repo *TransactionRepository) GetDailyReport() (*models.DailyReport, error) {
	// 1. Get Total Revenue & Total Transaction
	// COALESCE handles NULL when no rows match
	queryStats := "SELECT COALESCE(SUM(total_amount), 0), COUNT(id) FROM transactions WHERE DATE(created_at) = CURRENT_DATE"

	var totalRevenue, totalTransactions int
	err := repo.db.QueryRow(queryStats).Scan(&totalRevenue, &totalTransactions)
	if err != nil {
		return nil, err
	}

	// 2. Get Best Selling Product
	queryBestSeller := `
		SELECT p.name, SUM(td.quantity) as total_qty
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE DATE(t.created_at) = CURRENT_DATE
		GROUP BY p.name
		ORDER BY total_qty DESC
		LIMIT 1
	`

	var bestProduct models.BestSellingProduct
	err = repo.db.QueryRow(queryBestSeller).Scan(&bestProduct.Nama, &bestProduct.QtyTerjual)

	if err == sql.ErrNoRows {
		// Valid case: no transactions today
		bestProduct = models.BestSellingProduct{
			Nama:       "-",
			QtyTerjual: 0,
		}
	} else if err != nil {
		return nil, err
	}

	return &models.DailyReport{
		TotalRevenue:   totalRevenue,
		TotalTransaksi: totalTransactions,
		ProdukTerlaris: bestProduct,
	}, nil
}
