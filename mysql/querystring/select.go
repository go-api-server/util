package querystring

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type selectSQL struct {
	wherePtr *whereMaker
}

func Count(table string) *selectSQL {
	return &selectSQL{
		wherePtr: newWhereMaker("SELECT", table, "COUNT(*)"),
	}
}

func Select(fields string, from string) *selectSQL {
	return &selectSQL{
		wherePtr: newWhereMaker("SELECT", from, fields),
	}
}

func SelectArray(fieldArray []string, from string) *selectSQL {
	return &selectSQL{
		wherePtr: newWhereMaker("SELECT", from, strings.Join(fieldArray, ",")),
	}
}

func SelectObject(object interface{}, from string) *selectSQL {
	fieldArray := make([]string, 0)
	valueof := reflect.ValueOf(object)
	if valueof.Type().Kind() == reflect.Ptr {
		valueof = valueof.Elem()
	}
	if valueof.Type().Kind() != reflect.Struct {
		panic(fmt.Sprintf("querystring[SelectObject] %s is not a struct", valueof.Type().Name()))
	}
	for i := 0; i < valueof.Type().NumField(); i++ {
		tags := valueof.Type().Field(i).Tag.Get("db")
		if len(tags) == 0 {
			continue
		}
		tagArray := strings.Split(tags, ",")
		fieldArray = append(fieldArray, tagArray[0])
	}
	return &selectSQL{
		wherePtr: newWhereMaker("SELECT", from, strings.Join(fieldArray, ",")),
	}
}

func (this *selectSQL) Where(where string) *selectSQL {
	this.wherePtr.Where(where)
	return this
}

func (this *selectSQL) EQ(field string, value interface{}) *selectSQL {
	this.wherePtr.EQ(field, value)
	return this
}

func (this *selectSQL) GT(field string, value interface{}) *selectSQL {
	this.wherePtr.GT(field, value)
	return this
}

func (this *selectSQL) GE(field string, value interface{}) *selectSQL {
	this.wherePtr.GE(field, value)
	return this
}

func (this *selectSQL) LT(field string, value interface{}) *selectSQL {
	this.wherePtr.LT(field, value)
	return this
}

func (this *selectSQL) LE(field string, value interface{}) *selectSQL {
	this.wherePtr.LE(field, value)
	return this
}

func (this *selectSQL) IN(field string, intArray []int64) *selectSQL {
	this.wherePtr.IN(field, intArray)
	return this
}

func (this *selectSQL) Between(field string, min int, max int) *selectSQL {
	this.wherePtr.GE(field, min)
	this.wherePtr.LT(field, max)
	return this
}

func (this *selectSQL) LeftJoin(table string, on string) *selectSQL {
	this.wherePtr.LeftJoin(table, on)
	return this
}

func (this *selectSQL) RightJoin(table string, on string) *selectSQL {
	this.wherePtr.RightJoin(table, on)
	return this
}

func (this *selectSQL) GroupBy(group string) *selectSQL {
	this.wherePtr.GroupBy(group)
	return this
}

func (this *selectSQL) OrderBy(order string) *selectSQL {
	this.wherePtr.OrderBy(order)
	return this
}

func (this *selectSQL) Offset(pos int64) *selectSQL {
	this.wherePtr.Offset(pos)
	return this
}

func (this *selectSQL) Limit(limit int64) *selectSQL {
	this.wherePtr.Limit(limit)
	return this
}

func (this *selectSQL) GetSQL() string {
	return this.wherePtr.ToString()
}

func (this *selectSQL) GetObject(out interface{}, db *sql.DB) (bool, error) {
	dest := reflect.ValueOf(out)

	if dest.Type().Kind() == reflect.Ptr {
		dest = dest.Elem()
	}

	vtype := dest.Type()
	if vtype.Kind() != reflect.Struct {
		return false, errors.New(fmt.Sprintf("dest: %s is not a struct", vtype.Name()))
	}

	var fieldArray []string
	if this.wherePtr.fields != "*" {
		fieldArray = strings.Split(this.wherePtr.fields, ",")
	} else {
		for i := 0; i < vtype.NumField(); i++ {
			tag := strings.Trim(vtype.Field(i).Tag.Get("db"), " ")
			if len(tag) == 0 || tag == "-" {
				continue
			}
			arr := strings.Split(tag, " ")
			fieldArray = append(fieldArray, arr[0])
		}
	}
	fieldMapper := make(map[string]int)
	for i, v := range fieldArray {
		fieldMapper[v] = i
	}

	fieldCount := len(fieldArray)
	valueArray := make([][]byte, fieldCount)
	valueAddrArray := make([]interface{}, fieldCount)
	for i := 0; i < fieldCount; i++ {
		valueAddrArray[i] = &valueArray[i]
	}

	row := db.QueryRow(this.GetSQL())
	err := row.Scan(valueAddrArray...)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	for i := 0; i < vtype.NumField(); i++ {
		tag := strings.Trim(vtype.Field(i).Tag.Get("db"), " ")
		if len(tag) == 0 || tag == "-" {
			continue
		}
		arr := strings.Split(tag, " ")
		idx, ok := fieldMapper[arr[0]]
		if !ok {
			continue
		}
		str := string(valueArray[idx])
		switch vtype.Field(i).Type.Kind() {
		case reflect.String:
			dest.Field(i).SetString(str)
		case reflect.Bool:
			val, err := strconv.ParseBool(str)
			if err == nil {
				dest.Field(i).SetBool(val)
			}
		case reflect.Float32, reflect.Float64:
			val, err := strconv.ParseFloat(str, 64)
			if err == nil {
				dest.Field(i).SetFloat(val)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := strconv.ParseInt(str, 0, 64)
			if err == nil {
				dest.Field(i).SetInt(val)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, err := strconv.ParseUint(str, 0, 64)
			if err == nil {
				dest.Field(i).SetUint(val)
			}
		}
	}

	return true, nil
}

func (this *selectSQL) GetObjectArray(out interface{}, db *sql.DB) (int, error) {
	sliceType := reflect.TypeOf(out)
	if sliceType.Kind() == reflect.Ptr {
		sliceType = sliceType.Elem()
	}

	if sliceType.Kind() != reflect.Slice {
		return 0, errors.New(fmt.Sprintf("dest: %s is not a slice", sliceType.Name()))
	}

	memberType := sliceType.Elem()
	memberIsPtr := false
	if memberType.Kind() == reflect.Ptr {
		memberType = memberType.Elem()
		memberIsPtr = true
	}

	var fieldArray []string
	if this.wherePtr.fields != "*" {
		fieldArray = strings.Split(this.wherePtr.fields, ",")
	} else {
		for i := 0; i < sliceType.NumField(); i++ {
			tag := strings.Trim(sliceType.Field(i).Tag.Get("db"), " ")
			if len(tag) == 0 || tag == "-" {
				continue
			}
			arr := strings.Split(tag, " ")
			fieldArray = append(fieldArray, arr[0])
		}
	}
	fieldMapper := make(map[string]int)
	for i, v := range fieldArray {
		fieldMapper[v] = i
	}
	fieldCount := len(fieldArray)

	rows, err := db.Query(this.GetSQL())
	if err != nil {
		return 0, err
	}

	count := 0
	array := reflect.MakeSlice(sliceType, 0, 0)

	for rows.Next() {
		valueArray := make([][]byte, fieldCount)
		valueAddrArray := make([]interface{}, fieldCount)
		for i := 0; i < fieldCount; i++ {
			valueAddrArray[i] = &valueArray[i]
		}
		err := rows.Scan(valueAddrArray...)
		if err != nil {
			return 0, err
		}
		count++

		obj := reflect.New(memberType)
		dest := obj.Elem()
		for i := 0; i < memberType.NumField(); i++ {
			tag := strings.Trim(memberType.Field(i).Tag.Get("db"), " ")
			if len(tag) == 0 || tag == "-" {
				continue
			}
			arr := strings.Split(tag, " ")
			idx, ok := fieldMapper[arr[0]]
			if !ok {
				continue
			}
			str := string(valueArray[idx])
			switch memberType.Field(i).Type.Kind() {
			case reflect.String:
				dest.Field(i).SetString(str)
			case reflect.Bool:
				val, err := strconv.ParseBool(str)
				if err == nil {
					dest.Field(i).SetBool(val)
				}
			case reflect.Float32, reflect.Float64:
				val, err := strconv.ParseFloat(str, 64)
				if err == nil {
					dest.Field(i).SetFloat(val)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val, err := strconv.ParseInt(str, 0, 64)
				if err == nil {
					dest.Field(i).SetInt(val)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				val, err := strconv.ParseUint(str, 0, 64)
				if err == nil {
					dest.Field(i).SetUint(val)
				}
			}
		}
		if memberIsPtr {
			array = reflect.Append(array, obj)
		} else {
			array = reflect.Append(array, dest)
		}
	}

	dest := reflect.ValueOf(out)
	dest.Elem().Set(array)

	return count, nil
}
