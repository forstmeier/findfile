package csql

// CSQLer defines the methods for manipulating and interacting
// with CSQL and Elasticsearch queries.
type CSQLer interface {
	CSQLToES(csqlQuery map[string]interface{}) (map[string]interface{}, error)
}
