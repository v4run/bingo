/*
Package mysql provides the library to communicate to mysql
*/
package mysql

import (
	"log"

	_ "github.com/go-sql-driver/mysql" // Mysql driver
	"github.com/jmoiron/sqlx"
)

/*
Check http://jmoiron.github.io/sqlx/ for sqlx usage
*/

// Connect initializes mysql DB
func Connect(datasource string, maxactive, maxidle int) *sqlx.DB {
	db := sqlx.MustOpen("mysql", datasource)
	db.SetMaxOpenConns(maxactive)
	db.SetMaxIdleConns(maxidle)
	err := db.Ping()
	if err != nil {
		log.Fatal("unable to connect to mysql: ", datasource, " err: ", err)
	}
	return db
}
