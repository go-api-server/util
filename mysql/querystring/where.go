package querystring

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type whereMaker struct {
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

func newWhereMaker(command string, table string, fields string) *whereMaker {
	return &whereMaker{
		command:    command,
		table:      table,
		fields:     fields,
		whereArray: make([]string, 0),
		joinArray:  make([]string, 0),
	}
}

func (this *whereMaker) ToString() string {
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

func (this *whereMaker) Where(where string) *whereMaker {
	this.whereArray = append(this.whereArray, where)
	return this
}

func (this *whereMaker) appendWhereString(field string, cmp string, value interface{}) {
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

func (this *whereMaker) EQ(field string, value interface{}) *whereMaker {
	this.appendWhereString(field, "=", value)
	return this
}

func (this *whereMaker) GT(field string, value interface{}) *whereMaker {
	this.appendWhereString(field, ">", value)
	return this
}

func (this *whereMaker) GE(field string, value interface{}) *whereMaker {
	this.appendWhereString(field, ">=", value)
	return this
}

func (this *whereMaker) LT(field string, value interface{}) *whereMaker {
	this.appendWhereString(field, "<", value)
	return this
}

func (this *whereMaker) LE(field string, value interface{}) *whereMaker {
	this.appendWhereString(field, "<=", value)
	return this
}

func (this *whereMaker) IN(field string, intArray []int64) *whereMaker {
	stringArray := make([]string, len(intArray))
	for i, v := range intArray {
		stringArray[i] = strconv.FormatInt(v, 64)
	}
	this.appendWhereString(field, "IN", fmt.Sprintf("(%s)", strings.Join(stringArray, ",")))
	return this
}

func (this *whereMaker) Between(field string, min int, max int) *whereMaker {
	this.appendWhereString(field, ">=", min)
	this.appendWhereString(field, "<", max)
	return this
}

func (this *whereMaker) LeftJoin(table string, on string) *whereMaker {
	this.joinArray = append(this.joinArray, fmt.Sprintf("LEFT JOIN %s ON %s", table, on))
	return this
}

func (this *whereMaker) RightJoin(table string, on string) *whereMaker {
	this.joinArray = append(this.joinArray, fmt.Sprintf("RIGHT JOIN %s ON %s", table, on))
	return this
}

func (this *whereMaker) GroupBy(group string) *whereMaker {
	this.group = group
	return this
}

func (this *whereMaker) OrderBy(order string) *whereMaker {
	this.order = order
	return this
}

func (this *whereMaker) Offset(pos int64) *whereMaker {
	this.start = pos
	return this
}

func (this *whereMaker) Limit(limit int64) *whereMaker {
	this.limit = limit
	return this
}
