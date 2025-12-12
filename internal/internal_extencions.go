package internal

import (
	"fmt"
	"strings"
)

// ===== ИСПРАВЛЕНИЯ ФУНКЦИЙ COALESCE И NULLIF =====

// SelectCoalesce добавляет функцию COALESCE в SELECT
// COALESCE возвращает первое не-NULL значение из списка
// Пример: SelectCoalesce([]string{"description", "'No description'"}, "coalesce_desc")
// Генерирует: COALESCE(description, 'No description') AS coalesce_desc
func (qb *QueryBuilder) SelectCoalesceFixed(columns []string, alias string) *QueryBuilder {
	if len(columns) == 0 {
		return qb
	}

	coalesceSql := fmt.Sprintf("COALESCE(%s)", strings.Join(columns, ", "))
	if alias != "" {
		coalesceSql = fmt.Sprintf("%s AS %s", coalesceSql, alias)
	}

	qb.columns = append(qb.columns, coalesceSql)
	return qb
}

// SelectNullif добавляет функцию NULLIF в SELECT
// NULLIF возвращает NULL если оба значения равны, иначе возвращает первое значение
// Пример: SelectNullif("status", "'inactive'", "status_or_null")
// Генерирует: NULLIF(status, 'inactive') AS status_or_null
func (qb *QueryBuilder) SelectNullifFixed(column1, column2, alias string) *QueryBuilder {
	nullifSql := fmt.Sprintf("NULLIF(%s, %s)", column1, column2)
	if alias != "" {
		nullifSql = fmt.Sprintf("%s AS %s", nullifSql, alias)
	}

	qb.columns = append(qb.columns, nullifSql)
	return qb
}

// WhereCoalesceFixed добавляет условие WHERE с COALESCE
// Пример: WhereCoalesce([]string{"name", "'Unknown'"}, "=", "'John'")
// Генерирует: WHERE COALESCE(name, 'Unknown') = 'John'
func (qb *QueryBuilder) WhereCoalesceFixed(columns []string, operator, value string) *QueryBuilder {
	if len(columns) == 0 {
		return qb
	}

	condition := fmt.Sprintf("COALESCE(%s) %s %s", strings.Join(columns, ", "), operator, value)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// ===== РАСШИРЕННЫЕ ФУНКЦИИ ДЛЯ CASE ВЫРАЖЕНИЙ =====

// CaseSimple создаёт простое CASE выражение (CASE column WHEN value THEN result END)
// Пример: CaseSimple("status", map[string]string{"active": "'Активен'", "inactive": "'Неактивен'"}, "'Неизвестно'")
type SimpleCaseExpression struct {
	column string
	whens  map[string]string // value -> result
	else_  string
}

// NewSimpleCase создаёт новое простое CASE выражение
func NewSimpleCase(column string) *SimpleCaseExpression {
	return &SimpleCaseExpression{
		column: column,
		whens:  make(map[string]string),
	}
}

// When добавляет WHEN value THEN result
func (sce *SimpleCaseExpression) When(value, result string) *SimpleCaseExpression {
	sce.whens[value] = result
	return sce
}

// Else устанавливает ELSE значение
func (sce *SimpleCaseExpression) Else(value string) *SimpleCaseExpression {
	sce.else_ = value
	return sce
}

// Build собирает простое CASE выражение в SQL
func (sce *SimpleCaseExpression) Build() string {
	if len(sce.whens) == 0 {
		return ""
	}

	var caseStr strings.Builder
	caseStr.WriteString(fmt.Sprintf("CASE %s", sce.column))

	for value, result := range sce.whens {
		caseStr.WriteString(fmt.Sprintf(" WHEN %s THEN %s", value, result))
	}

	if sce.else_ != "" {
		caseStr.WriteString(fmt.Sprintf(" ELSE %s", sce.else_))
	}

	caseStr.WriteString(" END")
	return caseStr.String()
}

// SelectSimpleCase добавляет простое CASE выражение в SELECT
func (qb *QueryBuilder) SelectSimpleCase(caseExpr *SimpleCaseExpression, alias string) *QueryBuilder {
	caseSQL := caseExpr.Build()
	if caseSQL != "" {
		if alias != "" {
			caseSQL = fmt.Sprintf("%s AS %s", caseSQL, alias)
		}
		qb.columns = append(qb.columns, caseSQL)
	}
	return qb
}

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ С NULL =====

// WhereIsNull добавляет условие IS NULL
// Пример: WhereIsNull("description")
// Генерирует: WHERE description IS NULL
func (qb *QueryBuilder) WhereIsNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NULL", column)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereIsNotNull добавляет условие IS NOT NULL
// Пример: WhereIsNotNull("description")
// Генерирует: WHERE description IS NOT NULL
func (qb *QueryBuilder) WhereIsNotNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NOT NULL", column)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ С АГРЕГАТНЫМИ ФУНКЦИЯМИ И NULL =====

// AggregateWithCoalesce добавляет агрегатную функцию с COALESCE для обработки NULL
// Пример: AggregateWithCoalesce("price", Sum, "'0'") создаст SUM(COALESCE(price, '0'))
func (qb *QueryBuilder) AggregateWithCoalesce(column string, fn AggregateFunc, defaultValue string) *QueryBuilder {
	aggregateExpr := fmt.Sprintf("COALESCE(%s(%s), %s)", fn, column, defaultValue)
	qb.columns = append(qb.columns, aggregateExpr)
	return qb
}

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ СО СТРОКАМИ - ДОПОЛНЕНИЕ =====

// SelectCharLength возвращает длину строки
// Пример: SelectCharLength("name", "name_length")
func (qb *QueryBuilder) SelectCharLength(column string, alias string) *QueryBuilder {
	expr := fmt.Sprintf("CHAR_LENGTH(%s)", column)
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// SelectStringPosition возвращает позицию подстроки
// Пример: SelectStringPosition("name", "'John'", "pos")
func (qb *QueryBuilder) SelectStringPosition(column, substring, alias string) *QueryBuilder {
	expr := fmt.Sprintf("POSITION(%s IN %s)", substring, column)
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// SelectReplace заменяет подстроку
// Пример: SelectReplace("name", "'old'", "'new'", "modified_name")
func (qb *QueryBuilder) SelectReplace(column, from, to, alias string) *QueryBuilder {
	expr := fmt.Sprintf("REPLACE(%s, %s, %s)", column, from, to)
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ С ТИПАМИ ДАННЫХ =====

// SelectCast приводит столбец к другому типу
// Пример: SelectCast("price", "INTEGER", "price_int")
func (qb *QueryBuilder) SelectCast(column, targetType, alias string) *QueryBuilder {
	expr := fmt.Sprintf("CAST(%s AS %s)", column, targetType)
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ С ДАТОЙ И ВРЕМЕНЕМ =====

// SelectCurrentDate добавляет текущую дату
func (qb *QueryBuilder) SelectCurrentDate(alias string) *QueryBuilder {
	expr := "CURRENT_DATE"
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// SelectCurrentTimestamp добавляет текущее время
func (qb *QueryBuilder) SelectCurrentTimestamp(alias string) *QueryBuilder {
	expr := "CURRENT_TIMESTAMP"
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// SelectExtract извлекает часть из даты/времени
// Пример: SelectExtract("created_at", "YEAR", "year")
func (qb *QueryBuilder) SelectExtract(column, field, alias string) *QueryBuilder {
	expr := fmt.Sprintf("EXTRACT(%s FROM %s)", field, column)
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// ===== ФУНКЦИИ ДЛЯ УСЛОВНОЙ ЛОГИКИ =====

// SelectIf реализует простую IF логику (эквивалент CASE WHEN condition THEN true_val ELSE false_val END)
// Пример: SelectIf("price > 100", "'Дорогой'", "'Дешевый'", "price_category")
func (qb *QueryBuilder) SelectIf(condition, trueVal, falseVal, alias string) *QueryBuilder {
	expr := fmt.Sprintf("CASE WHEN %s THEN %s ELSE %s END", condition, trueVal, falseVal)
	if alias != "" {
		expr = fmt.Sprintf("%s AS %s", expr, alias)
	}
	qb.columns = append(qb.columns, expr)
	return qb
}

// ===== ФУНКЦИИ ДЛЯ РАСШИРЕННОЙ ФИЛЬТРАЦИИ =====

// WhereBetween добавляет условие BETWEEN
// Пример: WhereBetween("price", "10", "100")
// Генерирует: WHERE price BETWEEN 10 AND 100
func (qb *QueryBuilder) WhereBetween(column, start, end string) *QueryBuilder {
	condition := fmt.Sprintf("%s BETWEEN %s AND %s", column, start, end)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereNotBetween добавляет условие NOT BETWEEN
func (qb *QueryBuilder) WhereNotBetween(column, start, end string) *QueryBuilder {
	condition := fmt.Sprintf("%s NOT BETWEEN %s AND %s", column, start, end)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ С МАССИВАМИ И ТИПАМИ =====

// WhereArrayContains проверяет, содержит ли массив значение
// Пример: WhereArrayContains("tags", "'python'")
// Генерирует: WHERE 'python' = ANY(tags)
func (qb *QueryBuilder) WhereArrayContains(column, value string) *QueryBuilder {
	condition := fmt.Sprintf("%s = ANY(%s)", value, column)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// ===== ДОПОЛНИТЕЛЬНЫЕ МЕТОДЫ ПОСТРОЕНИЯ ЗАПРОСОВ =====

// Distinct добавляет DISTINCT
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	// Модифицируем Select для добавления DISTINCT
	if len(qb.columns) > 0 {
		qb.columns[0] = "DISTINCT " + qb.columns[0]
	}
	return qb
}

// SetSelectColumns переустанавливает выбранные столбцы
func (qb *QueryBuilder) SetSelectColumns(columns ...string) *QueryBuilder {
	qb.columns = []string{}
	if len(columns) > 0 {
		qb.columns = append(qb.columns, columns...)
	}
	return qb
}

// ClearConditions очищает все условия WHERE
func (qb *QueryBuilder) ClearConditions() *QueryBuilder {
	qb.conditions = []string{}
	return qb
}

// ClearGroupBy очищает GROUP BY
func (qb *QueryBuilder) ClearGroupBy() *QueryBuilder {
	qb.groupBy = []string{}
	return qb
}
