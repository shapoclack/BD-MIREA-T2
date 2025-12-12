package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateEnumType создаёт новый ENUM тип
func CreateEnumType(ctx context.Context, pool *pgxpool.Pool, typeName string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("список значений ENUM не может быть пустым")
	}

	// Валидация имени типа
	if err := validateSQLIdent(typeName); err != nil {
		return err
	}

	// Форматируем значения с кавычками
	var formattedValues []string
	for _, val := range values {
		if err := validateSQLIdent(val); err != nil {
			return fmt.Errorf("недопустимое значение ENUM '%s': %w", val, err)
		}
		formattedValues = append(formattedValues, fmt.Sprintf("'%s'", val))
	}

	query := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", typeName, strings.Join(formattedValues, ", "))

	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Ошибка создания ENUM типа: %v", err)
		return fmt.Errorf("ошибка создания ENUM типа %s: %w", typeName, err)
	}

	fmt.Printf("ENUM тип '%s' успешно создан с %d значениями\n", typeName, len(values))
	return nil
}

// CreateCompositeType создаёт составной (composite) тип
// Пример: CreateCompositeType(ctx, pool, "address_type",
//
//	map[string]string{"street": "VARCHAR(255)", "city": "VARCHAR(100)", "postal_code": "VARCHAR(10)"})
func CreateCompositeType(ctx context.Context, pool *pgxpool.Pool, typeName string, fields map[string]string) error {
	if len(fields) == 0 {
		return fmt.Errorf("список полей составного типа не может быть пустым")
	}

	// Валидация имени типа
	if err := validateSQLIdent(typeName); err != nil {
		return err
	}

	// Составляем определение полей
	var fieldDefinitions []string
	for fieldName, fieldType := range fields {
		if err := validateSQLIdent(fieldName); err != nil {
			return fmt.Errorf("недопустимое имя поля '%s': %w", fieldName, err)
		}
		fieldDefinitions = append(fieldDefinitions, fmt.Sprintf("%s %s", fieldName, fieldType))
	}

	query := fmt.Sprintf("CREATE TYPE %s AS (%s)", typeName, strings.Join(fieldDefinitions, ", "))

	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Ошибка создания составного типа: %v", err)
		return fmt.Errorf("ошибка создания составного типа %s: %w", typeName, err)
	}

	fmt.Printf("Составной тип '%s' успешно создан с %d полями\n", typeName, len(fields))
	return nil
}

// DropEnumType удаляет ENUM тип
// Пример: DropEnumType(ctx, pool, "status_enum")
func DropEnumType(ctx context.Context, pool *pgxpool.Pool, typeName string) error {
	if err := validateSQLIdent(typeName); err != nil {
		return err
	}

	query := fmt.Sprintf("DROP TYPE IF EXISTS %s CASCADE", typeName)

	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Ошибка удаления типа: %v", err)
		return fmt.Errorf("ошибка удаления типа %s: %w", typeName, err)
	}

	fmt.Printf("Тип '%s' успешно удален\n", typeName)
	return nil
}

// GetCustomTypes получает список всех пользовательских типов
// Возвращает: []map[string]string с полями {typeName, typeKind, definition}
func GetCustomTypes(ctx context.Context, pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `
	SELECT
		t.typname as type_name,
		CASE t.typtype
			WHEN 'e' THEN 'ENUM'
			WHEN 'c' THEN 'COMPOSITE'
			WHEN 'b' THEN 'BASE'
			ELSE 'OTHER'
		END as type_kind
	FROM pg_type t
	JOIN pg_namespace n ON n.oid = t.typnamespace
	WHERE n.nspname = 'public'
	AND t.typtype IN ('e', 'c')
	ORDER BY t.typname
`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения типов: %w", err)
	}
	defer rows.Close()

	var types []map[string]interface{}

	for rows.Next() {
		var typeName, typeKind string

		err := rows.Scan(&typeName, &typeKind)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения типа: %w", err)
		}

		typeInfo := map[string]interface{}{
			"type_name": typeName,
			"type_kind": typeKind,
		}

		types = append(types, typeInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по типам: %w", err)
	}

	return types, nil
}

// GetEnumValues получает все значения для конкретного ENUM типа
// Пример: GetEnumValues(ctx, pool, "status_enum")
func GetEnumValues(ctx context.Context, pool *pgxpool.Pool, enumTypeName string) ([]string, error) {
	if err := validateSQLIdent(enumTypeName); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
	SELECT enumlabel
	FROM pg_enum
	JOIN pg_type ON pg_enum.enumtypid = pg_type.oid
	WHERE pg_type.typname = '%s'
	ORDER BY enumsortorder
	`, enumTypeName)

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения значений ENUM: %w", err)
	}
	defer rows.Close()

	var values []string

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("ошибка чтения значения ENUM: %w", err)
		}
		values = append(values, value)
	}

	return values, rows.Err()
}

// GetCompositeTypeFields получает все поля для составного типа
// Пример: GetCompositeTypeFields(ctx, pool, "address_type")
// Возвращает: map[fieldName]fieldType
func GetCompositeTypeFields(ctx context.Context, pool *pgxpool.Pool, compositeTypeName string) (map[string]string, error) {
	if err := validateSQLIdent(compositeTypeName); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
	SELECT 
		a.attname as field_name,
		pg_catalog.format_type(a.atttypid, a.atttypmod) as field_type
	FROM pg_attribute a
	JOIN pg_type t ON a.attrelid = t.typrelid
	WHERE t.typname = '%s'
	AND a.attnum > 0
	ORDER BY a.attnum
	`, compositeTypeName)

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения полей типа: %w", err)
	}
	defer rows.Close()

	fields := make(map[string]string)

	for rows.Next() {
		var fieldName, fieldType string
		if err := rows.Scan(&fieldName, &fieldType); err != nil {
			return nil, fmt.Errorf("ошибка чтения поля: %w", err)
		}
		fields[fieldName] = fieldType
	}

	return fields, rows.Err()
}

// AddEnumValue добавляет новое значение к существующему ENUM типу
// Пример: AddEnumValue(ctx, pool, "status_enum", "archived", "before_value")
func AddEnumValue(ctx context.Context, pool *pgxpool.Pool, enumTypeName, newValue, beforeValue string) error {
	if err := validateSQLIdent(enumTypeName); err != nil {
		return err
	}

	if err := validateSQLIdent(newValue); err != nil {
		return fmt.Errorf("недопустимое значение ENUM: %w", err)
	}

	query := fmt.Sprintf("ALTER TYPE %s ADD VALUE '%s'", enumTypeName, newValue)

	if beforeValue != "" {
		if err := validateSQLIdent(beforeValue); err != nil {
			return fmt.Errorf("недопустимое значение beforeValue: %w", err)
		}
		query += fmt.Sprintf(" BEFORE '%s'", beforeValue)
	}

	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Ошибка добавления значения в ENUM: %v", err)
		return fmt.Errorf("ошибка добавления значения '%s' в ENUM '%s': %w", newValue, enumTypeName, err)
	}

	fmt.Printf("Значение '%s' успешно добавлено в ENUM '%s'\n", newValue, enumTypeName)
	return nil
}

// TypeInfo получает информацию о типе
type TypeInfo struct {
	Name        string
	Kind        string            // ENUM, COMPOSITE, BASE
	Values      []string          // для ENUM
	Fields      map[string]string // для COMPOSITE
	Description string
}

// GetTypeInfo получает полную информацию о типе
func GetTypeInfo(ctx context.Context, pool *pgxpool.Pool, typeName string) (*TypeInfo, error) {
	if err := validateSQLIdent(typeName); err != nil {
		return nil, err
	}

	types, err := GetCustomTypes(ctx, pool)
	if err != nil {
		return nil, err
	}

	var foundType map[string]interface{}
	for _, t := range types {
		if t["type_name"] == typeName {
			foundType = t
			break
		}
	}

	if foundType == nil {
		return nil, fmt.Errorf("тип '%s' не найден", typeName)
	}

	info := &TypeInfo{
		Name: typeName,
		Kind: foundType["type_kind"].(string),
	}

	if info.Kind == "ENUM" {
		values, err := GetEnumValues(ctx, pool, typeName)
		if err != nil {
			return nil, err
		}
		info.Values = values
	}

	if info.Kind == "COMPOSITE" {
		fields, err := GetCompositeTypeFields(ctx, pool, typeName)
		if err != nil {
			return nil, err
		}
		info.Fields = fields
	}

	return info, nil
}
