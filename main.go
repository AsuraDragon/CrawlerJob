package main

import (
	"database/sql"
	"fmt"
	"html"
	"regexp"

	"github.com/go-sql-driver/mysql"

	// "github.com/securisec/go-keywords"
	"io"
	"net/http"
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

		// get the HTML contents
		htmlContent, err := getHtml(article.Url)
		if err != nil {
			fmt.Println("crawler: HTML content retrieval failed, skipping")
			continue
		}

		fmt.Printf("HTML: %v\n\n", htmlContent[:1000])

		titleToUse := extractTitle(htmlContent)
		if titleToUse != "" {
			// update the title column
			_, err = db.Exec("UPDATE web_article SET title = ? WHERE id = ?", titleToUse, article.Id)
			if err != nil {
				fmt.Printf("crawler: query to update the title failed for article id %v\n", article.Id)
			} else {
				fmt.Printf("crawler: query to update the title for article id %v was successful\n", article.Id)
			}
		}
	}

	fmt.Println("------------------------------------------")
}

func connectDB() (*sql.DB, error) {

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

	cfg := mysql.Config{
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

func getHtml(url string) (string, error) {

	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("crawler: GetHTML() cannot connect to %v", url)
	}
	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", fmt.Errorf("crawler: GetHtml() cannot close the responder")
	}

	return string(content), nil
}

func extractTitle(content string) string {

	searchPrefix := "<title>"
	searchSuffix := "</title>"
	searchString2 := "og:title\" content="

	titleTagRegEx := regexp.MustCompile(searchPrefix + ".*" + searchSuffix)
	ogtitleTagRegEx := regexp.MustCompile("og:title\" content=\"(.*?)\"")

	list1 := titleTagRegEx.FindStringSubmatch(content)
	fmt.Printf("crawler: extractTitle() <title> %v\n", list1)

	list2 := ogtitleTagRegEx.FindStringSubmatch(content)
	fmt.Printf("crawler: extractTitle() og:title %v\n", list2)

	// og:title meta tag is available
	if len(list2) > 0 {
		// use og:title
		if len(list2[0]) > len(searchString2) {
			// strip og:title" content=" from the first element
			title := list2[0][len(searchString2)+1 : len(list2[0])-1]
			title = html.UnescapeString(title)
			fmt.Printf("crawler: extractTitle() using og:title = %v\n", title)
			return title
		}
	}

	// use <title> HTML tags
	if len(list1) > 0 {
		// use <title>...</title>
		if len(list1[0]) < len(searchPrefix+searchSuffix) {
			fmt.Printf("crawler: extractTitle() no suitable title found\n")
			return ""
		}
		// extract content between <title> tags of the first element
		title := list1[0][len(searchPrefix) : len(list1[0])-len(searchSuffix)]

		title = html.UnescapeString(title)

		// exclude invalid titles
		if title == "Redirecting" || title == "Just a moment..." {
			fmt.Printf("crawler: extractTitle() no suitable title found\n")
			return ""
		}

		fmt.Printf("crawler: extractTitle() using <title> = %v\n", title)
		return title
	}

	fmt.Printf("crawler: extractTitle() no suitable title found\n")
	return ""

}
