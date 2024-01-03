package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"query-sqlc/repository"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DB_NAME     = "gocustomquerymysqlsqlc"
	DB_USERNAME = "root"
	DB_PASSWORD = "password"
	DB_HOST     = "localhost"
	DB_PORT     = 3306
)

var mySQlConn *sql.DB

func init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", DB_USERNAME, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal("MySQl db connection problem")
	}

	mySQlConn = new(sql.DB)
	mySQlConn = db
}

func main() {
	defer func() {
		err := mySQlConn.Close()
		if err != nil {
			log.Fatal("MySQl db connection close problem")
		}
	}()

	ctx := context.Background()
	repo := repository.NewRepository(mySQlConn)
	authors, err := repo.GetAllAuthor(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	data, _ := json.MarshalIndent(authors, "", "  ")
	fmt.Println(string(data))
}
