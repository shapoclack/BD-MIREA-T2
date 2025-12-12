package table

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"BD_Mirea/internal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UIWindowFunctions модуль для работы с оконными функциями (RANK, LAG, LEAD)
func UIWindowFunctions(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	// Таблица
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("products, categories, orders...")

	// Выбор типа оконной функции
	funcTypeSelect := widget.NewSelect(
		[]string{"RANK", "LAG", "LEAD"},
		nil,
	)
	funcTypeSelect.PlaceHolder = "Выберите функцию"

	// PARTITION BY (как строка через запятую)
	partitionByEntry := widget.NewEntry()
	partitionByEntry.SetPlaceHolder("category, region (или оставьте пусто)")
	partitionByEntry.MultiLine = true
	partitionByEntry.SetMinRowsVisible(2)

	// ORDER BY - с направлением
	orderByEntry := widget.NewEntry()
	orderByEntry.SetPlaceHolder("price DESC, date ASC (или оставьте пусто)")
	orderByEntry.MultiLine = true
	orderByEntry.SetMinRowsVisible(3)

	// Для LAG/LEAD
	targetColumnLabel := widget.NewLabel("Целевая колонка (для LAG/LEAD):")
	targetColumnEntry := widget.NewEntry()
	targetColumnEntry.SetPlaceHolder("price, quantity, date...")

	offsetLabel := widget.NewLabel("Смещение (для LAG/LEAD):")
	offsetEntry := widget.NewEntry()
	offsetEntry.SetPlaceHolder("1")
	offsetEntry.SetText("1")

	// Alias для результирующей колонки
	aliasLabel := widget.NewLabel("Алиас результата (опционально):")
	aliasEntry := widget.NewEntry()
	aliasEntry.SetPlaceHolder("rank, prev_price, next_price...")

	// Инфо текст
	infoLabel := widget.NewRichTextFromMarkdown(`
## Оконные функции (Window Functions)

**RANK()** - ранжирует строки по ORDER BY, присваивая номер каждой строке. Одинаковые значения получают одинаковый ранг.

**LAG(column, offset)** - возвращает значение из предыдущей строки (с смещением).

**LEAD(column, offset)** - возвращает значение из следующей строки (с смещением).

### Примеры:

- **RANK()** - просто ранжирование, без PARTITION BY или с PARTITION BY

- **LAG(price, 1)** - предыдущая цена в очереди

- **LEAD(date, 2)** - дата через 2 строки вперед

### Заполнение:

1. **Таблица** - имя таблицы (обязательно)

2. **Функция** - выберите RANK, LAG или LEAD

3. **PARTITION BY** - разделение данных (опционально, через запятую)

4. **ORDER BY** - порядок обработки (ОБЯЗАТЕЛЬНО! Формат: "колонка ASC/DESC")

5. Для LAG/LEAD: укажите целевую колонку и смещение

6. **Алиас** - имя результирующей колонки в выводе
`)

	// Таблица для результатов
	var resultsTable *widget.Table
	var resultsData [][]string

	// Кнопка выполнения
	executeButton := widget.NewButton("Выполнить запрос", func() {
		tableName := strings.TrimSpace(tableEntry.Text)
		functionType := funcTypeSelect.Selected
		partitionByStr := strings.TrimSpace(partitionByEntry.Text)
		orderByStr := strings.TrimSpace(orderByEntry.Text)
		targetColumn := strings.TrimSpace(targetColumnEntry.Text)
		offsetStr := strings.TrimSpace(offsetEntry.Text)
		alias := strings.TrimSpace(aliasEntry.Text)

		// Валидация
		if tableName == "" {
			showError(window, "Ошибка")
			return
		}

		if functionType == "" {
			showError(window, "Ошибка")
			return
		}

		if orderByStr == "" {
			showError(window, "Ошибка")
			return
		}

		if functionType != "RANK" && targetColumn == "" {
			showError(window, "Ошибка")
			return
		}

		// Парсим PARTITION BY
		var partitionByCols []string
		if partitionByStr != "" {
			parts := strings.Split(partitionByStr, ",")
			for _, p := range parts {
				if col := strings.TrimSpace(p); col != "" {
					partitionByCols = append(partitionByCols, col)
				}
			}
		}

		// Парсим ORDER BY
		orderByCols := parseOrderByClause(orderByStr)
		if len(orderByCols) == 0 {
			showError(window, "Ошибка")
			return
		}

		// Парсим offset
		offset := 1
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o > 0 {
				offset = o
			}
		}

		// Строим QueryBuilder
		qb := internal.NewQueryBuilder(tableName)
		qb.Select("*") // Выбираем все базовые колонки

		// Применяем оконную функцию
		switch functionType {
		case "RANK":
			if alias == "" {
				alias = "rank"
			}
			qb.SelectRank(alias, partitionByCols, orderByCols)
		case "LAG":
			if alias == "" {
				alias = fmt.Sprintf("lag_%s", targetColumn)
			}
			qb.SelectLag(targetColumn, offset, alias, partitionByCols, orderByCols)
		case "LEAD":
			if alias == "" {
				alias = fmt.Sprintf("lead_%s", targetColumn)
			}
			qb.SelectLead(targetColumn, offset, alias, partitionByCols, orderByCols)
		}

		// Выполняем запрос
		results, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, fmt.Sprintf("%v", err))
			return
		}

		// Парсим результаты в таблицу
		resultsData = results
		resultsTable = createTableFromData(results)
		resultsTable.Refresh()
		showInfo(window, fmt.Sprintf("Успех. Найдено %d строк", len(results)-1))

	})

	// Таблица для вывода
	resultsTable = widget.NewTable(
		func() (int, int) {
			if len(resultsData) == 0 {
				return 0, 0
			}
			return len(resultsData), len(resultsData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			if id.Row < len(resultsData) && id.Col < len(resultsData[id.Row]) {
				obj.(*widget.Label).SetText(resultsData[id.Row][id.Col])
			}
		},
	)

	// Устанавливаем ширину колонок
	for col := 0; col < 10; col++ {
		resultsTable.SetColumnWidth(col, 120)
	}

	// Обновляем видимость элементов в зависимости от типа функции
	funcTypeSelect.OnChanged = func(s string) {
		isRank := s == "RANK"
		if isRank {
			targetColumnEntry.Hide()
			offsetEntry.Hide()
			targetColumnLabel.Hide()
			offsetLabel.Hide()
		} else {
			targetColumnEntry.Show()
			offsetEntry.Show()
			targetColumnLabel.Show()
			offsetLabel.Show()
		}
	}

	mainContent := container.NewHSplit(
		// Левая часть - форма
		container.NewScroll(
			container.NewVBox(
				infoLabel,
				widget.NewCard("Таблица и функция", "",
					container.NewVBox(
						widget.NewForm(
							widget.NewFormItem("Таблица:", tableEntry),
							widget.NewFormItem("Тип функции:", funcTypeSelect),
						),
					),
				),
				widget.NewCard("PARTITION BY", "(опционально)",
					container.NewVBox(
						widget.NewRichTextFromMarkdown("Разделить данные по столбцам (через запятую):"),
						partitionByEntry,
					),
				),
				widget.NewCard("ORDER BY", "(ОБЯЗАТЕЛЬНО!)",
					container.NewVBox(
						widget.NewRichTextFromMarkdown("Порядок обработки строк. Формат: `column1 ASC, column2 DESC`"),
						orderByEntry,
					),
				),
				widget.NewCard("Параметры LAG/LEAD", "",
					container.NewVBox(
						targetColumnLabel,
						targetColumnEntry,
						offsetLabel,
						offsetEntry,
					),
				),
				widget.NewCard("", "",
					container.NewVBox(
						aliasLabel,
						aliasEntry,
					),
				),
				executeButton,
			),
		),
		// Правая часть - результаты
		container.NewScroll(
			container.NewVBox(
				widget.NewCard("Результаты", "",
					resultsTable,
				),
			),
		),
	)

	// Устанавливаем соотношение 40/60
	mainContent.SetOffset(0.4)

	// Окно
	windowTitle := "Оконные функции (Window Functions)"
	windowFuncUI := fyne.CurrentApp().NewWindow(windowTitle)
	windowFuncUI.SetTitle(windowTitle)
	windowFuncUI.SetContent(mainContent)
	windowFuncUI.Resize(fyne.NewSize(1400, 950))
	windowFuncUI.CenterOnScreen()
	windowFuncUI.Show()

}

// parseOrderByClause парсит ORDER BY из строки типа "price DESC, date ASC"
func parseOrderByClause(orderByStr string) []internal.OrderByItem {
	var items []internal.OrderByItem
	parts := strings.Split(orderByStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Парсим "колонка ASC" или "колонка DESC" или просто "колонка"
		tokens := strings.Fields(part)
		if len(tokens) == 0 {
			continue
		}

		column := tokens[0]
		direction := "ASC" // по умолчанию
		if len(tokens) >= 2 {
			dir := strings.ToUpper(tokens[1])
			if dir == "ASC" || dir == "DESC" {
				direction = dir
			}
		}

		items = append(items, internal.OrderByItem{
			Column:    column,
			Direction: direction,
		})
	}

	return items
}

// createTableFromData создаёт таблицу из данных
func createTableFromData(data [][]string) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			if len(data) == 0 {
				return 0, 0
			}
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			if id.Row < len(data) && id.Col < len(data[id.Row]) {
				obj.(*widget.Label).SetText(data[id.Row][id.Col])
			}
		},
	)

	// Устанавливаем ширину колонок
	for col := 0; col < 15; col++ {
		table.SetColumnWidth(col, 100)
	}

	return table
}
