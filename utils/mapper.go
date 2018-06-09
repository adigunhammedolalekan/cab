//    sql-row-mapper provides utitlies to map db rows to structs, maps and arrays
//
//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.
//
//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"fmt"
	"reflect"
)

// RowsScanner can scan a db row and iterate over db rows
type RowsScanner interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(destination ...interface{}) error
}

// Scanner populates destination values
// or returns an error
type Scanner interface {
	Scan(destination ...interface{}) error
}

// MapRowsToSliceOfSlices maps db rows to a slice of slices
func MapRowsToSliceOfSlices(scanner RowsScanner, Slices *[][]interface{}) error {
	defer scanner.Close()
	for scanner.Next() {
		columns, err := scanner.Columns()
		if err != nil {
			return err
		}
		sliceOfResults := make([]interface{}, len(columns))
		for i := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// create a new interface{} than is not nil
			sliceOfResults[i] = new(interface{})
		}
		err = scanner.Scan(sliceOfResults...)
		if err != nil {
			return err
		}
		row := []interface{}{}
		for index := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// convert the sliceOfResults[index] value back to interface{}
			var v interface{} = *(sliceOfResults[index].(*interface{}))
			if u, ok := v.([]uint8); ok {
				// v is likely a string, so convert to string
				row = append(row, (interface{}(string(u))))
			} else {
				row = append(row, v)
			}
		}
		*Slices = append(*Slices, row)
	}
	return scanner.Err()
}

// MapRowsToSliceOfMaps maps db rows to maps
// the map keys are the column names (or the aliases if defined in the query)
func MapRowsToSliceOfMaps(scanner RowsScanner, Map *[]map[string]interface{}) error {
	defer scanner.Close()
	for scanner.Next() {
		columns, err := scanner.Columns()
		if err != nil {
			return err
		}
		sliceOfResults := make([]interface{}, len(columns))
		for i := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// create a new interface{} than is not nil
			sliceOfResults[i] = new(interface{})
		}
		err = scanner.Scan(sliceOfResults...)
		if err != nil {
			return err
		}
		row := map[string]interface{}{}
		for index, column := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// convert the sliceOfResults[index] value back to interface{}
			var v interface{} = *(sliceOfResults[index].(*interface{}))
			if u, ok := v.([]uint8); ok {
				// v is likely a string, so convert to string
				row[column] = (interface{}(string(u)))
			} else {
				row[column] = v
			}
		}
		*Map = append(*Map, row)
	}
	return scanner.Err()
}

// MapRowsToSliceOfStruct  maps db rows to structs
func MapRowsToSliceOfStruct(scanner RowsScanner, pointerToASliceOfStructs interface{}, ignoreMissingField bool, transforms ...func(string) string) error {
	defer scanner.Close()
	recordsPointerValue := reflect.ValueOf(pointerToASliceOfStructs)
	if recordsPointerValue.Kind() != reflect.Ptr {
		return fmt.Errorf("Expect pointer, got %#v", pointerToASliceOfStructs)
	}
	recordsValue := recordsPointerValue.Elem()
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("The underlying type is not a slice,pointer to slice expected for %#v ", recordsValue)
	}

	columns, err := scanner.Columns()
	if err != nil {
		return err
	}

	// get the underlying type of a slice
	// @see http://stackoverflow.com/questions/24366895/golang-reflect-slice-underlying-type
	for scanner.Next() {
		//
		var t reflect.Type
		if recordsValue.Type().Elem().Kind() == reflect.Ptr {
			// the sliceOfStructs type is like []*T
			t = recordsValue.Type().Elem().Elem()
		} else {
			// the sliceOfStructs type is like []T
			t = recordsValue.Type().Elem()
		}
		pointerOfElement := reflect.New(t)

		err = MapRowToStruct(columns, scanner, pointerOfElement.Interface(), ignoreMissingField, transforms...)
		if err != nil {
			return err
		}
		recordsValue = reflect.Append(recordsValue, pointerOfElement)
	}
	recordsPointerValue.Elem().Set(recordsValue)
	return scanner.Err()
}

// MapRowToStruct  automatically maps a db row to a struct.
//
// columns are the names of the columns in the row, they should match the fieldnames of Struct unless an optional transform function
// is passed.
//
// scanner is the Scanner (a sql.Row type for instance).
//
// Struct is a pointer to the struct that needs to be populated by the row data.
//
// ignoreMissingFields will ignore missing fields if the number of columns in the row doesn't match the number of fields in the struct.
//
// transforms is an optinal function that changes the name of the columns to match the name of the fields.
func MapRowToStruct(columns []string, scanner Scanner, Struct interface{}, ignoreMissingFields bool, transforms ...func(string) string) error {
	if len(transforms) == 0 {
		transforms = []func(string) string{noop}
	}
	structPointer := reflect.ValueOf(Struct)
	if structPointer.Kind() != reflect.Ptr {
		return fmt.Errorf("Pointer expected, got %#v", Struct)
	}
	structValue := reflect.Indirect(structPointer)
	zeroValue := reflect.Value{}
	arrayOfResults := []interface{}{}
	for _, column := range columns {
		column = transforms[0](column)
		field := structValue.FieldByName(column)
		if field == zeroValue {
			if ignoreMissingFields {
				pointer := reflect.New(reflect.TypeOf([]byte{}))
				pointer.Elem().Set(reflect.ValueOf([]byte{}))
				arrayOfResults = append(arrayOfResults, pointer.Interface())

			} else {
				return fmt.Errorf("No field found for column %s in struct %#v", column, Struct)

			}
		} else {
			if !field.CanSet() {
				return fmt.Errorf("Unexported field %s cannot be set in struct %#v", column, Struct)
			}
			arrayOfResults = append(arrayOfResults, field.Addr().Interface())
		}
	}
	err := scanner.Scan(arrayOfResults...)
	if err != nil {
		return err
	}
	return nil
}

func noop(s string) string { return s }

// CreateTagMapperFunc creates a function that
// can be used to map db fields to struct fields
// through the use of struct tags.
//
// For instance :
//
//      type Foo struct{
//         Bar `sql:"bar"`
//      }
//
//      foo := new(Foo)
//      tagMapper := CreateTagMapperFunc(Foo{})
//      err := MapRowToStruct([]string{"bar"},someRow,foo,true,tagMapper)
//
// Will map Bar field in struct to bar DB field in the row
func CreateTagMapperFunc(Struct interface{}, tagname ...string) (func(string) string, error) {
	structValue := reflect.Indirect(reflect.ValueOf(Struct))
	if structValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Struct expected, got %#v", Struct)
	}
	if len(tagname) == 0 {
		tagname = []string{"sql"}
	}
	m := map[string]string{}
	for i := 0; i < structValue.NumField(); i++ {
		name := structValue.Type().Field(i).Name
		tag := structValue.Type().Field(i).Tag.Get(tagname[0])
		if tag == "" {
			tag = name
		}
		m[tag] = name
	}
	return func(s string) string {
		if r, ok := m[s]; ok {
			return r
		}
		return s
	}, nil
}
