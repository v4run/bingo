/*
Package solr provides the library to communicate to solr
*/
package solr

import (
	"fmt"

	"github.com/rtt/Go-Solr"
)

// Connect initializes the solr connection object
// Note: this doesn't actually hold a connection, its just
// a container for the URL
func Connect(host string, port int, core string) (*solr.Connection, error) {
	sc, err := solr.Init(host, port, core)

	if nil != err {
		return nil, err
	}

	status, qtime, _ := ping(sc.URL)

	if status != "OK" {
		return nil, fmt.Errorf("unable to connect to solr '%s/%s' status expected to be 'OK' but got '%s'", host, core, status)
	}

	if qtime < 0 {
		return nil, fmt.Errorf("unable to connect to solr '%s/%s' qtime expected to be larger than '-1' but got '%d'", host, core, qtime)
	}

	return sc, nil
}

// Return 'status' and QTime from solr, if everything is fine status should have value 'OK'
// QTime will have value -1 if can not determine
func ping(url string) (status string, qtime int, err error) {
	body, err := solr.HTTPGet(fmt.Sprintf("%s/admin/ping?wt=json", url))

	if err != nil {
		return "", -1, err
	}

	resp, err := solr.BytesToJSON(&body)

	if err != nil {
		return "", -1, err
	}

	result := (*resp).(map[string]interface{})

	status, ok := result["status"].(string)

	if ok == false {
		return "", -1, fmt.Errorf("Unexpected response returned")
	}

	header := result["responseHeader"].(map[string]interface{})

	if QTime, ok := header["QTime"].(float64); ok {
		qtime = int(QTime)
	} else {
		qtime = -1
	}

	return status, qtime, nil
}
