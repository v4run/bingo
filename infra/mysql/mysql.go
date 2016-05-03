/*
Package mysql provides the library to communicate to mysql
*/
package mysql

import (
	"fmt"

	"math"
	"time"

	"strconv"

	"database/sql"

	_ "github.com/go-sql-driver/mysql" // Mysql driver
	"github.com/hifx/errgo"
	"github.com/jmoiron/sqlx"
)

type sleepType int

const (
	consistentRetry sleepType = iota
	exponentialRetry
)

type RetryDB struct {
	*sqlx.DB
	retryTimes        int //maximum number of retry.
	retryWaitStrategy sleepType
	delayBetweenRetry time.Duration //sleep time before next retry.
	maxDelayCap       time.Duration //maximum duration of sleep time before next retry(Used in incremental sleep strategy: choosing minimum among calculated sleep value and maxDelayCap).
	queryTimeout      time.Duration //timeout of a query to finish in a single retry.
	retryUntil        time.Duration //the time in which query to
	retryFactor       int           //exponential factor for retry
}

func (rdb *RetryDB) sleep(attempt int) time.Duration {

	if rdb.retryWaitStrategy == consistentRetry {
		return rdb.delayBetweenRetry
	} else if rdb.retryWaitStrategy == exponentialRetry {
		return time.Duration(math.Min(float64(rdb.maxDelayCap), float64(float64(rdb.delayBetweenRetry)*math.Pow(float64(rdb.retryFactor), float64(attempt)))))
	} else {
		return time.Duration(math.Min(float64(rdb.maxDelayCap), float64(int64(attempt)*int64(rdb.delayBetweenRetry))))
	}
}

// Connect initializes mysql DB
func Connect(datasource string, maxactive, maxidle int) (*sqlx.DB, error) {
	db := sqlx.MustOpen("mysql", datasource)
	db.SetMaxOpenConns(maxactive)
	db.SetMaxIdleConns(maxidle)
	err := db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to mysql: %s err: %s", datasource, err)
	}
	return db, nil
}

func defaultRetryStrategy() RetryDB {
	return RetryDB{retryTimes: 3, retryWaitStrategy: consistentRetry, delayBetweenRetry: time.Duration(0), maxDelayCap: time.Duration(0), queryTimeout: time.Duration(2 * time.Second), retryUntil: time.Duration(3 * time.Second), retryFactor: 2}
}

// Connect initializes mysql DB
func ConnectWithRetry(datasource string, maxactive, maxidle int) (*RetryDB, error) {

	rdb := defaultRetryStrategy()
	var err error

	var retryFinishedInformerChanel chan error = make(chan error, 1)
	go func(rdb *RetryDB, ch chan error) {

		var err error
		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {
			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}
			var queryFinishedInformerChanel chan error = make(chan error, 1)
			go func(rdb *RetryDB, qch chan error) {
				//connecting to mysql
				db := sqlx.MustOpen("mysql", datasource)
				db.SetMaxOpenConns(maxactive)
				db.SetMaxIdleConns(maxidle)
				err := db.Ping()
				if err != nil {
					qch <- err
				} else {
					//connection successfully created
					(*rdb).DB = db
					qch <- nil
				}
			}(rdb, queryFinishedInformerChanel)

			select {
			case err = <-queryFinishedInformerChanel:
				//error in query execution
				// avoid if there is more retry chances is remaining.
				if err == nil {
					ch <- nil
					return
				}
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				err = errgo.New("Query time out error")

			}
		}
		ch <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + err.Error())
		return

	}(&rdb, retryFinishedInformerChanel)

	select {
	case err = <-retryFinishedInformerChanel:
		if err == nil {
			return &rdb, nil
		} else {
			return nil, err
		}
	case <-time.After(rdb.retryUntil):
		return nil, errgo.New("Retry timeout error: The allowed time of retryUntil is expired")
	}
}

func (rdb *RetryDB) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {

	if rdb == nil {
		//This function can be execute even rdb object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *sqlx.Rows = make(chan *sqlx.Rows, 1)
	go func(rdb *RetryDB, rrch chan *sqlx.Rows, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}

			var queryResultChanel chan *sqlx.Rows = make(chan *sqlx.Rows, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *sqlx.Rows

			go func(rdb *RetryDB, qrch chan *sqlx.Rows, qech chan *error) {
				//Calling sqlx library function
				r, qErr := rdb.DB.Queryx(query, args...)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(rdb, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rdb, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rdb.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}
}

func (rdb *RetryDB) NamedExec(query string, arg interface{}) (sql.Result, error) {

	if rdb == nil {
		//This function can be execute even rdb object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan sql.Result = make(chan sql.Result, 1)
	go func(rdb *RetryDB, rrch chan sql.Result, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}

			var queryResultChanel chan sql.Result = make(chan sql.Result, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r sql.Result

			go func(rdb *RetryDB, qrch chan sql.Result, qech chan *error) {
				//Calling sqlx library function
				r, qErr := rdb.DB.NamedExec(query, arg)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(rdb, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rdb, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rdb.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}

}

func (rdb *RetryDB) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {

	if rdb == nil {
		//This function can be execute even rdb object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *sqlx.Rows = make(chan *sqlx.Rows, 1)
	go func(rdb *RetryDB, rrch chan *sqlx.Rows, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}

			var queryResultChanel chan *sqlx.Rows = make(chan *sqlx.Rows, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *sqlx.Rows

			go func(rdb *RetryDB, qrch chan *sqlx.Rows, qech chan *error) {
				//Calling sqlx library function
				r, qErr := rdb.DB.NamedQuery(query, arg)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(rdb, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rdb, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rdb.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}

}

func (rdb *RetryDB) Query(query string, args ...interface{}) (*sql.Rows, error) {

	if rdb == nil {
		//This function can be execute even rdb object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *sql.Rows = make(chan *sql.Rows, 1)
	go func(rdb *RetryDB, rrch chan *sql.Rows, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}

			var queryResultChanel chan *sql.Rows = make(chan *sql.Rows, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *sql.Rows

			go func(rdb *RetryDB, qrch chan *sql.Rows, qech chan *error) {
				//Calling sqlx library function
				r, qErr := rdb.DB.Query(query, args...)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(rdb, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rdb, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rdb.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}

}

func (rdb *RetryDB) Exec(query string, args ...interface{}) (sql.Result, error) {

	if rdb == nil {
		//This function can be execute even rdb object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan sql.Result = make(chan sql.Result, 1)
	go func(rdb *RetryDB, rrch chan sql.Result, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}

			var queryResultChanel chan sql.Result = make(chan sql.Result, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r sql.Result

			go func(rdb *RetryDB, qrch chan sql.Result, qech chan *error) {
				//Calling sqlx library function
				r, qErr := rdb.DB.Exec(query, args...)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(rdb, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rdb, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rdb.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}

}

func (rdb *RetryDB) Get(dest interface{}, query string, args ...interface{}) error {

	var err error

	var retryFinishedInformerChanel chan error = make(chan error, 1)
	go func(dest interface{}, ch chan error) {

		var err error
		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {
			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}
			var queryFinishedInformerChanel chan error = make(chan error, 1)
			go func(dest interface{}, qch chan error) {
				//connecting to mysql

				err := rdb.DB.Get(dest, query, args...)
				if err != nil {
					qch <- err
				} else {
					//connection successfully created
					qch <- nil
				}
			}(dest, queryFinishedInformerChanel)

			select {
			case err = <-queryFinishedInformerChanel:
				//error in query execution
				// avoid if there is more retry chances is remaining.
				if err == nil {
					ch <- nil
					return
				}
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				err = errgo.New("Query time out error")

			}
		}
		ch <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + err.Error())
		return

	}(dest, retryFinishedInformerChanel)

	select {
	case err = <-retryFinishedInformerChanel:
		if err == nil {
			return nil
		} else {
			return err
		}
	case <-time.After(rdb.retryUntil):
		return errgo.New("Retry timeout error: The allowed time of retryUntil is expired")
	}
}

func (rdb *RetryDB) Select(dest interface{}, query string, args ...interface{}) error {

	var err error

	var retryFinishedInformerChanel chan error = make(chan error, 1)
	go func(dest interface{}, ch chan error) {

		var err error
		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {
			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}
			var queryFinishedInformerChanel chan error = make(chan error, 1)
			go func(dest interface{}, qch chan error) {
				//connecting to mysql

				err := rdb.DB.Select(dest, query, args...)
				//todo
				fmt.Println("-------------------")
				fmt.Println(dest)
				if err != nil {
					qch <- err
				} else {
					//connection successfully created
					qch <- nil
				}
			}(dest, queryFinishedInformerChanel)

			select {
			case err = <-queryFinishedInformerChanel:
				//error in query execution
				// avoid if there is more retry chances is remaining.
				if err == nil {
					ch <- nil
					return
				}
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				err = errgo.New("Query time out error")

			}
		}
		ch <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + err.Error())
		return

	}(dest, retryFinishedInformerChanel)

	select {
	case err = <-retryFinishedInformerChanel:
		if err == nil {
			return nil
		} else {
			return err
		}
	case <-time.After(rdb.retryUntil):
		return errgo.New("Retry timeout error: The allowed time of retryUntil is expired")
	}
}

func (rdb *RetryDB) PrepareNamed(query string) (*RetryNamedStmt, error) {

	if rdb == nil {
		//This function can be execute even rdb object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *RetryNamedStmt = make(chan *RetryNamedStmt, 1)
	go func(rdb *RetryDB, rrch chan *RetryNamedStmt, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}

			var queryResultChanel chan *RetryNamedStmt = make(chan *RetryNamedStmt, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *RetryNamedStmt

			go func(rdb *RetryDB, qrch chan *RetryNamedStmt, qech chan *error) {
				//Calling sqlx library function
				r, qErr := rdb.DB.PrepareNamed(query)

				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					stmt := RetryNamedStmt{Stmt: r}

					stmt.retryTimes = rdb.retryTimes
					stmt.retryWaitStrategy = rdb.retryWaitStrategy
					stmt.delayBetweenRetry = rdb.delayBetweenRetry
					stmt.maxDelayCap = rdb.maxDelayCap
					stmt.queryTimeout = rdb.queryTimeout
					stmt.retryUntil = rdb.retryUntil
					stmt.retryFactor = rdb.retryFactor

					qrch <- &stmt
				}
			}(rdb, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rdb, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rdb.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}

}

func (rdb *RetryDB) Preparex(query string) (*RetryStmt, error) {

	if rdb == nil {
		//This function can be execute even rdb object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *RetryStmt = make(chan *RetryStmt, 1)
	go func(rdb *RetryDB, rrch chan *RetryStmt, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*rdb).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*rdb).sleep(iRetry)
			}

			var queryResultChanel chan *RetryStmt = make(chan *RetryStmt, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *RetryStmt

			go func(rdb *RetryDB, qrch chan *RetryStmt, qech chan *error) {
				//Calling sqlx library function
				r, qErr := rdb.DB.Preparex(query)

				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					stmt := RetryStmt{Stmt: r}

					stmt.retryTimes = rdb.retryTimes
					stmt.retryWaitStrategy = rdb.retryWaitStrategy
					stmt.delayBetweenRetry = rdb.delayBetweenRetry
					stmt.maxDelayCap = rdb.maxDelayCap
					stmt.queryTimeout = rdb.queryTimeout
					stmt.retryUntil = rdb.retryUntil
					stmt.retryFactor = rdb.retryFactor

					qrch <- &stmt
				}
			}(rdb, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(rdb.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(rdb.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rdb, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rdb.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}

}

type RetryNamedStmt struct {
	Stmt              *sqlx.NamedStmt
	retryTimes        int //maximum number of retry.
	retryWaitStrategy sleepType
	delayBetweenRetry time.Duration //sleep time before next retry.
	maxDelayCap       time.Duration //maximum duration of sleep time before next retry(Used in incremental sleep strategy: choosing minimum among calculated sleep value and maxDelayCap).
	queryTimeout      time.Duration //timeout of a query to finish in a single retry.
	retryUntil        time.Duration //the time in which query to
	retryFactor       int           //exponential factor for retry
}

func (n *RetryNamedStmt) sleep(attempt int) time.Duration {

	if n.retryWaitStrategy == consistentRetry {
		return n.delayBetweenRetry
	} else if n.retryWaitStrategy == exponentialRetry {
		return time.Duration(math.Min(float64(n.maxDelayCap), float64(float64(n.delayBetweenRetry)*math.Pow(float64(n.retryFactor), float64(attempt)))))
	} else {
		return time.Duration(math.Min(float64(n.maxDelayCap), float64(int64(attempt)*int64(n.delayBetweenRetry))))
	}
}

func (n *RetryNamedStmt) Close() error {
	return n.Stmt.Close()
}

func (n *RetryNamedStmt) Exec(arg interface{}) (sql.Result, error) {

	if n == nil {
		//This function can be execute even s object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan sql.Result = make(chan sql.Result, 1)
	go func(s *RetryNamedStmt, rrch chan sql.Result, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*s).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*s).sleep(iRetry)
			}

			var queryResultChanel chan sql.Result = make(chan sql.Result, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r sql.Result

			go func(s *RetryNamedStmt, qrch chan sql.Result, qech chan *error) {
				//Calling sqlx library function
				r, qErr := s.Stmt.Exec(arg)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(s, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(s.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(s.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(n, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(n.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}
}

func (n *RetryNamedStmt) MustExec(arg interface{}) sql.Result {
	return n.Stmt.MustExec(arg)
}

func (n *RetryNamedStmt) QueryRow(arg interface{}) *sqlx.Row {
	return n.Stmt.QueryRow(arg)
}

func (n *RetryNamedStmt) QueryRowx(arg interface{}) *sqlx.Row {
	return n.Stmt.QueryRowx(arg)
}

func (n *RetryNamedStmt) Query(args interface{}) (*sql.Rows, error) {
	if n == nil {
		//This function can be execute even s object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *sql.Rows = make(chan *sql.Rows, 1)
	go func(s *RetryNamedStmt, rrch chan *sql.Rows, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*s).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*s).sleep(iRetry)
			}

			var queryResultChanel chan *sql.Rows = make(chan *sql.Rows, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *sql.Rows

			go func(s *RetryNamedStmt, qrch chan *sql.Rows, qech chan *error) {
				//Calling sqlx library function
				r, qErr := s.Stmt.Query(args)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(s, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(s.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(s.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(n, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(n.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}
}

func (n *RetryNamedStmt) Queryx(args interface{}) (*sqlx.Rows, error) {
	if n == nil {
		//This function can be execute even s object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *sqlx.Rows = make(chan *sqlx.Rows, 1)
	go func(s *RetryNamedStmt, rrch chan *sqlx.Rows, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*s).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*s).sleep(iRetry)
			}

			var queryResultChanel chan *sqlx.Rows = make(chan *sqlx.Rows, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *sqlx.Rows

			go func(s *RetryNamedStmt, qrch chan *sqlx.Rows, qech chan *error) {
				//Calling sqlx library function
				r, qErr := s.Stmt.Queryx(args)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(s, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(s.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(s.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(n, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(n.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}
}

func (n *RetryNamedStmt) Select(dest interface{}, arg interface{}) error {
	//todo add retry mechanism
	return n.Stmt.Select(dest, arg)
}

// Get using this NamedStmt
func (n *RetryNamedStmt) Get(dest interface{}, arg interface{}) error {
	//todo add retry mechanism
	return n.Stmt.Get(dest, arg)
}

func (n *RetryNamedStmt) Unsafe() *RetryNamedStmt {
	r := n.Stmt.Unsafe()
	rr := &RetryNamedStmt{Stmt: r, retryTimes: n.retryTimes, retryWaitStrategy: n.retryWaitStrategy, delayBetweenRetry: n.delayBetweenRetry, maxDelayCap: n.maxDelayCap, queryTimeout: n.queryTimeout, retryUntil: n.retryUntil, retryFactor: n.retryFactor}
	return rr
}

type RetryStmt struct {
	*sqlx.Stmt
	retryTimes        int //maximum number of retry.
	retryWaitStrategy sleepType
	delayBetweenRetry time.Duration //sleep time before next retry.
	maxDelayCap       time.Duration //maximum duration of sleep time before next retry(Used in incremental sleep strategy: choosing minimum among calculated sleep value and maxDelayCap).
	queryTimeout      time.Duration //timeout of a query to finish in a single retry.
	retryUntil        time.Duration //the time in which query to
	retryFactor       int           //exponential factor for retry
}

func (rs *RetryStmt) Unsafe() *RetryStmt {
	r := rs.Stmt.Unsafe()
	rr := &RetryStmt{Stmt: r, retryTimes: rs.retryTimes, retryWaitStrategy: rs.retryWaitStrategy, delayBetweenRetry: rs.delayBetweenRetry, maxDelayCap: rs.maxDelayCap, queryTimeout: rs.queryTimeout, retryUntil: rs.retryUntil, retryFactor: rs.retryFactor}
	return rr
}

func (rs *RetryStmt) sleep(attempt int) time.Duration {

	if rs.retryWaitStrategy == consistentRetry {
		return rs.delayBetweenRetry
	} else if rs.retryWaitStrategy == exponentialRetry {
		return time.Duration(math.Min(float64(rs.maxDelayCap), float64(float64(rs.delayBetweenRetry)*math.Pow(float64(rs.retryFactor), float64(attempt)))))
	} else {
		return time.Duration(math.Min(float64(rs.maxDelayCap), float64(int64(attempt)*int64(rs.delayBetweenRetry))))
	}
}

func (rs *RetryStmt) Query(args ...interface{}) (*sql.Rows, error) {
	if rs == nil {
		//This function can be execute even s object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan *sql.Rows = make(chan *sql.Rows, 1)
	go func(s *RetryStmt, rrch chan *sql.Rows, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*s).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*s).sleep(iRetry)
			}

			var queryResultChanel chan *sql.Rows = make(chan *sql.Rows, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r *sql.Rows

			go func(s *RetryStmt, qrch chan *sql.Rows, qech chan *error) {
				//Calling sqlx library function
				r, qErr := s.Stmt.Query(args...)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(s, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(s.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(s.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rs, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rs.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}
}

func (rs *RetryStmt) Select(dest interface{}, args ...interface{}) error {
	//todo add retry mechanism
	return rs.Stmt.Select(dest, args...)
}

// Get using the prepared statement.
func (rs *RetryStmt) Get(dest interface{}, args ...interface{}) error {
	//todo add retry mechanism
	return rs.Stmt.Get(dest, args...)
}

func (rs *RetryStmt) Exec(arg interface{}) (sql.Result, error) {

	if rs == nil {
		//This function can be execute even s object is nil.
		return nil, errgo.New("Connection error: Please connect to the database")
	}

	var lastErr error
	var retryErrorChanel chan error = make(chan error, 1)
	var retryResultChanel chan sql.Result = make(chan sql.Result, 1)
	go func(s *RetryStmt, rrch chan sql.Result, rech chan error, lastErr *error) {

		for iRetry := 0; iRetry < (*s).retryTimes; iRetry++ {

			if iRetry != 0 {
				//Sleeping before every iteration except first
				(*s).sleep(iRetry)
			}

			var queryResultChanel chan sql.Result = make(chan sql.Result, 1)
			var queryErrorChanel chan *error = make(chan *error, 1)
			var r sql.Result

			go func(s *RetryStmt, qrch chan sql.Result, qech chan *error) {
				//Calling sqlx library function
				r, qErr := s.Stmt.Exec(arg)
				if qErr != nil {
					qech <- &qErr
				} else {
					//sqlx library function executed successfully
					qrch <- r
				}
			}(s, queryResultChanel, queryErrorChanel)

			select {
			case r = <-queryResultChanel:
				rrch <- r
				return
			case lastErr = <-queryErrorChanel:
			//error in query execution
			//go for next iteration
			case <-time.After(s.queryTimeout):
				//maximum time allowed for executing query is up.
				//Go for the next retry is retryTimes is greater than attempt count.
				tmpEr := errgo.New("Query time out error")
				lastErr = &tmpEr
			}
		}
		rech <- errgo.New("Retry times(" + strconv.Itoa(s.retryTimes) + ") completed\nLast error message:" + (*lastErr).Error())

	}(rs, retryResultChanel, retryErrorChanel, &lastErr)

	select {
	case r := <-retryResultChanel:
		return r, nil
	case er := <-retryErrorChanel:
		return nil, er
	case <-time.After(rs.retryUntil):
		if lastErr == nil {
			return nil, errgo.New("Retry timeout error: The allowed time")
		}
		return nil, errgo.New("Retry timeout error: The allowed time\nLast error message:" + lastErr.Error())
	}
}

//TODO add wrapper to sql.TX
