package querystring

import (
	"fmt"
	"reflect"
	"strings"
)

type selectString struct {
	wherePtr   *whereString
	fieldArray []string
}

func Select(fields string, from string) *selectString {
	return &selectString{
		wherePtr: newWhereString("SELECT", from, fields),
	}
}

func SelectArray(fieldArray []string, from string) *selectString {
	return &selectString{
		wherePtr: newWhereString("SELECT", from, strings.Join(fieldArray, ",")),
	}
}

func SelectObject(object string, from string) *selectString {
	fieldArray := make([]string, 0)
	valueof := reflect.ValueOf(object)
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
	return &selectString{
		wherePtr: newWhereString("SELECT", from, strings.Join(fieldArray, ",")),
	}
}

func (this *selectString) Where(where string) *selectString {
	this.wherePtr.Where(where)
	return this
}

func (this *selectString) EQ(field string, value interface{}) *selectString {
	this.wherePtr.EQ(field, value)
	return this
}

func (this *selectString) GT(field string, value interface{}) *selectString {
	this.wherePtr.GT(field, value)
	return this
}

func (this *selectString) GE(field string, value interface{}) *selectString {
	this.wherePtr.GE(field, value)
	return this
}

func (this *selectString) LT(field string, value interface{}) *selectString {
	this.wherePtr.LT(field, value)
	return this
}

func (this *selectString) LE(field string, value interface{}) *selectString {
	this.wherePtr.LE(field, value)
	return this
}

func (this *selectString) IN(field string, intArray []int64) *selectString {
	this.wherePtr.IN(field, intArray)
	return this
}

func (this *selectString) Between(field string, min int, max int) *selectString {
	this.wherePtr.GE(field, min)
	this.wherePtr.LT(field, max)
	return this
}

func (this *selectString) LeftJoin(table string, on string) *selectString {
	this.wherePtr.LeftJoin(table, on)
	return this
}

func (this *selectString) RightJoin(table string, on string) *selectString {
	this.wherePtr.RightJoin(table, on)
	return this
}

func (this *selectString) GroupBy(group string) *selectString {
	this.wherePtr.GroupBy(group)
	return this
}

func (this *selectString) OrderBy(order string) *selectString {
	this.wherePtr.OrderBy(order)
	return this
}

func (this *selectString) Offset(pos int64) *selectString {
	this.wherePtr.Offset(pos)
	return this
}

func (this *selectString) Limit(limit int64) *selectString {
	this.wherePtr.Limit(limit)
	return this
}

func (this *selectString) GetSQL() string {
	return this.wherePtr.ToString()
}
