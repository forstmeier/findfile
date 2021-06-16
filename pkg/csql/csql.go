package csql

// CSQLer defines the method for converting CSQL queries
// into database implementation-compatible queries as
// a byte array.
type CSQLer interface {
	ConvertCSQL(csqlQuery map[string]interface{}) ([]byte, error)
}