package gocql

import "reflect"

type rowData struct {
	Columns []string
	Values []interface{}
}

// New creates a pointer to an empty version of whatever type 
// is referenced by the TypeInfo receiver
func (t *TypeInfo) New() interface{} {
	return reflect.New(goType(t)).Interface()
}

func goType(t *TypeInfo) reflect.Type {
	switch t.Type {
	case TypeVarchar, TypeAscii:
		return reflect.TypeOf(*new(string))
	case TypeBigInt, TypeCounter, TypeTimestamp:
		return reflect.TypeOf(*new(int64))
	case TypeBlob:
		return reflect.TypeOf(*new([]byte))
	case TypeBoolean:
		return reflect.TypeOf(*new(bool))
	case TypeFloat:
		return reflect.TypeOf(*new(float32))
	case TypeDouble:
		return reflect.TypeOf(*new(float64))
	case TypeInt:
		return reflect.TypeOf(*new(int))
	case TypeUUID, TypeTimeUUID:
		return reflect.TypeOf(*new(UUID))
	case TypeList, TypeSet:
		return reflect.SliceOf(goType(t.Elem))
	case TypeMap:
		return reflect.MapOf(goType(t.Key), goType(t.Elem))
	default:
		return nil
	}
}

func dereference(i interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}

func (r *rowData) rowMap(m map[string]interface{}) {
	for i, column := range r.Columns {
		m[column] = dereference(r.Values[i])
	}
}

func (iter *Iter) rowData() (rowData, error) {
	if iter.err != nil {
		return rowData{}, iter.err
	}
	columns := make([]string, 0)
	values := make([]interface{}, 0)
	for _, column := range iter.Columns() {
		val := column.TypeInfo.New()
		columns = append(columns, column.Name)
		values = append(values, val)
	}
	rowData := rowData{
		Columns: columns,
		Values: values,
	}
	return rowData, nil
}

// SliceMap is a helper function to make the API easier to use
// returns the data from the query in the form of []map[string]interface{}
func (iter *Iter) SliceMap() ([]map[string]interface{}, error) {
	if iter.err != nil {
		return nil, iter.err
	}

	// Not checking for the error because we just did
	rowData, _ := iter.rowData()
	dataToReturn := make([]map[string]interface{}, 0)
	for iter.Scan(rowData.Values...) {
		m := make(map[string]interface{})
		rowData.rowMap(m)
		dataToReturn = append(dataToReturn, m)
	}
	if iter.err != nil {
		return nil, iter.err
	}
	return dataToReturn, nil
}

// MapScan takes a map[string]interface{} and populates it with a row
// That is returned from cassandra.
func (iter *Iter) MapScan(m map[string]interface{}) bool {
	if iter.err != nil {
		return false
	}

	// Not checking for the error because we just did
	rowData, _ := iter.rowData()

	if iter.Scan(rowData.Values...) {
		rowData.rowMap(m)	
		return true
	}
	return false
}
