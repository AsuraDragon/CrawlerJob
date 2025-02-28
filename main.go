package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	// "github.com/securisec/go-keywords"
	"os"
)

func main() {

	// connect to the database

	var db *sql.DB

	db, err := connectDB()

	if err != nil {
		fmt.Println(err)
		fmt.Println("crawler: could not connect to database, terminating")
		os.Exit(-1)
	}

	defer db.Close()
	defer fmt.Println("crawler: (deferred) database connection closed")

	// get a cursor to go through the table
	
	rows, err := db.Query("SELECT * FROM web_article")
    if err != nil {
		fmt.Println("crawler: query with SELECT failed")
		os.Exit(-1)
    }

	defer rows.Close()
	defer println("crawler: (deferred) database row cursor closed")
	// Loop through rows, using Scan to assign column data to struct fields.

	for rows.Next() {

		fmt.Println("------------------------------------------")
		
		var article Article
		if err := rows.Scan(&article.Id, &article.Title, &article.Url, &article.Keywords); err != nil {
			fmt.Println("crawler: parsing article from rows failed, skipping")
			continue
		}

		fmt.Printf("crawler: Article ID: %v Title: %v\n", article.Id, article.Title)
		fmt.Printf("crawler: URL: %v\n\n", article.Url)

		// process each article here

	}

	fmt.Println("------------------------------------------")
}

func connectDB () (*sql.DB, error) {

	var db *sql.DB

	user := os.Getenv("DBUSER")
	if user == "" {
		err := fmt.Errorf("crawler: connectDB(): environmental variable DBUSER not set")
		return nil, err
	}

	password := os.Getenv("DBPASSWORD")
	if password == "" {
		err := fmt.Errorf("crawler: connectDB(): environmental variable DBPASSWORD not set")
		return nil, err
	}

	cfg := mysql.Config {
        User:   user,
        Passwd: password,
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "it_support",
    }

	// connect to the database
	var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
		fmt.Printf("crawler: connectDB(): failed to connect to database %v\n", cfg.DBName)
        return nil, err
    }

	// we can only be sure by pinging the database
    pingErr := db.Ping()
    if pingErr != nil {
		fmt.Printf("crawler: connectDB(): failed to ping database %v (check log in credentials)\n", cfg.DBName)
        return nil, pingErr
    }

    fmt.Printf("crawler: connectDB(): successfully connected to database %v\n", cfg.DBName)
	return db, nil
}
