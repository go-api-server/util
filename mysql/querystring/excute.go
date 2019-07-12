package querystring

import (
	"fmt"
	"reflect"
	"strings"
)

type excuteString struct {
	wherePtr    *whereString
	updateArray []string
}

func Delete(table string) *excuteString {
	return &excuteString{
		wherePtr:    newWhereString("DELETE", table, ""),
		updateArray: make([]string, 0),
	}
}

func Update(table string) *excuteString {
	return &excuteString{
		wherePtr:    newWhereString("UPDATE", table, ""),
		updateArray: make([]string, 0),
	}
}

func (this *excuteString) Set(field string, value interface{}) *excuteString {
	valueof := reflect.ValueOf(value)
	switch valueof.Type().Kind() {
	case reflect.String:
		this.updateArray = append(this.updateArray, fmt.Sprintf("%s = '%s'", field, Escape(valueof.String())))
	case reflect.Bool:
		this.updateArray = append(this.updateArray, fmt.Sprintf("%s = %t", field, valueof.Bool()))
	case reflect.Int, reflect.Int32, reflect.Int64:
		this.updateArray = append(this.updateArray, fmt.Sprintf("%s = %d", field, valueof.Int()))
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		this.updateArray = append(this.updateArray, fmt.Sprintf("%s = %d", field, valueof.Uint()))
	case reflect.Float32, reflect.Float64:
		this.updateArray = append(this.updateArray, fmt.Sprintf("%s = %f", field, valueof.Float()))
	}
	return this
}

func (this *excuteString) SetMapping(fieldAndValue map[string]interface{}) *excuteString {
	for k, v := range fieldAndValue {
		this.Set(k, v)
	}
	return this
}

func (this *excuteString) Where(where string) *excuteString {
	this.wherePtr.Where(where)
	return this
}

func (this *excuteString) EQ(field string, value interface{}) *excuteString {
	this.wherePtr.EQ(field, value)
	return this
}

func (this *excuteString) GT(field string, value interface{}) *excuteString {
	this.wherePtr.GT(field, value)
	return this
}

func (this *excuteString) GE(field string, value interface{}) *excuteString {
	this.wherePtr.GE(field, value)
	return this
}

func (this *excuteString) LT(field string, value interface{}) *excuteString {
	this.wherePtr.LT(field, value)
	return this
}

func (this *excuteString) LE(field string, value interface{}) *excuteString {
	this.wherePtr.LE(field, value)
	return this
}

func (this *excuteString) IN(field string, intArray []int64) *excuteString {
	this.wherePtr.IN(field, intArray)
	return this
}

func (this *excuteString) Between(field string, min int, max int) *excuteString {
	this.wherePtr.GE(field, min)
	this.wherePtr.LT(field, max)
	return this
}

func (this *excuteString) GetSQL() string {
	this.wherePtr.fields = strings.Join(this.updateArray, ",")
	return this.wherePtr.ToString()
}
