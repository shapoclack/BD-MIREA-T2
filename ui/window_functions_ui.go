package table

import (
	"context"
	"fmt"
	"log"
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
	log.Println("[WINDOW_FUNCTIONS] Инициализация UI для оконных функций")

	// Таблица
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("products, categories, orders...")

	// Тип оконной функции
	funcTypeSelect := widget.NewSelect(
		[]string{"RANK", "LAG", "LEAD"},
		nil,
	)
	funcTypeSelect.PlaceHolder = "Выберите функцию"

	// PARTITION BY
	partitionByEntry := widget.NewEntry()
	partitionByEntry.SetPlaceHolder("category_id, region (или оставьте пусто)")
	partitionByEntry.MultiLine = true
	partitionByEntry.SetMinRowsVisible(2)

	// ORDER BY
	orderByEntry := widget.NewEntry()
	orderByEntry.SetPlaceHolder("price DESC, created_at ASC (или оставьте пусто)")
	orderByEntry.MultiLine = true
	orderByEntry.SetMinRowsVisible(3)

	// Для LAG/LEAD
	targetColumnLabel := widget.NewLabel("Целевая колонка (для LAG/LEAD):")
	targetColumnEntry := widget.NewEntry()
	targetColumnEntry.SetPlaceHolder("price, quantity, created_at...")

	offsetLabel := widget.NewLabel("Смещение (для LAG/LEAD):")
	offsetEntry := widget.NewEntry()
	offsetEntry.SetPlaceHolder("1")
	offsetEntry.SetText("1")

	// Алиас
	aliasLabel := widget.NewLabel("Алиас результата (опционально):")
	aliasEntry := widget.NewEntry()
	aliasEntry.SetPlaceHolder("rank, prev_price, next_price...")

	// Информация
	infoLabel := widget.NewRichTextFromMarkdown(`## Оконные функции (Window Functions)

**RANK()** – ранг строки по ORDER BY.  
**LAG(column, offset)** – значение из предыдущей строки.  
**LEAD(column, offset)** – значение из следующей строки.

1. Таблица – имя таблицы.  
2. Функция – RANK / LAG / LEAD.  
3. PARTITION BY – разбиение (опционально).  
4. ORDER BY – порядок (обязательно).  
5. Для LAG/LEAD – целевая колонка и смещение.  
6. Алиас – имя результирующей колонки.
`)

	// Данные и таблица результатов
	var resultsData [][]string
	resultsTable := widget.NewTable(
		func() (int, int) {
			if len(resultsData) == 0 {
				return 0, 0
			}
			return len(resultsData), len(resultsData[0])
		},
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Wrapping = fyne.TextWrapWord
			lbl.Truncation = fyne.TextTruncateEllipsis
			return lbl
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			if id.Row < len(resultsData) && id.Col < len(resultsData[id.Row]) {
				text := resultsData[id.Row][id.Col]
				if len(text) > 80 {
					text = text[:80] + "..."
				}
				obj.(*widget.Label).SetText(text)
			}
		},
	)
	// ширина колонок
	for col := 0; col < 15; col++ {
		switch col {
		case 0: // id
			resultsTable.SetColumnWidth(col, 60)
		case 1, 2: // name, description
			resultsTable.SetColumnWidth(col, 220)
		default:
			resultsTable.SetColumnWidth(col, 160)
		}
	}

	// Обработчик кнопки
	executeButton := widget.NewButton("Выполнить запрос", func() {
		log.Println("\n" + strings.Repeat("=", 80))
		log.Println("[EXECUTE] Начало выполнения запроса")
		log.Println(strings.Repeat("=", 80))

		tableName := strings.TrimSpace(tableEntry.Text)
		functionType := funcTypeSelect.Selected
		partitionByStr := strings.TrimSpace(partitionByEntry.Text)
		orderByStr := strings.TrimSpace(orderByEntry.Text)
		targetColumn := strings.TrimSpace(targetColumnEntry.Text)
		offsetStr := strings.TrimSpace(offsetEntry.Text)
		alias := strings.TrimSpace(aliasEntry.Text)

		// Валидация
		if tableName == "" {
			showError(window, "Ошибка: укажите таблицу")
			return
		}
		if functionType == "" {
			showError(window, "Ошибка: выберите тип функции")
			return
		}
		if orderByStr == "" {
			showError(window, "Ошибка: ORDER BY обязателен")
			return
		}
		if functionType != "RANK" && targetColumn == "" {
			showError(window, "Ошибка: для LAG/LEAD укажите целевую колонку")
			return
		}

		// PARTITION BY
		var partitionByCols []string
		if partitionByStr != "" {
			parts := strings.Split(partitionByStr, ",")
			for _, p := range parts {
				if col := strings.TrimSpace(p); col != "" {
					partitionByCols = append(partitionByCols, col)
				}
			}
		}

		// ORDER BY
		orderByCols := parseOrderByClause(orderByStr)
		if len(orderByCols) == 0 {
			showError(window, "Ошибка: неправильный формат ORDER BY (column ASC|DESC)")
			return
		}

		// offset
		offset := 1
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o > 0 {
				offset = o
			}
		}

		// QueryBuilder
		qb := internal.NewQueryBuilder(tableName)

		// Базовые столбцы
		switch tableName {
		case "categories":
			qb.Select("id", "name", "description", "created_at")
		default:
			qb.Select("*")
		}

		// Оконная функция
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

		sqlQuery := qb.Build()
		log.Println("[SQL] " + sqlQuery)

		// Выполнение
		res, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, fmt.Sprintf("Ошибка при выполнении:\n%v", err))
			return
		}

		resultsData = res
		resultsTable.Refresh()

		rows := len(res) - 1
		if rows < 0 {
			rows = 0
		}
		showInfo(window, fmt.Sprintf("✓ Успех! Найдено %d строк", rows))
	})

	// Показ/скрытие полей LAG/LEAD
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

	// Левая панель
	leftPanel := container.NewScroll(
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
					widget.NewRichTextFromMarkdown("Формат: column1 ASC, column2 DESC"),
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
			widget.NewCard("Алиас результата", "",
				container.NewVBox(
					aliasLabel,
					aliasEntry,
				),
			),
			executeButton,
		),
	)

	// Правая панель
	resultsPanel := container.NewBorder(
		widget.NewLabel("Результаты"),
		nil, nil, nil,
		container.NewScroll(resultsTable),
	)

	// Основной контейнер
	mainContent := container.NewHSplit(
		leftPanel,
		container.NewMax(resultsPanel),
	)
	mainContent.SetOffset(0.45)

	// Окно
	windowTitle := "Оконные функции (Window Functions)"
	windowFuncUI := fyne.CurrentApp().NewWindow(windowTitle)
	windowFuncUI.SetTitle(windowTitle)
	windowFuncUI.SetContent(mainContent)
	windowFuncUI.Resize(fyne.NewSize(1400, 950))
	windowFuncUI.CenterOnScreen()
	windowFuncUI.Show()
}

// parseOrderByClause парсит ORDER BY
func parseOrderByClause(orderByStr string) []internal.OrderByItem {
	var items []internal.OrderByItem
	parts := strings.Split(orderByStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tokens := strings.Fields(part)
		if len(tokens) == 0 {
			continue
		}
		column := tokens[0]
		direction := "ASC"
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
