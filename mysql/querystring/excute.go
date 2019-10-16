package querystring

import (
	"fmt"
	"reflect"
	"strings"
)

type excuteSQL struct {
	wherePtr    *whereMaker
	updateArray []string
}

func Delete(table string) *excuteSQL {
	return &excuteSQL{
		wherePtr:    newWhereMaker("DELETE", table, ""),
		updateArray: make([]string, 0),
	}
}

func Update(table string) *excuteSQL {
	return &excuteSQL{
		wherePtr:    newWhereMaker("UPDATE", table, ""),
		updateArray: make([]string, 0),
	}
}

func (this *excuteSQL) Set(field string, value interface{}) *excuteSQL {
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

func (this *excuteSQL) SetFieldAndValue(fieldAndValue map[string]interface{}) *excuteSQL {
	for k, v := range fieldAndValue {
		this.Set(k, v)
	}
	return this
}

func (this *excuteSQL) Where(where string) *excuteSQL {
	this.wherePtr.Where(where)
	return this
}

func (this *excuteSQL) EQ(field string, value interface{}) *excuteSQL {
	this.wherePtr.EQ(field, value)
	return this
}

func (this *excuteSQL) GT(field string, value interface{}) *excuteSQL {
	this.wherePtr.GT(field, value)
	return this
}

func (this *excuteSQL) GE(field string, value interface{}) *excuteSQL {
	this.wherePtr.GE(field, value)
	return this
}

func (this *excuteSQL) LT(field string, value interface{}) *excuteSQL {
	this.wherePtr.LT(field, value)
	return this
}

func (this *excuteSQL) LE(field string, value interface{}) *excuteSQL {
	this.wherePtr.LE(field, value)
	return this
}

func (this *excuteSQL) IN(field string, intArray []int64) *excuteSQL {
	this.wherePtr.IN(field, intArray)
	return this
}

func (this *excuteSQL) Between(field string, min int, max int) *excuteSQL {
	this.wherePtr.GE(field, min)
	this.wherePtr.LT(field, max)
	return this
}

func (this *excuteSQL) GetSQL() string {
	this.wherePtr.fields = strings.Join(this.updateArray, ",")
	return this.wherePtr.ToString()
}
