package internal

import (
	"fmt"
	"strings"
)

// WindowFunction структура для хранения оконной функции
type WindowFunction struct {
	Alias        string        // Имя результирующей колонки
	Function     string        // RANK, LAG, LEAD
	TargetColumn string        // Для LAG/LEAD - колонка, по которой брать значение
	Offset       int           // Для LAG/LEAD - смещение (по умолчанию 1)
	PartitionBy  []string      // Колонки для PARTITION BY
	OrderBy      []OrderByItem // Колонки и направления для ORDER BY
}

// OrderByItem для хранения колонки и её направления сортировки
type OrderByItem struct {
	Column    string // Имя колонки
	Direction string // ASC или DESC (по умолчанию ASC)
}

// windowFunctions слайс оконных функций в QueryBuilder
// (добавить в структуру QueryBuilder)

// SelectRank добавляет оконную функцию RANK()
// Пример: qb.SelectRank("row_rank", []string{"category"}, []OrderByItem{{Column: "price", Direction: "DESC"}})
// Генерирует: SELECT ..., RANK() OVER (PARTITION BY category ORDER BY price DESC) AS row_rank
func (qb *QueryBuilder) SelectRank(alias string, partitionBy []string, orderBy []OrderByItem) *QueryBuilder {
	if alias == "" {
		alias = "rank"
	}

	wf := WindowFunction{
		Alias:       alias,
		Function:    "RANK",
		PartitionBy: partitionBy,
		OrderBy:     orderBy,
	}

	// Инициализируем слайс если это первый раз
	if qb.windowFunctions == nil {
		qb.windowFunctions = make([]WindowFunction, 0)
	}

	qb.windowFunctions = append(qb.windowFunctions, wf)
	return qb
}

// SelectLag добавляет оконную функцию LAG()
// Пример: qb.SelectLag("price", 1, "prev_price", []string{"category"}, []OrderByItem{{Column: "date", Direction: "ASC"}})
// Генерирует: SELECT ..., LAG(price, 1) OVER (PARTITION BY category ORDER BY date ASC) AS prev_price
func (qb *QueryBuilder) SelectLag(targetColumn string, offset int, alias string, partitionBy []string, orderBy []OrderByItem) *QueryBuilder {
	if alias == "" {
		alias = fmt.Sprintf("lag_%s", targetColumn)
	}

	if offset <= 0 {
		offset = 1
	}

	wf := WindowFunction{
		Alias:        alias,
		Function:     "LAG",
		TargetColumn: targetColumn,
		Offset:       offset,
		PartitionBy:  partitionBy,
		OrderBy:      orderBy,
	}

	if qb.windowFunctions == nil {
		qb.windowFunctions = make([]WindowFunction, 0)
	}

	qb.windowFunctions = append(qb.windowFunctions, wf)
	return qb
}

// SelectLead добавляет оконную функцию LEAD()
// Пример: qb.SelectLead("price", 1, "next_price", []string{"category"}, []OrderByItem{{Column: "date", Direction: "ASC"}})
// Генерирует: SELECT ..., LEAD(price, 1) OVER (PARTITION BY category ORDER BY date ASC) AS next_price
func (qb *QueryBuilder) SelectLead(targetColumn string, offset int, alias string, partitionBy []string, orderBy []OrderByItem) *QueryBuilder {
	if alias == "" {
		alias = fmt.Sprintf("lead_%s", targetColumn)
	}

	if offset <= 0 {
		offset = 1
	}

	wf := WindowFunction{
		Alias:        alias,
		Function:     "LEAD",
		TargetColumn: targetColumn,
		Offset:       offset,
		PartitionBy:  partitionBy,
		OrderBy:      orderBy,
	}

	if qb.windowFunctions == nil {
		qb.windowFunctions = make([]WindowFunction, 0)
	}

	qb.windowFunctions = append(qb.windowFunctions, wf)
	return qb
}

// buildWindowFunction генерирует SQL для оконной функции
func (wf WindowFunction) buildSQL() string {
	var funcSQL string

	switch wf.Function {
	case "RANK":
		funcSQL = "RANK()"
	case "LAG":
		funcSQL = fmt.Sprintf("LAG(%s, %d)", wf.TargetColumn, wf.Offset)
	case "LEAD":
		funcSQL = fmt.Sprintf("LEAD(%s, %d)", wf.TargetColumn, wf.Offset)
	default:
		return ""
	}

	// Строим OVER clause
	var overParts []string

	if len(wf.PartitionBy) > 0 {
		partitionStr := "PARTITION BY " + strings.Join(wf.PartitionBy, ", ")
		overParts = append(overParts, partitionStr)
	}

	if len(wf.OrderBy) > 0 {
		var orderByStrs []string
		for _, item := range wf.OrderBy {
			direction := item.Direction
			if direction == "" {
				direction = "ASC"
			}
			orderByStrs = append(orderByStrs, fmt.Sprintf("%s %s", item.Column, direction))
		}
		orderByStr := "ORDER BY " + strings.Join(orderByStrs, ", ")
		overParts = append(overParts, orderByStr)
	}

	overClause := strings.Join(overParts, " ")

	// Если нет PARTITION BY и ORDER BY, то просто пустой OVER()
	if overClause == "" {
		overClause = "OVER ()"
	} else {
		overClause = fmt.Sprintf("OVER (%s)", overClause)
	}

	return fmt.Sprintf("%s %s AS %s", funcSQL, overClause, wf.Alias)
}

// GetWindowFunctions возвращает слайс оконных функций
func (qb *QueryBuilder) GetWindowFunctions() []WindowFunction {
	return qb.windowFunctions
}

// ClearWindowFunctions очищает оконные функции
func (qb *QueryBuilder) ClearWindowFunctions() *QueryBuilder {
	qb.windowFunctions = make([]WindowFunction, 0)
	return qb
}
