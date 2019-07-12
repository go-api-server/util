package main

import (
	"fmt"

	"github.com/go-api-server/util/mysql/querystring"
)

func main() {
	{
		sql := querystring.Select("a,b,c,d,e", "tbl").EQ("tblA.id", 1)
		sql.LeftJoin("tblB", "tblA.id = tblB.aid")
		sql.RightJoin("tblC", "tblC.id = tblB.cid")
		sql.OrderBy("tblA.id DESC")
		fmt.Println("select: ", sql.GetSQL())
	}

	{
		data := make(map[string]interface{})
		data["a"] = 1
		data["b"] = "z"
		data["c"] = true
		sql := querystring.InsertInto("A").Insert(data)
		fmt.Println("insert: ", sql.GetSQL())
	}

	{
		data := make(map[string]interface{})
		data["a"] = 1
		data["b"] = "z"
		data["c"] = true
		sql := querystring.Update("A").SetMapping(data).EQ("id", 10)
		fmt.Println("update: ", sql.GetSQL())
	}

	{
		sql := querystring.Delete("A").EQ("ID", 1)
		fmt.Println("update: ", sql.GetSQL())
	}
}
