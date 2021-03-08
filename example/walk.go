package main

import (
	"log"

	"github.com/auxten/postgresql-parser/pkg/walk"
)

func main() {
	sql := `select marr
			from (select marr_stat_cd AS marr, label AS l
				  from root_loan_mock_v4
				  order by root_loan_mock_v4.age desc, l desc
				  limit 5) as v4
			LIMIT 1;`
	w := &walk.AstWalker{
		Fn: func(ctx interface{}, node interface{}) (stop bool) {
			log.Printf("node type %T", node)
			return false
		},
	}
	_, _ = w.Walk(sql, nil)
	return
}
