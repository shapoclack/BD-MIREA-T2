package main

import (
	"BD_Mirea/internal"
	table "BD_Mirea/ui"
	"context"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Подключение к PostgreSQL
	ctx := context.Background()
	connStr := "postgres://postgres:19frol67@localhost:1703/postgres?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer pool.Close()

	// Тестирование подключения
	if err := internal.TestConnection(ctx, pool); err != nil {
		log.Fatalf("Тест подключения провален: %v", err)
	}

	// Создание таблиц
	if err := internal.CreateTables(ctx, pool); err != nil {
		log.Fatalf("Ошибка создания таблиц: %v", err)
	}

	// Создание GUI приложения
	mainApp := app.NewWithID("PostgreSQL-UI-Client")

	// Применяем кастомную тему с улучшенными цветами
	mainApp.Settings().SetTheme(&table.CustomTheme{})

	window := mainApp.NewWindow("PostgreSQL UI Client - Управление таблицами")
	window.Resize(fyne.NewSize(1400, 800))
	window.CenterOnScreen()

	// Создание UI компонентов
	table.CreateAdvancedUI(window, ctx, pool)

	// Показываем окно
	window.ShowAndRun()
}
