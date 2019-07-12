package querystring

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type whereString struct {
	command    string
	table      string
	fields     string
	whereArray []string
	joinArray  []string
	group      string
	order      string
	start      int64
	limit      int64
}

func newWhereString(command string, table string, fields string) *whereString {
	return &whereString{
		command:    command,
		table:      table,
		fields:     fields,
		whereArray: make([]string, 0),
		joinArray:  make([]string, 0),
	}
}

func (this *whereString) ToString() string {
	sqlString := ""

	if this.command == "SELECT" {
		sqlString = fmt.Sprintf("%s %s FROM %s", this.command, this.fields, this.table)
	} else if this.command == "DELETE" {
		sqlString = fmt.Sprintf("%s FROM %s", this.command, this.table)
	} else if this.command == "UPDATE" {
		sqlString = fmt.Sprintf("%s %s SET %s", this.command, this.table, this.fields)
	}

	if len(this.whereArray) > 0 {
		sqlString += fmt.Sprintf(" WHERE %s", strings.Join(this.whereArray, " AND "))
	}

	if len(this.joinArray) > 0 {
		sqlString += " "
		sqlString += strings.Join(this.joinArray, " ")
	}

	if len(this.group) > 0 {
		sqlString += fmt.Sprintf(" GROUP BY %s", this.group)
	}

	if len(this.order) > 0 {
		sqlString += fmt.Sprintf(" ORDER BY %s", this.order)
	}

	if this.limit > 0 {
		if this.start > 0 {
			sqlString += fmt.Sprintf(" LIMIT %d, %d", this.start, this.limit)
		} else {
			sqlString += fmt.Sprintf(" LIMIT %d", this.limit)
		}
	}

	return sqlString
}

func (this *whereString) Where(where string) *whereString {
	this.whereArray = append(this.whereArray, where)
	return this
}

func (this *whereString) appendWhereString(field string, cmp string, value interface{}) {
	valueof := reflect.ValueOf(value)
	switch valueof.Type().Kind() {
	case reflect.String:
		this.whereArray = append(this.whereArray, fmt.Sprintf("%s = '%s'", field, valueof.String()))
	case reflect.Bool:
		this.whereArray = append(this.whereArray, fmt.Sprintf("%s = %t", field, valueof.Bool()))
	case reflect.Int, reflect.Int32, reflect.Int64:
		this.whereArray = append(this.whereArray, fmt.Sprintf("%s = %d", field, valueof.Int()))
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		this.whereArray = append(this.whereArray, fmt.Sprintf("%s = %d", field, valueof.Uint()))
	case reflect.Float32, reflect.Float64:
		this.whereArray = append(this.whereArray, fmt.Sprintf("%s = %f", field, valueof.Float()))
	}
}

func (this *whereString) EQ(field string, value interface{}) *whereString {
	this.appendWhereString(field, "=", value)
	return this
}

func (this *whereString) GT(field string, value interface{}) *whereString {
	this.appendWhereString(field, ">", value)
	return this
}

func (this *whereString) GE(field string, value interface{}) *whereString {
	this.appendWhereString(field, ">=", value)
	return this
}

func (this *whereString) LT(field string, value interface{}) *whereString {
	this.appendWhereString(field, "<", value)
	return this
}

func (this *whereString) LE(field string, value interface{}) *whereString {
	this.appendWhereString(field, "<=", value)
	return this
}

func (this *whereString) IN(field string, intArray []int64) *whereString {
	stringArray := make([]string, len(intArray))
	for i, v := range intArray {
		stringArray[i] = strconv.FormatInt(v, 64)
	}
	this.appendWhereString(field, "IN", fmt.Sprintf("(%s)", strings.Join(stringArray, ",")))
	return this
}

func (this *whereString) Between(field string, min int, max int) *whereString {
	this.appendWhereString(field, ">=", min)
	this.appendWhereString(field, "<", max)
	return this
}

func (this *whereString) LeftJoin(table string, on string) *whereString {
	this.joinArray = append(this.joinArray, fmt.Sprintf("LEFT JOIN %s ON %s", table, on))
	return this
}

func (this *whereString) RightJoin(table string, on string) *whereString {
	this.joinArray = append(this.joinArray, fmt.Sprintf("RIGHT JOIN %s ON %s", table, on))
	return this
}

func (this *whereString) GroupBy(group string) *whereString {
	this.group = group
	return this
}

func (this *whereString) OrderBy(order string) *whereString {
	this.order = order
	return this
}

func (this *whereString) Offset(pos int64) *whereString {
	this.start = pos
	return this
}

func (this *whereString) Limit(limit int64) *whereString {
	this.limit = limit
	return this
}
