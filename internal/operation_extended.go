package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ============ ROLLUP/CUBE/GROUPING SETS Functions ============

// ExecuteRollupQuery executes a query with ROLLUP
func ExecuteRollupQuery(ctx context.Context, pool *pgxpool.Pool, tableName string, groupColumns []string, aggregateFunc string, aggregateColumn string) ([][]string, error) {
	if len(groupColumns) == 0 {
		return nil, fmt.Errorf("at least one grouping column required")
	}

	qb := NewQueryBuilder(tableName)
	qb.Select(groupColumns...)
	qb.Aggregate(aggregateColumn, AggregateFunc(aggregateFunc))
	qb.Rollup(groupColumns...)
	qb.OrderBy(groupColumns...)

	return qb.Execute(ctx, pool)
}

// ExecuteCubeQuery executes a query with CUBE
func ExecuteCubeQuery(ctx context.Context, pool *pgxpool.Pool, tableName string, groupColumns []string, aggregateFunc string, aggregateColumn string) ([][]string, error) {
	if len(groupColumns) == 0 {
		return nil, fmt.Errorf("at least one grouping column required")
	}

	qb := NewQueryBuilder(tableName)
	qb.Select(groupColumns...)
	qb.Aggregate(aggregateColumn, AggregateFunc(aggregateFunc))
	qb.Cube(groupColumns...)
	qb.OrderBy(groupColumns...)

	return qb.Execute(ctx, pool)
}

// ExecuteGroupingSetsQuery executes a query with GROUPING SETS
func ExecuteGroupingSetsQuery(ctx context.Context, pool *pgxpool.Pool, tableName string, sets [][]string, aggregateFunc string, aggregateColumn string) ([][]string, error) {
	if len(sets) == 0 {
		return nil, fmt.Errorf("at least one grouping set required")
	}

	qb := NewQueryBuilder(tableName)
	qb.Select("*")
	qb.Aggregate(aggregateColumn, AggregateFunc(aggregateFunc))
	qb.GroupingSets(sets...)

	return qb.Execute(ctx, pool)
}

// ============ CTE (WITH) Functions ============

// ExecuteCTEQuery executes a query with Common Table Expressions
func ExecuteCTEQuery(ctx context.Context, pool *pgxpool.Pool, cteDefinitions []CTEDefinition, mainQuery *QueryBuilder) ([][]string, error) {
	if len(cteDefinitions) == 0 {
		return nil, fmt.Errorf("at least one CTE definition required")
	}
	if mainQuery == nil {
		return nil, fmt.Errorf("main query cannot be nil")
	}

	qbc := &QueryBuilderWithCTE{
		CTEs: cteDefinitions,
		QB:   mainQuery,
	}

	sql := qbc.Build()
	log.Printf("CTE Query: %s", sql)

	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to execute CTE query: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	var result [][]string
	var header []string

	for _, fd := range fieldDescriptions {
		header = append(header, string(fd.Name))
	}
	result = append(result, header)

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("error reading row: %w", err)
		}
		var rowData []string
		for _, v := range values {
			if v == nil {
				rowData = append(rowData, "NULL")
			} else {
				rowData = append(rowData, fmt.Sprintf("%v", v))
			}
		}
		result = append(result, rowData)
	}

	return result, nil
}

// ============ Helper function for QueryBuilderWithCTE ============

// Build method implementation (completing the earlier partial implementation)
func (qbc *QueryBuilderWithCTE) BuildComplete() string {
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
