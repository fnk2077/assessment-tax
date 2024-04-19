package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func New() (*Postgres, error) {

    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
        return nil, err
    }
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    postgresInstance := &Postgres{Db: db}
    if err := postgresInstance.MigrateTable("deductions"); err != nil {
        log.Fatal(err)
        return nil, err
    }

    return postgresInstance, nil
}


func (p *Postgres) CheckTableExists(tableName string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s')", tableName)
	err := p.Db.QueryRow(query).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *Postgres) MigrateTable(tableName string) error {
	exists, err := p.CheckTableExists(tableName)
	if err != nil {
		return err
	}
	if !exists {
		_, err := p.Db.Exec(migrationQuery(tableName))
		if err != nil {
			return err
		}
		fmt.Printf("Table %s migrated successfully\n", tableName)
	}
	return nil
}


func migrationQuery(tableName string) string {
    switch tableName {
    case "deductions":
        return `CREATE TABLE IF NOT EXISTS deductions (
            id SERIAL PRIMARY KEY,
            personal_deduction FLOAT,
            k_receipt FLOAT
        );`
    default:
        return ""
    }
}


