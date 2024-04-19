package postgres

import (
	"database/sql"
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

    return &Postgres{Db: db}, nil
}



// CheckTableExists checks if a table exists in the database
// func (p *Postgres) CheckTableExists(tableName string) (bool, error) {
// 	var exists bool
// 	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s')", tableName)
// 	err := p.Db.QueryRow(query).Scan(&exists)
// 	if err != nil {
// 		return false, err
// 	}
// 	return exists, nil
// }

// // MigrateTable migrates the table if it doesn't exist
// func (p *Postgres) MigrateTable(tableName string) error {
// 	exists, err := p.CheckTableExists(tableName)
// 	if err != nil {
// 		return err
// 	}
// 	if !exists {
// 		_, err := p.Db.Exec(migrationQuery(tableName))
// 		if err != nil {
// 			return err
// 		}
// 		fmt.Printf("Table %s migrated successfully\n", tableName)
// 	}
// 	return nil
// }


// func migrationQuery(tableName string) string {
//     switch tableName {
//     case "users":
//         return `CREATE TABLE IF NOT EXISTS users (
//             id SERIAL PRIMARY KEY,
//             first_name TEXT NOT NULL,
//             last_name TEXT NOT NULL,
//             email TEXT NOT NULL UNIQUE,
//             password TEXT NOT NULL,
//             role TEXT NOT NULL,
//             user_image_path TEXT,
//             created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
//         );`
//     // Add cases for other table names here if needed
//     default:
//         return ""
//     }
// }


