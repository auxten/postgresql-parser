package walk

import (
	"fmt"
	"log"
	"strings"

	"github.com/auxten/postgresql-parser/pkg/sql/parser"
	"github.com/auxten/postgresql-parser/pkg/sql/sem/tree"
	"github.com/auxten/postgresql-parser/pkg/util/set"
)

type AstWalker struct {
	UnknownNodes []interface{}
	Fn           func(ctx interface{}, node interface{}) (stop bool)
}
type ReferredCols map[string]int

func (rc ReferredCols) ToList() []string {
	cols := make([]string, len(rc))
	i := 0
	for k := range rc {
		cols[i] = k
		i++
	}
	return set.SortDeDup(cols)
}

func (w *AstWalker) Walk(stmts parser.Statements, ctx interface{}) (ok bool, err error) {

	w.UnknownNodes = make([]interface{}, 0)
	asts := make([]tree.NodeFormatter, len(stmts))
	for si, stmt := range stmts {
		asts[si] = stmt.AST
	}

	// nodeCount is incremented on each visited node per statement. It is
	// currently used to determine if walk is at the top-level statement
	// or not.
	var walk func(...interface{})
	walk = func(nodes ...interface{}) {
		for _, node := range nodes {
			if w.Fn != nil {
				if w.Fn(ctx, node) {
					break
				}
			}

			if node == nil {
				continue
			}
			if _, ok := node.(tree.Datum); ok {
				continue
			}

			switch node := node.(type) {
			case *tree.AliasedTableExpr:
				walk(node.Expr)
			case *tree.AndExpr:
				walk(node.Left, node.Right)
			case *tree.AnnotateTypeExpr:
				walk(node.Expr)
			case *tree.Array:
				walk(node.Exprs)
			case tree.AsOfClause:
				walk(node.Expr)
			case *tree.BinaryExpr:
				walk(node.Left, node.Right)
			case *tree.CaseExpr:
				walk(node.Expr, node.Else)
				for _, when := range node.Whens {
					walk(when.Cond, when.Val)
				}
			case *tree.RangeCond:
				walk(node.Left, node.From, node.To)
			case *tree.CastExpr:
				walk(node.Expr)
			case *tree.CoalesceExpr:
				for _, expr := range node.Exprs {
					walk(expr)
				}
			case *tree.ColumnTableDef:
			case *tree.ComparisonExpr:
				walk(node.Left, node.Right)
			case *tree.CreateTable:
				for _, def := range node.Defs {
					walk(def)
				}
				if node.AsSource != nil {
					walk(node.AsSource)
				}
			case *tree.CTE:
				walk(node.Stmt)
			case *tree.DBool:
			case tree.Exprs:
				for _, expr := range node {
					walk(expr)
				}
			case *tree.FamilyTableDef:
			case *tree.From:
				walk(node.AsOf)
				for _, table := range node.Tables {
					walk(table)
				}
			case *tree.FuncExpr:
				if node.WindowDef != nil {
					walk(node.WindowDef)
				}
				walk(node.Exprs, node.Filter)
			case *tree.IndexTableDef:
			case *tree.JoinTableExpr:
				walk(node.Left, node.Right, node.Cond)
			case *tree.NotExpr:
				walk(node.Expr)
			case *tree.NumVal:
			case *tree.OnJoinCond:
				walk(node.Expr)
			case *tree.Order:
				walk(node.Expr, node.Table)
			case tree.OrderBy:
				for _, order := range node {
					walk(order)
				}
			case *tree.OrExpr:
				walk(node.Left, node.Right)
			case *tree.ParenExpr:
				walk(node.Expr)
			case *tree.ParenSelect:
				walk(node.Select)
			case *tree.RowsFromExpr:
				for _, expr := range node.Items {
					walk(expr)
				}
			case *tree.Select:
				if node.With != nil {
					walk(node.With)
				}
				if node.OrderBy != nil {
					walk(node.OrderBy)
				}
				if node.Limit != nil {
					walk(node.Limit)
				}
				walk(node.Select)
			case *tree.Limit:
				walk(node.Count)
			case *tree.SelectClause:
				walk(node.Exprs)
				if node.Where != nil {
					walk(node.Where)
				}
				if node.Having != nil {
					walk(node.Having)
				}
				if node.DistinctOn != nil {
					for _, distinct := range node.DistinctOn {
						walk(distinct)
					}
				}
				if node.GroupBy != nil {
					for _, group := range node.GroupBy {
						walk(group)
					}
				}
				walk(&node.From)
			case tree.SelectExpr:
				walk(node.Expr)
			case tree.SelectExprs:
				for _, expr := range node {
					walk(expr)
				}
			case *tree.SetVar:
				for _, expr := range node.Values {
					walk(expr)
				}
			case *tree.StrVal:
			case *tree.Subquery:
				walk(node.Select)
			case tree.TableExprs:
				for _, expr := range node {
					walk(expr)
				}
			case *tree.TableName, tree.TableName:
			case *tree.Tuple:
				for _, expr := range node.Exprs {
					walk(expr)
				}
			case *tree.UnaryExpr:
				walk(node.Expr)
			case *tree.UniqueConstraintTableDef:
			case *tree.UnionClause:
				walk(node.Left, node.Right)
			case tree.UnqualifiedStar:
			case *tree.UnresolvedName:
			case *tree.ValuesClause:
				for _, row := range node.Rows {
					walk(row)
				}
			case *tree.Where:
				walk(node.Expr)
			case tree.Window:
				for _, windowDef := range node {
					walk(windowDef)
				}
			case *tree.WindowDef:
				walk(node.Partitions)
				if node.Frame != nil {
					walk(node.Frame)
				}
			case *tree.WindowFrame:
				if node.Bounds.StartBound != nil {
					walk(node.Bounds.StartBound)
				}
				if node.Bounds.EndBound != nil {
					walk(node.Bounds.EndBound)
				}
			case *tree.WindowFrameBound:
				walk(node.OffsetExpr)
			case *tree.With:
				for _, expr := range node.CTEList {
					walk(expr)
				}
			default:
				if w.UnknownNodes != nil {
					w.UnknownNodes = append(w.UnknownNodes, node)
				}
			}
		}
	}

	for _, ast := range asts {
		walk(ast)
	}

	return true, nil
}

func isColumn(node interface{}) bool {
	switch node.(type) {
	// it's wired that the "Subquery" type is also "VariableExpr" type
	// we have to ignore that case.
	case *tree.Subquery:
		return false
	case tree.VariableExpr:
		return true
	}
	return false
}

// ColNamesInSelect finds all referred variables in a Select Statement.
// (variables = sub-expressions, placeholders, indexed vars, etc.)
// Implementation limits:
//	1. Table with AS is not normalized.
//  2. Columns referred from outer query are not translated.
func ColNamesInSelect(sql string) (referredCols ReferredCols, err error) {
	referredCols = make(ReferredCols, 0)

	w := &AstWalker{
		Fn: func(ctx interface{}, node interface{}) (stop bool) {
			rCols := ctx.(ReferredCols)
			if isColumn(node) {
				nodeName := fmt.Sprint(node)
				// just drop the "table." part
				tableCols := strings.Split(nodeName, ".")
				colName := tableCols[len(tableCols)-1]
				rCols[colName] = 1
			}
			return false
		},
	}
	stmts, err := parser.Parse(sql)
	if err != nil {
		return
	}

	_, err = w.Walk(stmts, referredCols)
	if err != nil {
		return
	}
	for _, col := range w.UnknownNodes {
		log.Printf("unhandled column type %T", col)
	}
	return
}

func AllColsContained(set ReferredCols, cols []string) bool {
	if cols == nil {
		if set == nil {
			return true
		} else {
			return false
		}
	}
	if len(set) != len(cols) {
		return false
	}
	for _, col := range cols {
		if _, exist := set[col]; !exist {
			return false
		}
	}
	return true
}
