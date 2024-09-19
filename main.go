package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

type SalesData struct {
    ID        int       `json:"id"`
    Date      time.Time `json:"date"`
    Amount    float64   `json:"amount"`
    ProductID int       `json:"product_id"`
    Region    string    `json:"region"`
}

type AggregatedData struct {
    ProductID  int     `json:"product_id"`
    TotalSales float64 `json:"total_sales"`
    Region     string  `json:"region"`
}

var db *sql.DB

func initDB() (*sql.DB, error) {
    connStr := "user=analyticsuser password=secret dbname=analyticsdb sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    return db, nil
}

func createTable(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS sales_data (
            id SERIAL PRIMARY KEY,
            date TIMESTAMPTZ,
            amount FLOAT,
            product_id INT,
            region TEXT
        )
    `)
    return err
}

func parseCSV(filePath string) ([]SalesData, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    var salesData []SalesData

    // Skip the header row
    _, err = reader.Read()
    if err != nil {
        return nil, err
    }

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        id, err := strconv.Atoi(record[0])
        if err != nil {
            return nil, err
        }

        date, err := time.Parse(time.RFC3339, record[1]+"T00:00:00Z")
        if err != nil {
            return nil, err
        }

        amount, err := strconv.ParseFloat(record[2], 64)
        if err != nil {
            return nil, err
        }

        productID, err := strconv.Atoi(record[3])
        if err != nil {
            return nil, err
        }

        salesData = append(salesData, SalesData{
            ID:        id,
            Date:      date,
            Amount:    amount,
            ProductID: productID,
            Region:    record[4],
        })
    }

    return salesData, nil
}

func insertSalesData(db *sql.DB, sale SalesData) error {
    _, err := db.Exec(`
        INSERT INTO sales_data (date, amount, product_id, region)
        VALUES ($1, $2, $3, $4)
    `, sale.Date, sale.Amount, sale.ProductID, sale.Region)
    return err
}

func aggregateData(db *sql.DB) ([]AggregatedData, error) {
    rows, err := db.Query(`
        SELECT product_id, SUM(amount) as total_sales, region
        FROM sales_data
        GROUP BY product_id, region
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var aggregated []AggregatedData
    for rows.Next() {
        var data AggregatedData
        err := rows.Scan(&data.ProductID, &data.TotalSales, &data.Region)
        if err != nil {
            return nil, err
        }
        aggregated = append(aggregated, data)
    }

    return aggregated, nil
}

func getSalesData(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query(`
        SELECT id, date, amount, product_id, region
        FROM sales_data
    `)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var sales []SalesData
    for rows.Next() {
        var data SalesData
        err := rows.Scan(&data.ID, &data.Date, &data.Amount, &data.ProductID, &data.Region)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        sales = append(sales, data)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(sales)
}

func getAggregatedData(w http.ResponseWriter, r *http.Request) {
    aggregated, err := aggregateData(db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(aggregated)
}

func addSalesData(w http.ResponseWriter, r *http.Request) {
    // Implementation for adding sales data
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func main() {
    var err error
    db, err = initDB()
    if err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }
    defer db.Close()

    err = createTable(db)
    if err != nil {
        log.Fatalf("Error creating table: %v", err)
    }

    sales, err := parseCSV("sales_data.csv")
    if err != nil {
        log.Fatalf("Error parsing CSV: %v", err)
    }
    log.Println("Ingested sales data:", sales)

    for _, sale := range sales {
        err = insertSalesData(db, sale)
        if err != nil {
            log.Fatalf("Error inserting sales data: %v", err)
        }
    }

    aggregated, err := aggregateData(db)
    if err != nil {
        log.Fatalf("Error aggregating data: %v", err)
    }
    log.Println("Aggregated data:", aggregated)

    http.HandleFunc("/sales", getSalesData)
    http.HandleFunc("/aggregated", getAggregatedData)
    http.HandleFunc("/add", addSalesData)
    http.HandleFunc("/health", healthCheck)

    // Serve static files
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/", fs)

    // Graceful shutdown
    server := &http.Server{Addr: ":8080"}

    go func() {
        log.Println("Server listening on port 8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Error starting server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    if err := server.Close(); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exiting")
}