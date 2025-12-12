package internal

import (
	"fmt"
	"strings"
)

// ROLLUP support
type RollupClause struct {
	Columns []string
}

// CUBE support
type CubeClause struct {
	Columns []string
}

// GROUPING SETS support
type GroupingSetClause struct {
	Sets [][]string
}

// CTE support (Common Table Expressions)
type CTEDefinition struct {
	Name    string
	Query   string
	Columns []string
}

// Add ROLLUP support to QueryBuilder
func (qb *QueryBuilder) Rollup(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		rollupExpr := "ROLLUP(" + strings.Join(columns, ", ") + ")"
		qb.groupBy = append(qb.groupBy, rollupExpr)
	}
	return qb
}

// Add CUBE support to QueryBuilder
func (qb *QueryBuilder) Cube(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		cubeExpr := "CUBE(" + strings.Join(columns, ", ") + ")"
		qb.groupBy = append(qb.groupBy, cubeExpr)
	}
	return qb
}

// Add GROUPING SETS support to QueryBuilder
func (qb *QueryBuilder) GroupingSets(sets ...[]string) *QueryBuilder {
	if len(sets) > 0 {
		var setClauses []string
		for _, set := range sets {
			if len(set) > 0 {
				setClauses = append(setClauses, "("+strings.Join(set, ", ")+")")
			}
		}
		if len(setClauses) > 0 {
			groupingExpr := "GROUPING SETS (" + strings.Join(setClauses, ", ") + ")"
			qb.groupBy = append(qb.groupBy, groupingExpr)
		}
	}
	return qb
}

// CTE field in QueryBuilder
type QueryBuilderWithCTE struct {
	CTEs []CTEDefinition
	QB   *QueryBuilder
}

// Create new QueryBuilder with CTE support
func NewQueryBuilderWithCTE(table string) *QueryBuilderWithCTE {
	return &QueryBuilderWithCTE{
		CTEs: []CTEDefinition{},
		QB:   NewQueryBuilder(table),
	}
}

// Add CTE definition
func (qbc *QueryBuilderWithCTE) AddCTE(name string, query string, columns ...string) *QueryBuilderWithCTE {
	qbc.CTEs = append(qbc.CTEs, CTEDefinition{
		Name:    name,
		Query:   query,
		Columns: columns,
	})
	return qbc
}

// Build complete query with CTEs
func (qbc *QueryBuilderWithCTE) Build() string {
	if len(qbc.CTEs) == 0 {
		return qbc.QB.Build()
	}

	var result strings.Builder
	result.WriteString("WITH ")

	for i, cte := range qbc.CTEs {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(cte.Name)
		if len(cte.Columns) > 0 {
			result.WriteString(" (" + strings.Join(cte.Columns, ", ") + ")")
		}
		result.WriteString(" AS (" + cte.Query + ")")
	}

	result.WriteString(" " + qbc.QB.Build())
	return result.String()
}

// GROUPING function support (returns 0 or 1 depending on grouping)
func (qb *QueryBuilder) SelectGrouping(column string, alias string) *QueryBuilder {
	if alias != "" {
		alias = " AS " + alias
	}
	qb.columns = append(qb.columns, fmt.Sprintf("GROUPING(%s)%s", column, alias))
	return qb
}

// GROUPING_ID function support (combines multiple GROUPING results)
func (qb *QueryBuilder) SelectGroupingID(columns []string, alias string) *QueryBuilder {
	if len(columns) == 0 {
		return qb
	}
	expr := "GROUPING_ID(" + strings.Join(columns, ", ") + ")"
	if alias != "" {
		expr = expr + " AS " + alias
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// Utility function for complex aggregations with FILTER clause (PostgreSQL)
func (qb *QueryBuilder) SelectAggregateWithFilter(aggFunc string, column string, filterCondition string, alias string) *QueryBuilder {
	expr := fmt.Sprintf("%s(%s) FILTER (WHERE %s)", aggFunc, column, filterCondition)
	if alias != "" {
		expr = expr + " AS " + alias
	}
	qb.columns = append(qb.columns, expr)
	return qb
}
