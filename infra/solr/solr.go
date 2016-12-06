/*
Package solr provides the library to communicate to solr
*/
package solr

import (
	"fmt"

	"github.com/hifx/go-solr/solr"
)

// Connect initializes the solr connection object
// Note: this doesn't actually hold a connection, its just
// a container for the URL
func Connect(host string, core string) (*solr.SolrInterface, error) {
	si, _ := solr.NewSolrInterface(host, core)
	status, qtime, _ := si.Ping()
	if status != "OK" {
		return nil, fmt.Errorf("unable to connect to solr '%s/%s' status expected to be 'OK' but got '%s'", host, core, status)
	}
	if qtime < 0 {
		return nil, fmt.Errorf("unable to connect to solr '%s/%s' qtime expected to be larger than '-1' but got '%d'", host, core, qtime)
	}
	return si, nil
}
