package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryBuilder построитель для создания сложных SQL запросов
type QueryBuilder struct {
	table      string
	columns    []string
	conditions []string
	groupBy    []string
	having     []string
	orderBy    []string
	aggregates map[string]string // колонка -> агрегатная функция
	limit      int
	offset     int
	joins      []JoinClause // Для поддержки JOIN
}

// JoinClause структура для хранения информации о JOIN
type JoinClause struct {
	Type        string // INNER, LEFT, RIGHT, FULL
	Table       string
	OnCondition string
}

// AggregateFunc типы агрегатных функций
type AggregateFunc string

const (
	Count AggregateFunc = "COUNT"
	Sum   AggregateFunc = "SUM"
	Avg   AggregateFunc = "AVG"
	Min   AggregateFunc = "MIN"
	Max   AggregateFunc = "MAX"
)

// RegexType тип регулярного выражения
type RegexType string

const (
	RegexMatch          AggregateFunc = "~"   // Соответствует регулярному выражению
	RegexMatchNoCase    AggregateFunc = "~*"  // Соответствует без учета регистра
	RegexNotMatch       AggregateFunc = "!~"  // Не соответствует регулярному выражению
	RegexNotMatchNoCase AggregateFunc = "!~*" // Не соответствует без учета регистра
)

// NewQueryBuilder создаёт новый построитель запросов
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:      table,
		columns:    []string{},
		conditions: []string{},
		groupBy:    []string{},
		having:     []string{},
		orderBy:    []string{},
		aggregates: make(map[string]string),
		joins:      []JoinClause{},
	}
}

// ===== SELECT методы =====

// Select указывает столбцы для выбора
// Если columns пусто, выбираются все (*)
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		qb.columns = append(qb.columns, columns...)
	}
	return qb
}

// ===== WHERE методы базовые =====

// Where добавляет условие WHERE
// Пример: Where("age > 18 AND status = 'active'")
func (qb *QueryBuilder) Where(condition string) *QueryBuilder {
	if strings.TrimSpace(condition) != "" {
		qb.conditions = append(qb.conditions, condition)
	}
	return qb
}

// WhereEq добавляет простое условие равенства
// Пример: WhereEq("status", "active")
func (qb *QueryBuilder) WhereEq(column, value string) *QueryBuilder {
	condition := fmt.Sprintf("%s = '%s'", column, value)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereGt добавляет условие больше
// Пример: WhereGt("price", "100")
func (qb *QueryBuilder) WhereGt(column, value string) *QueryBuilder {
	condition := fmt.Sprintf("%s > %s", column, value)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereLt добавляет условие меньше
// Пример: WhereLt("price", "1000")
func (qb *QueryBuilder) WhereLt(column, value string) *QueryBuilder {
	condition := fmt.Sprintf("%s < %s", column, value)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereGte добавляет условие больше или равно
func (qb *QueryBuilder) WhereGte(column, value string) *QueryBuilder {
	condition := fmt.Sprintf("%s >= %s", column, value)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereLte добавляет условие меньше или равно
func (qb *QueryBuilder) WhereLte(column, value string) *QueryBuilder {
	condition := fmt.Sprintf("%s <= %s", column, value)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereLike добавляет условие LIKE для поиска по шаблону
// Пример: WhereLike("name", "%товар%")
func (qb *QueryBuilder) WhereLike(column, pattern string) *QueryBuilder {
	condition := fmt.Sprintf("%s LIKE '%s'", column, pattern)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereIn добавляет условие IN
// Пример: WhereIn("status", "'active', 'pending'")
func (qb *QueryBuilder) WhereIn(column, values string) *QueryBuilder {
	condition := fmt.Sprintf("%s IN (%s)", column, values)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// ===== WHERE методы для POSIX регулярных выражений =====

// WhereRegex добавляет условие для POSIX регулярного выражения (~)
// Пример: WhereRegex("name", "^A.*")
func (qb *QueryBuilder) WhereRegex(column, pattern string) *QueryBuilder {
	condition := fmt.Sprintf("%s ~ '%s'", column, pattern)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereRegexNoCase добавляет условие для POSIX без учета регистра (~*)
// Пример: WhereRegexNoCase("name", "^a.*")
func (qb *QueryBuilder) WhereRegexNoCase(column, pattern string) *QueryBuilder {
	condition := fmt.Sprintf("%s ~* '%s'", column, pattern)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereNotRegex добавляет условие отрицания POSIX (!~)
// Пример: WhereNotRegex("name", "^B.*")
func (qb *QueryBuilder) WhereNotRegex(column, pattern string) *QueryBuilder {
	condition := fmt.Sprintf("%s !~ '%s'", column, pattern)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// WhereNotRegexNoCase добавляет условие отрицания POSIX без учета регистра (!~*)
// Пример: WhereNotRegexNoCase("name", "^b.*")
func (qb *QueryBuilder) WhereNotRegexNoCase(column, pattern string) *QueryBuilder {
	condition := fmt.Sprintf("%s !~* '%s'", column, pattern)
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// ===== GROUP BY и агрегатные функции =====

// GroupBy добавляет группировку по столбцам
// Пример: GroupBy("category", "status")
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// Aggregate добавляет агрегатную функцию
// Пример: Aggregate("total_price", Sum) добавит SUM(total_price) в SELECT
func (qb *QueryBuilder) Aggregate(column string, fn AggregateFunc) *QueryBuilder {
	qb.aggregates[column] = string(fn)
	return qb
}

// Having добавляет условие HAVING для фильтрации групп
// Пример: Having("COUNT(*) > 5")
func (qb *QueryBuilder) Having(condition string) *QueryBuilder {
	if strings.TrimSpace(condition) != "" {
		qb.having = append(qb.having, condition)
	}
	return qb
}

// ===== ORDER BY методы =====

// OrderBy добавляет сортировку
// Пример: OrderBy("name ASC", "price DESC")
func (qb *QueryBuilder) OrderBy(columns ...string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, columns...)
	return qb
}

// OrderByAsc добавляет сортировку по возрастанию
func (qb *QueryBuilder) OrderByAsc(columns ...string) *QueryBuilder {
	for _, col := range columns {
		qb.orderBy = append(qb.orderBy, col+" ASC")
	}
	return qb
}

// OrderByDesc добавляет сортировку по убыванию
func (qb *QueryBuilder) OrderByDesc(columns ...string) *QueryBuilder {
	for _, col := range columns {
		qb.orderBy = append(qb.orderBy, col+" DESC")
	}
	return qb
}

// ===== LIMIT и OFFSET =====

// Limit устанавливает количество возвращаемых записей
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset устанавливает смещение
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// ===== JOIN методы =====

// InnerJoin добавляет INNER JOIN
// Пример: InnerJoin("categories", "products.category_id = categories.id")
func (qb *QueryBuilder) InnerJoin(table, onCondition string) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Type:        "INNER",
		Table:       table,
		OnCondition: onCondition,
	})
	return qb
}

// LeftJoin добавляет LEFT JOIN
// Пример: LeftJoin("categories", "products.category_id = categories.id")
func (qb *QueryBuilder) LeftJoin(table, onCondition string) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Type:        "LEFT",
		Table:       table,
		OnCondition: onCondition,
	})
	return qb
}

// RightJoin добавляет RIGHT JOIN
// Пример: RightJoin("categories", "products.category_id = categories.id")
func (qb *QueryBuilder) RightJoin(table, onCondition string) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Type:        "RIGHT",
		Table:       table,
		OnCondition: onCondition,
	})
	return qb
}

// FullJoin добавляет FULL OUTER JOIN
// Пример: FullJoin("categories", "products.category_id = categories.id")
func (qb *QueryBuilder) FullJoin(table, onCondition string) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Type:        "FULL",
		Table:       table,
		OnCondition: onCondition,
	})
	return qb
}

// ===== Функции работы со строками =====

// SelectUpper преобразует столбец в верхний регистр
// Пример: SelectUpper("name") добавит UPPER(name) в SELECT
func (qb *QueryBuilder) SelectUpper(column string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("UPPER(%s)", column))
	return qb
}

// SelectLower преобразует столбец в нижний регистр
// Пример: SelectLower("name") добавит LOWER(name) в SELECT
func (qb *QueryBuilder) SelectLower(column string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("LOWER(%s)", column))
	return qb
}

// SelectTrim удаляет пробелы с обеих сторон
// Пример: SelectTrim("name") добавит TRIM(name) в SELECT
func (qb *QueryBuilder) SelectTrim(column string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("TRIM(%s)", column))
	return qb
}

// SelectLTrim удаляет пробелы слева
// Пример: SelectLTrim("name") добавит LTRIM(name) в SELECT
func (qb *QueryBuilder) SelectLTrim(column string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("LTRIM(%s)", column))
	return qb
}

// SelectRTrim удаляет пробелы справа
// Пример: SelectRTrim("name") добавит RTRIM(name) в SELECT
func (qb *QueryBuilder) SelectRTrim(column string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("RTRIM(%s)", column))
	return qb
}

// SelectSubstring извлекает подстроку
// Пример: SelectSubstring("name", 1, 3) добавит SUBSTRING(name, 1, 3) в SELECT
func (qb *QueryBuilder) SelectSubstring(column string, start, length int) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("SUBSTRING(%s, %d, %d)", column, start, length))
	return qb
}

// SelectLPad дополняет строку слева
// Пример: SelectLPad("id", 5, "0") добавит LPAD(id, 5, "0") в SELECT
func (qb *QueryBuilder) SelectLPad(column string, length int, padChar string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("LPAD(%s, %d, '%s')", column, length, padChar))
	return qb
}

// SelectRPad дополняет строку справа
// Пример: SelectRPad("id", 5, "0") добавит RPAD(id, 5, "0") в SELECT
func (qb *QueryBuilder) SelectRPad(column string, length int, padChar string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("RPAD(%s, %d, '%s')", column, length, padChar))
	return qb
}

// SelectConcat объединяет строки через ||
// Пример: SelectConcat("name", "email") добавит name || ' ' || email в SELECT
func (qb *QueryBuilder) SelectConcat(columns ...string) *QueryBuilder {
	if len(columns) == 0 {
		return qb
	}
	concatExpr := strings.Join(columns, " || ' ' || ")
	qb.columns = append(qb.columns, concatExpr)
	return qb
}

// ===== BUILD SQL =====

// Build строит SQL запрос
func (qb *QueryBuilder) Build() string {
	var query strings.Builder

	// SELECT часть
	query.WriteString("SELECT ")
	if len(qb.columns) == 0 && len(qb.aggregates) == 0 {
		// Если не указаны столбцы и агрегаты, выбираем все
		query.WriteString("*")
	} else {
		// Если указаны столбцы, добавляем их
		if len(qb.columns) > 0 {
			query.WriteString(strings.Join(qb.columns, ", "))
		}

		// Добавляем агрегатные функции
		if len(qb.aggregates) > 0 {
			if len(qb.columns) > 0 {
				query.WriteString(", ")
			}

			var aggFunctions []string
			for col, fn := range qb.aggregates {
				aggFunctions = append(aggFunctions, fmt.Sprintf("%s(%s)", fn, col))
			}
			query.WriteString(strings.Join(aggFunctions, ", "))
		}
	}

	// FROM часть
	query.WriteString(" FROM ")
	query.WriteString(qb.table)

	// JOIN части
	for _, join := range qb.joins {
		query.WriteString(" ")
		query.WriteString(join.Type)
		query.WriteString(" JOIN ")
		query.WriteString(join.Table)
		query.WriteString(" ON ")
		query.WriteString(join.OnCondition)
	}

	// WHERE часть
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.conditions, " AND "))
	}

	// GROUP BY часть
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING часть
	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		query.WriteString(strings.Join(qb.having, " AND "))
	}

	// ORDER BY часть
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// LIMIT часть
	if qb.limit > 0 {
		query.WriteString(" LIMIT ")
		query.WriteString(fmt.Sprintf("%d", qb.limit))
	}

	// OFFSET часть
	if qb.offset > 0 {
		query.WriteString(" OFFSET ")
		query.WriteString(fmt.Sprintf("%d", qb.offset))
	}

	return query.String()
}

// Execute выполняет запрос и возвращает результаты в виде [][]string
func (qb *QueryBuilder) Execute(ctx context.Context, pool *pgxpool.Pool) ([][]string, error) {
	sql := qb.Build()
	log.Printf("Выполняемый SQL: %s", sql)
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	// Получаем описание столбцов
	fieldDescriptions := rows.FieldDescriptions()
	var result [][]string

	// Заголовок
	var header []string
	for _, fd := range fieldDescriptions {
		header = append(header, string(fd.Name))
	}
	result = append(result, header)

	// Данные
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %w", err)
		}

		var row []string
		for _, v := range values {
			row = append(row, fmt.Sprintf("%v", v))
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return result, nil
}

// String возвращает SQL запрос в виде строки
func (qb *QueryBuilder) String() string {
	return qb.Build()
}

// ===== Вспомогательные функции =====

// GetColumns возвращает список выбранных столбцов
func (qb *QueryBuilder) GetColumns() []string {
	return qb.columns
}

// GetConditions возвращает список условий WHERE
func (qb *QueryBuilder) GetConditions() []string {
	return qb.conditions
}

// GetGroupBy возвращает список столбцов GROUP BY
func (qb *QueryBuilder) GetGroupBy() []string {
	return qb.groupBy
}

// GetOrderBy возвращает список столбцов ORDER BY
func (qb *QueryBuilder) GetOrderBy() []string {
	return qb.orderBy
}

// GetHaving возвращает список условий HAVING
func (qb *QueryBuilder) GetHaving() []string {
	return qb.having
}

// GetAggregates возвращает карту агрегатных функций
func (qb *QueryBuilder) GetAggregates() map[string]string {
	return qb.aggregates
}

// GetJoins возвращает список JOIN clauses
func (qb *QueryBuilder) GetJoins() []JoinClause {
	return qb.joins
}

// Reset сбрасывает все условия (но сохраняет таблицу)
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.columns = []string{}
	qb.conditions = []string{}
	qb.groupBy = []string{}
	qb.having = []string{}
	qb.orderBy = []string{}
	qb.aggregates = make(map[string]string)
	qb.joins = []JoinClause{}
	qb.limit = 0
	qb.offset = 0
	return qb
}

// Copy создаёт копию текущего построителя
func (qb *QueryBuilder) Copy() *QueryBuilder {
	newQB := NewQueryBuilder(qb.table)
	newQB.columns = append(newQB.columns, qb.columns...)
	newQB.conditions = append(newQB.conditions, qb.conditions...)
	newQB.groupBy = append(newQB.groupBy, qb.groupBy...)
	newQB.having = append(newQB.having, qb.having...)
	newQB.orderBy = append(newQB.orderBy, qb.orderBy...)
	newQB.joins = append(newQB.joins, qb.joins...)
	newQB.limit = qb.limit
	newQB.offset = qb.offset

	// Копируем агрегаты
	for k, v := range qb.aggregates {
		newQB.aggregates[k] = v
	}

	return newQB
}
