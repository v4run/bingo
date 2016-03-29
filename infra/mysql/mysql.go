/*
Package mysql provides the library to communicate to mysql
*/
package mysql

import (
	_ "github.com/go-sql-driver/mysql" // Mysql driver
	"github.com/jmoiron/sqlx"
)

/*
Check http://jmoiron.github.io/sqlx/ for sqlx usage
*/

// Connect initializes mysql DB
func Connect(datasource string) *sqlx.DB {
	return sqlx.MustOpen("mysql", datasource)
}
