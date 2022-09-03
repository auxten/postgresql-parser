# What's this
PostgreSQL style Parser splitted from [CockroachDB](https://github.com/cockroachdb/cockroach)

See: [Complex SQL format example](example/format/format.go)

I tried to import `github.com/cockroachdb/cockroach/pkg/sql/parser`, but the dependencies is too complex to make it work. 

To make things easy, I did these things:

1. Copy all the `pkg/sql/parser`, `pkg/sql/lex` and simplify the dependencies
2. Simplify the Makefile to just generate the goyacc stuff
3. Add the goyacc generated files in parser and lex to make `go get` work easily, see the `.gitignore` files
4. Trim the `etcd` dependency, see the `go.mod`
5. Rename all test file except some `pkg/sql/parser` tests
6. Add all necessary imports to vendor
7. Remove the `panic` of meeting unregistried functions, see the [WrapFunction](pkg/sql/sem/tree/function_name.go#L67)
8. Other nasty things make the parser just work that I forgot :p

# Who is using this

- [Atlas](https://github.com/ariga/atlas) 1.8k stars
- [ByteBase](https://github.com/bytebase/bytebase) 3.6k stars
- [More](https://github.com/auxten/postgresql-parser/network/dependents)

# Features
- Pure golang implementation
- *Almost* full support of PostgreSQL (`cockroachdb` style PostgreSQL)

## SQL Standard Compliance

The code is derived from CockroachDB v20.1.11 which supports most of the major features of SQL:2011. See:

- https://www.cockroachlabs.com/docs/v20.1/postgresql-compatibility

- https://www.postgresql.org/docs/9.5/features.html

# How to use

```go
package main

import (
	"log"
	
	"github.com/auxten/postgresql-parser/pkg/sql/parser"
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
	
	stmts, err := parser.Parse(sql)
	if err != nil {
		return
	}
	
	_, _ = w.Walk(stmts, nil)
	return
}

```

# SQL parser

This project contains code that is automatically generated using `goyacc`.
`goyacc` reads the SQL expressions ([`sql.y`](https://github.com/auxten/postgresql-parser/blob/main/pkg/sql/parser/sql.y)) and generates a parser which could be used to tokenize a given input.
You could update the generated code using the `generate` target inside the project's Makefile.

```bash
$ make generate
```

# Progress
- 2021-02-16 `github.com/auxten/postgresql-parser/pkg/sql/parser` Unit tests works now!
- 2021-03-08 Add walker package.
- 2022-08-03 Remove vendored dependencies by @mostafa in https://github.com/auxten/postgresql-parser/pull/19

# Todo
- Fix more unit tests
- Make built-in function parsing work
