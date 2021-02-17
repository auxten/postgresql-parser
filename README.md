# What's this
PostgreSQL style Parser splitted from [CockroachDB](github.com/cockroachdb/cockroach)

See: [Complex SQL format example](example/format.go)

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

# Features
- Pure golang implementation
- *Almost* full support of PostgreSQL (`cockroachdb` style PostgreSQL)

### 🚧🚧🚧 still under construction 🚧🚧🚧

# Progress
- 2020-02-16 `github.com/auxten/postgresql-parser/pkg/sql/parser` Unit tests works now!

# Todo
- Fix more unit tests
- Make built-in function parsing work
