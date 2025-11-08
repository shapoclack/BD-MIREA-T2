package internal

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ColumnDefinition struct {
	Name        string
	Type        string
	Constraints string
}

func CreateTablesWithTypes(ctx context.Context, pool *pgxpool.Pool, tableName string, columns []ColumnDefinition) error {
	if len(columns) == 0 {
		return fmt.Errorf("список столбцов не может быть пустым")
	}

	// Формируем SQL для создания столбцов
	var columnDefinitions []string
	for _, col := range columns {
		// Базовое определение: имя и тип
		colDef := fmt.Sprintf("%s %s", col.Name, col.Type)

		// Добавляем ограничения, если они указаны
		if col.Constraints != "" {
			colDef += " " + col.Constraints
		}

		columnDefinitions = append(columnDefinitions, colDef)
	}

	// Создаём SQL запрос CREATE TABLE
	sql := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (\n\t%s\n)",
		tableName,
		strings.Join(columnDefinitions, ",\n\t"),
	)

	// Выполняем запрос
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("ошибка создания таблицы %s: %w", tableName, err)
	}

	fmt.Printf("Таблица '%s' успешно создана с %d столбцами\n", tableName, len(columns))
	return nil
}

func CreateTablesWithTypesAdvanced(ctx context.Context, pool *pgxpool.Pool, tableName string, columns []ColumnDefinition, tableConstraints []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("список столбцов не может быть пустым")
	}

	// Формируем SQL для создания столбцов
	var columnDefinitions []string
	for _, col := range columns {
		colDef := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.Constraints != "" {
			colDef += " " + col.Constraints
		}
		columnDefinitions = append(columnDefinitions, colDef)
	}

	// Добавляем ограничения уровня таблицы (например, FOREIGN KEY, CHECK, UNIQUE)
	allDefinitions := columnDefinitions
	if len(tableConstraints) > 0 {
		allDefinitions = append(allDefinitions, tableConstraints...)
	}

	// Создаём SQL запрос
	sql := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (\n\t%s\n)",
		tableName,
		strings.Join(allDefinitions, ",\n\t"),
	)

	// Выполняем запрос
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("ошибка создания таблицы %s: %w", tableName, err)
	}

	fmt.Printf("Таблица '%s' успешно создана с %d столбцами и %d ограничениями\n",
		tableName, len(columns), len(tableConstraints))
	return nil
}

// Создание основных таблиц для демонстрации
func CreateTables(ctx context.Context, pool *pgxpool.Pool) error {
	// Таблица категорий
	categoriesSQL := `
    CREATE TABLE IF NOT EXISTS categories (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL UNIQUE,
        description TEXT,
        created_at TIMESTAMP DEFAULT NOW()
    )`

	// Таблица продуктов с различными типами данных и ограничениями
	productsSQL := `
    CREATE TABLE IF NOT EXISTS products (
        id SERIAL PRIMARY KEY,
        name VARCHAR(200) NOT NULL,
        description TEXT,
        price NUMERIC(10,2) CHECK (price >= 0),
        quantity INTEGER DEFAULT 0 CHECK (quantity >= 0),
        is_active BOOLEAN DEFAULT true,
        category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL ON UPDATE CASCADE,
        tags TEXT[],
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW()
    )`

	// Создаем таблицы
	if _, err := pool.Exec(ctx, categoriesSQL); err != nil {
		return fmt.Errorf("ошибка создания таблицы categories: %w", err)
	}

	if _, err := pool.Exec(ctx, productsSQL); err != nil {
		return fmt.Errorf("ошибка создания таблицы products: %w", err)
	}

	// Добавляем только базовые категории (без продуктов)
	if err := insertInitialCategories(ctx, pool); err != nil {
		return fmt.Errorf("ошибка добавления категорий: %w", err)
	}

	fmt.Println("Таблицы успешно созданы")
	return nil
}

// alter table 2.1
func validateSQLIdent(name string) error {
	if name == "" {
		return fmt.Errorf("Индификатор не может быть пустым")
	}
	matched, err := regexp.MatchString(`^[A-Za-z_][A-Za-z0-9_]*$`, name)
	if err != nil || !matched {
		return fmt.Errorf("недопустимый SQL индификатор: %s", name)
	}
	return nil
}

// Добавление столбца
func AddColumn(ctx context.Context, pool *pgxpool.Pool, table, col, typ, constraints string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(col); err != nil {
		return err
	}
	if typ == "" {
		return fmt.Errorf("типо столбца не может быть пустым")
	}
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s %s", table, col, typ, constraints)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Добавление столбца: %v", err)
		return fmt.Errorf("Не удалось добавить столбец: %v", err)
	}
	return nil
}

func DropColumn(ctx context.Context, pool *pgxpool.Pool, table, col string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(col); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, col)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Удаление столбца: %v", err)
		return fmt.Errorf("Не удалось удалить столбец: %v", err)
	}
	return nil
}

func AlterColumnType(ctx context.Context, pool *pgxpool.Pool, table, col, newtyp string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(col); err != nil {
		return err
	}
	if newtyp == "" {
		return fmt.Errorf("Новый тип столбца не может быть пустым")
	}
	query := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", table, col, newtyp)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Изменение типа столбца: %v", err)
		return fmt.Errorf("Не удалось изменить тип столбца: %v", err)
	}
	return nil
}

func RenameColumn(ctx context.Context, pool *pgxpool.Pool, table, oldCol, newCol string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(oldCol); err != nil {
		return err
	}
	if err := validateSQLIdent(newCol); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, oldCol, newCol)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Переименование столбца: %v", err)
		return fmt.Errorf("Не удалось переименовать столбец: %v", err)
	}
	return nil
}

// Операции над таблицами
func RenameTable(ctx context.Context, pool *pgxpool.Pool, oldTbl, newTbl string) error {
	if err := validateSQLIdent(oldTbl); err != nil {
		return err
	}
	if err := validateSQLIdent(newTbl); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s RENAME TO %s", oldTbl, newTbl)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Переименование таблицы: %v", err)
		return fmt.Errorf("Не удалось переименовать таблицу: %v", err)
	}
	return nil
}

// Добавление проверки
func AddCheck(ctx context.Context, pool *pgxpool.Pool, table, constraintName, expression string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(constraintName); err != nil {
		return err
	}
	if strings.TrimSpace(expression) == "" {
		return fmt.Errorf(" Условия проверки не может быть пустым")
	}
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s CHECK (%s)", table, constraintName, expression)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Добавление проверки: %v", err)
		return fmt.Errorf("Не удалось добавить проверку: %v", err)
	}
	return nil
}

func DropConstraint(ctx context.Context, pool *pgxpool.Pool, table, constraintName string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(constraintName); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", table, constraintName)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Удаление проверки: %v", err)
		return fmt.Errorf("Не удалось удалить проверку: %v", err)
	}
	return nil
}

func SetNotNull(ctx context.Context, pool *pgxpool.Pool, table, col string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(col); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", table, col)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Установка NOT NULL: %v", err)
		return fmt.Errorf("Не удалось установить NOT NULL: %v", err)
	}
	return nil
}

func DropNotNull(ctx context.Context, pool *pgxpool.Pool, table, col string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(col); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", table, col)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Удаление NOT NULL: %v", err)
		return fmt.Errorf("Не удалось удалить NOT NULL: %v", err)
	}
	return nil
}

func AddUnique(ctx context.Context, pool *pgxpool.Pool, table, constraintName, col string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(constraintName); err != nil {
		return err
	}
	if err := validateSQLIdent(col); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)", table, constraintName, col)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Добавление UNIQUE: %v", err)
		return fmt.Errorf("Не удалось добавить UNIQUE: %v", err)
	}
	return nil
}

func AddForeignKey(ctx context.Context, pool *pgxpool.Pool, table, constraintName, col, refTable, refCol string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(constraintName); err != nil {
		return err
	}
	if err := validateSQLIdent(col); err != nil {
		return err
	}
	if err := validateSQLIdent(refTable); err != nil {
		return err
	}
	if err := validateSQLIdent(refCol); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)", table, constraintName, col, refTable, refCol)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Добавление FOREIGN KEY: %v", err)
		return fmt.Errorf("Не удалось добавить FOREIGN KEY: %v", err)
	}
	return nil
}

func DropForeignKey(ctx context.Context, pool *pgxpool.Pool, table, constraintName string) error {
	if err := validateSQLIdent(table); err != nil {
		return err
	}
	if err := validateSQLIdent(constraintName); err != nil {
		return err
	}
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", table, constraintName)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Удаление FOREIGN KEY: %v", err)
		return fmt.Errorf("Не удалось удалить FOREIGN KEY: %v", err)
	}
	return nil
}

// Добавление только основных категорий
func insertInitialCategories(ctx context.Context, pool *pgxpool.Pool) error {
	// Проверяем, есть ли уже категории
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil || count > 0 {
		return nil // Категории уже есть
	}

	// Добавляем только категории (без продуктов)
	categories := [][]string{
		{"Электроника", "Электронные устройства и компоненты"},
		{"Книги", "Печатные и электронные книги"},
		{"Одежда", "Мужская и женская одежда"},
		{"Продукты", "Продукты питания и напитки"},
		{"Другое", "Прочие товары"},
	}

	for _, cat := range categories {
		_, err = pool.Exec(ctx,
			"INSERT INTO categories (name, description) VALUES ($1, $2)",
			cat[0], cat[1])
		if err != nil {
			return err
		}
	}

	return nil
}

// Получение всех продуктов для отображения в таблице
func GetAllProducts(ctx context.Context, pool *pgxpool.Pool) ([][]string, error) {
	query := `
    SELECT 
        p.id, 
        p.name, 
        p.description, 
        p.price, 
        p.quantity, 
        p.is_active,
        COALESCE(c.name, 'Без категории') as category_name
    FROM products p
    LEFT JOIN categories c ON p.category_id = c.id
    ORDER BY p.id`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	// Заголовки таблицы
	result := [][]string{
		{"ID", "Название", "Описание", "Цена", "Количество", "Активен", "Категория"},
	}

	for rows.Next() {
		var id int
		var name, description, category string
		var price float64
		var quantity int
		var isActive bool

		err := rows.Scan(&id, &name, &description, &price, &quantity, &isActive, &category)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}

		result = append(result, []string{
			strconv.Itoa(id),
			name,
			description,
			fmt.Sprintf("%.2f", price),
			strconv.Itoa(quantity),
			fmt.Sprintf("%t", isActive),
			category,
		})
	}

	return result, nil
}

// Добавление нового продукта
func InsertProduct(ctx context.Context, pool *pgxpool.Pool, name, description string, price float64, quantity int, categoryID *int) error {
	query := `
    INSERT INTO products (name, description, price, quantity, category_id) 
    VALUES ($1, $2, $3, $4, $5)`

	_, err := pool.Exec(ctx, query, name, description, price, quantity, categoryID)
	if err != nil {
		return fmt.Errorf("ошибка добавления продукта: %w", err)
	}

	return nil
}

// WARNING
// Вот нужно эту функцию подправить в теории
// Или иной способ!
func UpdateProduct(ctx context.Context, pool *pgxpool.Pool, id int, name, description string, price float64, quantity int, categoryID *int) error {
	query := `
	UPDATE products
	SET name = $2, description = $3, price = $4, quantity = $5, category_id = $6, updated_at = NOW()
	WHERE id = $1`

	commandTag, err := pool.Exec(ctx, query, id, name, description, price, quantity, categoryID)
	if err != nil {
		return fmt.Errorf("ошибка обновления продукта: %w", err)
	}

	// Проверяем, была ли обновлена хотя бы одна строка
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("продукт с ID %d не найден", id)
	}

	fmt.Printf("Обновлен продукт ID: %d, затронуто строк: %d\n", id, commandTag.RowsAffected())
	return nil
}

// Получение категорий для выпадающего списка
func GetCategories(ctx context.Context, pool *pgxpool.Pool) ([][]string, error) {
	query := "SELECT id, name FROM categories ORDER BY name"

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result [][]string
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result = append(result, []string{strconv.Itoa(id), name})
	}

	return result, nil
}

// Удаление продукта
func DeleteProduct(ctx context.Context, pool *pgxpool.Pool, id int) error {
	query := "DELETE FROM products WHERE id = $1"
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления продукта: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("продукт с ID %d не найден", id)
	}

	return nil
}

// Тестирование подключения к БД
func TestConnection(ctx context.Context, pool *pgxpool.Pool) error {
	var version string
	err := pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	fmt.Printf("Подключение к PostgreSQL успешно: %s\n", version[:50]+"...")
	return nil
}
