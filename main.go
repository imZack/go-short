package main

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
	"os"
	"time"
)

const DB_PATH string = "data.db"

var (
	CODE_CHAR = []rune("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	DB        *sql.DB
)

func get_code(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = CODE_CHAR[rand.Intn(len(CODE_CHAR))]
	}
	return string(b)
}

type Url struct {
	Id         int64     `json:"id"`
	Url        string    `json:"url" binding:"required"`
	Code       string    `json:"code"`
	Hits       int64     `json:"hits"`
	Created_at time.Time `json:"created_at"`
}

func check(err error) bool {
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func init_db() {
	if _, err := os.Stat(DB_PATH); os.IsNotExist(err) {
		log.Printf("no such file or directory: %s\n", DB_PATH)
		log.Printf("create one for you.\n")
		create_db(DB, DB_PATH)
	}
	var err error
	DB, err = sql.Open("sqlite3", DB_PATH)
	check(err)
}

func create_db(db *sql.DB, db_path string) {

	os.Remove(db_path)

	var err error
	db, err = sql.Open("sqlite3", db_path)
	check(err)

	sqlStmt := `
	create table urls (
		id INTEGER not null primary key,
		url TEXT not null,
		code TEXT default '',
		hits INTEGER default 0,
		created_at DATETIME default current_timestamp);
	create index urls_code_index on urls(code);
	`
	_, err = db.Exec(sqlStmt)
	check(err)
}

func redirect(c *gin.Context) {
	url, err := get_url(c.Params.ByName("code"))
	if err != nil {
		c.Redirect(302, "/static/404.html")
		return
	}
	c.Redirect(301, url.Url)
}

func get(c *gin.Context) {
	url, err := get_url(c.Params.ByName("code"))
	if err != nil {
		if err.Error() == "Not found" {
			c.JSON(404, gin.H{"message": "URL not found!"})
			return
		}
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	inc(url)
	c.JSON(200, url)
}

func get_url(code string) (url *Url, err error) {

	url, err = &Url{}, nil
	var stmt *sql.Stmt

	stmt, err = DB.Prepare(
		"select id, url, code, hits, created_at from urls where code = ?")
	check(err)

	if err = stmt.QueryRow(code).Scan(
		&url.Id, &url.Url, &url.Code, &url.Hits, &url.Created_at); err != nil {
		log.Println(err)
		err = errors.New("Not found")
	}
	log.Println(url.Id)

	return
}

func inc(url *Url) (int64, error) {
	url.Hits += 1
	stmt, err := DB.Prepare("update urls set hits = ? where id= ?")
	check(err)

	_, err = stmt.Exec(url.Hits, url.Id)
	check(err)
	return url.Hits, err
}

func create(c *gin.Context) {
	var (
		data Url
		err  error = errors.New("Internal Error")
		code int   = 500
	)

	for {
		if !c.Bind(&data) {
			code, err = 400, errors.New("Needs url")
			break
		}

		if data.Code == "" {
			data.Code = get_code(5)
		}

		var url *Url
		url, err = get_url(data.Code)
		if (err != nil && err.Error() != "Not found") || url.Id != 0 {
			err = errors.New("Duplicate code!")
			break
		}

		tx, err := DB.Begin()
		if !check(err) {
			break
		}

		stmt, err := tx.Prepare("insert into urls(url, code) values(?, ?)")
		if !check(err) {
			break
		}
		defer stmt.Close()

		result, err := stmt.Exec(data.Url, data.Code)
		if !check(err) {
			break
		}

		data.Id, err = result.LastInsertId()
		if !check(err) {
			break
		}

		tx.Commit()

		c.JSON(200, data)
		return
	}

	c.JSON(code, gin.H{"message": err.Error()})

}

func main() {
	init_db()
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/urls", create)
		v1.GET("/urls/:code", get)
	}
	r.GET("/r/:code", redirect)
	r.Static("/static", "static")
	r.Run(":3001")
}
