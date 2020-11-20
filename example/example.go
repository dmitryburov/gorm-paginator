package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dmitryburov/gorm-paginator"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	// see https://gorm.io/docs/connecting_to_the_database.html
	dbConn = "root:123456@tcp(localhost:3306)/test?parseTime=True"
)

type Book struct {
	gorm.Model
	Title string
}

func main() {

	var err error

	if err = initDatabase(); err != nil {
		log.Fatal("Error init database: ", err)
	}

	http.HandleFunc("/", getBookList)
	err = http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatal("Error serve: ", err)
	}

}

func getBookList(w http.ResponseWriter, r *http.Request) {

	var (
		err      error
		query    = r.URL.Query()
		dbEntity = db
		paging   = paginator.Paging{}
		bookList = struct {
			Items      []*Book               `json:"items"`
			Pagination *paginator.Pagination `json:"pagination"`
		}{}
	)

	// get paging params from query
	if len(query.Get("page")) > 0 && query.Get("page") != "" {
		paging.Page, _ = strconv.Atoi(query.Get("page"))
	}
	if len(query.Get("limit")) > 0 && query.Get("limit") != "" {
		paging.Limit, _ = strconv.Atoi(query.Get("limit"))
	}
	if orders, ok := query["order"]; ok || len(orders) > 0 {
		for i := range orders {
			paging.OrderBy = append(paging.OrderBy, orders[i])
		}
	}
	// show sql log if debug
	// paging.ShowSQL = true

	// if need conditions or more
	dbEntity = dbEntity.Where("id = ?", 1)

	// get data with pagination
	bookList.Pagination, err = paginator.Pages(&paginator.Param{
		DB:     dbEntity,
		Paging: &paging,
	}, &bookList.Items)
	if err != nil {
		log.Fatal("Error get list: ", err.Error())
	}

	// if empty list
	//if bookList.Pagination.IsEmpty() {
	//
	//}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(bookList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func initDatabase() (err error) {

	db, err = gorm.Open(mysql.Open(dbConn), &gorm.Config{})
	if err == nil {

		if !db.Migrator().HasTable(&Book{}) {
			fmt.Println("Start migrate")

			err = db.Migrator().CreateTable(&Book{})
			if err != nil {
				return
			}

			books := []Book{
				{Title: "Green mile"},
				{Title: "The Hobbit"},
				{Title: "The Da Vinci Code"},
				{Title: "Angels & Demons"},
				{Title: "Harry Potter"},
				{Title: "The Ginger Man"},
				{Title: "Cosmos"},
				{Title: "Angels & Demons"},
				{Title: "The Godfather"},
				{Title: "Dune"},
			}

			err = db.Create(&books).Error
			if err != nil {
				return
			}

			fmt.Println("Migrate success!")
		}
	}

	return
}
