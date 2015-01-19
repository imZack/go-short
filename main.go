package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const DB_PATH string = "data.db"

var (
	CODE_CHAR = []rune("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	DB        *sql.DB
	PORT      int = 3000
)

func init() {
	rand.Seed(time.Now().Unix())
	if os.Getenv("PORT") != "" {
		var err error
		PORT, err = strconv.Atoi(os.Getenv("PORT"))
		check(err)
	}
}

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

func init_db() (err error) {
	DB, err = sql.Open(
		"postgres", os.Getenv("DATABASE_URL"))

	err = create_db(DB)

	if !check(err) {
		return
	}
	return
}

func create_db(db *sql.DB) (err error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS urls (
		id serial PRIMARY KEY,
		url TEXT NOT NULL,
		code varchar(5) default '' UNIQUE,
		hits INTEGER default 0,
		created_at timestamp without time zone default (now() at time zone 'utc')
	);
	`
	_, err = db.Exec(sqlStmt)
	check(err)

	return
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
		"SELECT id, url, code, hits, created_at FROM urls WHERE code = $1")
	if !check(err) {
		return
	}

	if err = stmt.QueryRow(code).Scan(
		&url.Id, &url.Url, &url.Code, &url.Hits, &url.Created_at); err != nil {
		log.Println(err)
		err = errors.New("Not found")
	}
	return
}

func inc(url *Url) (hits int64, err error) {
	url.Hits += 1
	var stmt *sql.Stmt
	stmt, err = DB.Prepare("UPDATE urls SET hits = $1 WHERE id= $2")
	if !check(err) {
		return
	}

	_, err = stmt.Exec(url.Hits, url.Id)
	if !check(err) {
		return
	}

	hits = url.Hits
	return
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

		stmt, err := tx.Prepare(
			"INSERT INTO urls(url, code) VALUES($1, $2) RETURNING id")
		if !check(err) {
			break
		}
		defer stmt.Close()

		err = stmt.QueryRow(data.Url, data.Code).Scan(&data.Id)
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
	r.Run(fmt.Sprintf(":%d", PORT))
}
