package querystring

import (
	"testing"
)

func TestSelect(t *testing.T) {
	sql := Select("a,b,c,d,e", "tbl").EQ("tblA.id", 1).LeftJoin("tblB", "tblA.id = tblB.aid").OrderBy("tblA.id DESC").GetSQL()
	t.Logf("sql: %s", sql)
}
