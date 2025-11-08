package table

import (
	operation "BD_Mirea/internal"
	"context"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ========== ТРЕБОВАНИЕ 3: ПОИСК ПО ТЕКСТУ (LIKE И REGEX) ==========

// UISearchDialog создаёт диалог для поиска по текстовым полям
func UISearchDialog(ctx context.Context, pool *pgxpool.Pool, window fyne.Window, tableName string) {
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Столбец для поиска")

	patternEntry := widget.NewEntry()
	patternEntry.SetPlaceHolder("Паттерн поиска")

	searchTypeSelect := widget.NewSelect(
		[]string{"LIKE", "REGEX (~)", "REGEX NoCase (~*)", "NOT REGEX (!~)", "NOT REGEX NoCase (!~*)"},
		func(s string) {},
	)
	searchTypeSelect.SetSelected("LIKE")

	resultsLabel := widget.NewLabel("Результаты: 0 записей")
	var resultsData [][]string

	resultsTable := widget.NewTable(
		func() (int, int) {
			if len(resultsData) == 0 {
				return 0, 0
			}
			return len(resultsData), len(resultsData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			if i.Row < len(resultsData) && i.Col < len(resultsData[i.Row]) {
				o.(*widget.Label).SetText(resultsData[i.Row][i.Col])
				if i.Row == 0 {
					o.(*widget.Label).TextStyle = fyne.TextStyle{Bold: true}
				}
			}
		},
	)

	// Кнопка выполнения поиска
	searchButton := widget.NewButton("Выполнить поиск", func() {
		column := strings.TrimSpace(columnEntry.Text)
		pattern := strings.TrimSpace(patternEntry.Text)
		searchType := searchTypeSelect.Selected

		if column == "" || pattern == "" {
			showError(window, "Укажите столбец и паттерн поиска")
			return
		}

		qb := operation.NewQueryBuilder(tableName)

		// Применяем выбранный тип поиска
		switch searchType {
		case "LIKE":
			qb.WhereLike(column, "%"+pattern+"%")
		case "REGEX (~)":
			qb.WhereRegex(column, pattern)
		case "REGEX NoCase (~*)":
			qb.WhereRegexNoCase(column, pattern)
		case "NOT REGEX (!~)":
			qb.WhereNotRegex(column, pattern)
		case "NOT REGEX NoCase (!~*)":
			qb.WhereNotRegexNoCase(column, pattern)
		}

		results, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, "Ошибка поиска: "+err.Error())
			return
		}

		if len(results) == 0 {
			resultsLabel.SetText("Результаты: 0 записей")
			resultsData = [][]string{}
		} else {
			resultsData = results
			resultsLabel.SetText(fmt.Sprintf("Результаты: %d записей", len(results)-1))
		}

		resultsTable.Refresh()
	})

	// Очистка формы
	clearButton := widget.NewButton("Очистить", func() {
		columnEntry.SetText("")
		patternEntry.SetText("")
		resultsData = [][]string{}
		resultsLabel.SetText("Результаты: 0 записей")
		resultsTable.Refresh()
	})

	form := container.NewVBox(
		widget.NewCard("Поиск по тексту", "",
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("Столбец", columnEntry),
					widget.NewFormItem("Паттерн", patternEntry),
					widget.NewFormItem("Тип поиска", searchTypeSelect),
				),
				container.NewGridWithColumns(2,
					searchButton,
					clearButton,
				),
			)),
		resultsLabel,
		widget.NewCard("Результаты", "", container.NewScroll(resultsTable)),
	)

	// Показываем в новом окне
	searchWindow := fyne.CurrentApp().NewWindow("Поиск по тексту")
	searchWindow.SetContent(container.NewScroll(form))
	searchWindow.Resize(fyne.NewSize(900, 600))
	searchWindow.CenterOnScreen()
	searchWindow.Show()
}

// ========== ТРЕБОВАНИЕ 4: ФУНКЦИИ РАБОТЫ СО СТРОКАМИ ==========

// UIStringFunctions создаёт диалог с функциями преобразования строк
func UIStringFunctions(ctx context.Context, pool *pgxpool.Pool, window fyne.Window, tableName string) {
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Столбец для преобразования")

	functionSelect := widget.NewSelect([]string{
		"UPPER - верхний регистр",
		"LOWER - нижний регистр",
		"TRIM - удалить пробелы",
		"LTRIM - удалить пробелы слева",
		"RTRIM - удалить пробелы справа",
		"SUBSTRING - извлечь подстроку",
		"LPAD - дополнить слева",
		"RPAD - дополнить справа",
		"CONCAT - объединить столбцы",
	}, func(s string) {})
	functionSelect.SetSelected("UPPER - верхний регистр")

	// Дополнительные параметры
	startEntry := widget.NewEntry()
	startEntry.SetPlaceHolder("Начало (для SUBSTRING)")
	startEntry.Hide()

	lengthEntry := widget.NewEntry()
	lengthEntry.SetPlaceHolder("Длина (для SUBSTRING)")
	lengthEntry.Hide()

	padCharEntry := widget.NewEntry()
	padCharEntry.SetPlaceHolder("Символ заполнения (для LPAD/RPAD)")
	padCharEntry.Hide()

	padLengthEntry := widget.NewEntry()
	padLengthEntry.SetPlaceHolder("Итоговая длина (для LPAD/RPAD)")
	padLengthEntry.Hide()

	column2Entry := widget.NewEntry()
	column2Entry.SetPlaceHolder("Второй столбец (для CONCAT)")
	column2Entry.Hide()

	// Переключение видимости параметров
	functionSelect.OnChanged = func(s string) {
		startEntry.Hide()
		lengthEntry.Hide()
		padCharEntry.Hide()
		padLengthEntry.Hide()
		column2Entry.Hide()

		if strings.Contains(s, "SUBSTRING") {
			startEntry.Show()
			lengthEntry.Show()
		} else if strings.Contains(s, "LPAD") || strings.Contains(s, "RPAD") {
			padLengthEntry.Show()
			padCharEntry.Show()
		} else if strings.Contains(s, "CONCAT") {
			column2Entry.Show()
		}
	}

	// Кнопка применения функции
	applyButton := widget.NewButton("Применить и показать результаты", func() {
		column := strings.TrimSpace(columnEntry.Text)
		if column == "" {
			showError(window, "Укажите столбец")
			return
		}

		functionType := functionSelect.Selected
		qb := operation.NewQueryBuilder(tableName)
		qb.Select(column)

		// Применяем выбранную функцию
		if strings.Contains(functionType, "UPPER") {
			qb.SelectUpper(column)
		} else if strings.Contains(functionType, "LOWER") {
			qb.SelectLower(column)
		} else if strings.Contains(functionType, "TRIM") && !strings.Contains(functionType, "LTRIM") && !strings.Contains(functionType, "RTRIM") {
			qb.SelectTrim(column)
		} else if strings.Contains(functionType, "LTRIM") {
			qb.SelectLTrim(column)
		} else if strings.Contains(functionType, "RTRIM") {
			qb.SelectRTrim(column)
		} else if strings.Contains(functionType, "SUBSTRING") {
			start, _ := strconv.Atoi(startEntry.Text)
			length, _ := strconv.Atoi(lengthEntry.Text)
			if start <= 0 || length <= 0 {
				showError(window, "Укажите корректные начало и длину")
				return
			}
			qb.SelectSubstring(column, start, length)
		} else if strings.Contains(functionType, "LPAD") {
			padChar := strings.TrimSpace(padCharEntry.Text)
			padLen, _ := strconv.Atoi(padLengthEntry.Text)
			if padChar == "" || padLen <= 0 {
				showError(window, "Укажите символ и длину дополнения")
				return
			}
			qb.SelectLPad(column, padLen, padChar)
		} else if strings.Contains(functionType, "RPAD") {
			padChar := strings.TrimSpace(padCharEntry.Text)
			padLen, _ := strconv.Atoi(padLengthEntry.Text)
			if padChar == "" || padLen <= 0 {
				showError(window, "Укажите символ и длину дополнения")
				return
			}
			qb.SelectRPad(column, padLen, padChar)
		} else if strings.Contains(functionType, "CONCAT") {
			column2 := strings.TrimSpace(column2Entry.Text)
			if column2 == "" {
				showError(window, "Укажите второй столбец")
				return
			}
			qb.SelectConcat(column, column2)
		}

		qb.Limit(5) // Показываем только 5 примеров
		results, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, "Ошибка: "+err.Error())
			return
		}

		if len(results) > 1 {
			showInfo(window, fmt.Sprintf("Функция применена. Показано %d результатов.", len(results)-1))
		}
	})

	infoText := widget.NewRichTextFromMarkdown(`
**Доступные функции для преобразования строк:**

- **UPPER** - Преобразует текст в ВЕРХНИЙ РЕГИСТР
  Пример: UPPER('hello') → 'HELLO'

- **LOWER** - Преобразует текст в нижний регистр
  Пример: LOWER('HELLO') → 'hello'

- **TRIM** - Удаляет пробелы с обеих сторон
  Пример: TRIM('  hello  ') → 'hello'

- **LTRIM** - Удаляет пробелы слева
  Пример: LTRIM('  hello') → 'hello'

- **RTRIM** - Удаляет пробелы справа
  Пример: RTRIM('hello  ') → 'hello'

- **SUBSTRING** - Извлекает часть строки
  Пример: SUBSTRING('hello', 2, 3) → 'ell'

- **LPAD** - Дополняет строку слева
  Пример: LPAD('123', 5, '0') → '00123'

- **RPAD** - Дополняет строку справа
  Пример: RPAD('123', 5, '0') → '12300'

- **CONCAT** - Объединяет две строки
  Пример: CONCAT('hello', ' ', 'world') → 'hello world'
`)

	form := container.NewVBox(
		widget.NewCard("Функции преобразования строк", "",
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("Столбец", columnEntry),
					widget.NewFormItem("Функция", functionSelect),
					widget.NewFormItem("", startEntry),
					widget.NewFormItem("", lengthEntry),
					widget.NewFormItem("", padCharEntry),
					widget.NewFormItem("", padLengthEntry),
					widget.NewFormItem("", column2Entry),
				),
				applyButton,
			)),
		infoText,
	)

	stringWindow := fyne.CurrentApp().NewWindow("Функции работы со строками")
	stringWindow.SetContent(container.NewScroll(form))
	stringWindow.Resize(fyne.NewSize(800, 700))
	stringWindow.CenterOnScreen()
	stringWindow.Show()
}

// ========== ТРЕБОВАНИЕ 5: МАСТЕР СОЕДИНЕНИЙ (JOIN) ==========

// UIJoinWizard создаёт многошаговый мастер для создания JOIN запросов
func UIJoinWizard(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	var currentStep int = 1
	var table1, table2 string
	var field1, field2, joinType string

	wizardWindow := fyne.CurrentApp().NewWindow("Мастер соединений (JOIN)")
	wizardWindow.Resize(fyne.NewSize(600, 500))
	wizardWindow.CenterOnScreen()

	content := container.NewVBox()

	prevButton := widget.NewButton("← Назад", nil)
	nextButton := widget.NewButton("Далее →", nil)

	prevButton.OnTapped = func() {
		if currentStep > 1 {
			currentStep--
			showJoinStep(ctx, pool, content, &currentStep, &table1, &table2, &field1, &field2, &joinType, prevButton, nextButton, window)
		}
	}
	prevButton.Disable()

	nextButton.OnTapped = func() {
		if currentStep < 3 {
			currentStep++
			showJoinStep(ctx, pool, content, &currentStep, &table1, &table2, &field1, &field2, &joinType, prevButton, nextButton, window)
		} else if currentStep == 3 {
			executeJoinQuery(ctx, pool, window, table1, table2, field1, field2, joinType)
			wizardWindow.Close()
		}
	}

	showJoinStep(ctx, pool, content, &currentStep, &table1, &table2, &field1, &field2, &joinType, prevButton, nextButton, window)

	final := container.NewVBox(
		content,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, prevButton, nextButton),
	)

	wizardWindow.SetContent(final)
	wizardWindow.Show()
}

// showJoinStep показывает нужный шаг мастера JOIN
func showJoinStep(ctx context.Context, pool *pgxpool.Pool, content *fyne.Container,
	step *int, table1, table2, field1, field2, joinType *string,
	prevButton, nextButton *widget.Button, window fyne.Window) {

	content.Objects = []fyne.CanvasObject{}

	switch *step {
	case 1:
		prevButton.Disable()
		nextButton.Enable()
		nextButton.SetText("Далее →")

		table1Select := widget.NewSelect([]string{}, func(s string) {
			*table1 = s
		})
		table1Select.PlaceHolder = "Выберите первую таблицу..."

		table2Select := widget.NewSelect([]string{}, func(s string) {
			*table2 = s
		})
		table2Select.PlaceHolder = "Выберите вторую таблицу..."

		// Загружаем таблицы из БД
		go func() {
			tables, err := getAllTables(ctx, pool)
			if err == nil {
				table1Select.Options = tables
				table2Select.Options = tables
			}
		}()

		content.Add(widget.NewCard("Шаг 1: Выбор таблиц", "",
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("Первая таблица", table1Select),
					widget.NewFormItem("Вторая таблица", table2Select),
				),
			)))

	case 2:
		prevButton.Enable()
		nextButton.Enable()
		nextButton.SetText("Далее →")

		field1Select := widget.NewSelect([]string{}, func(s string) {
			*field1 = s
		})
		field1Select.PlaceHolder = "Выберите поле первой таблицы..."

		opSelect := widget.NewSelect([]string{"=", ">", "<", ">=", "<=", "LIKE", "IN"}, nil)
		opSelect.SetSelected("=")

		field2Select := widget.NewSelect([]string{}, func(s string) {
			*field2 = s
		})
		field2Select.PlaceHolder = "Выберите поле второй таблицы..."

		// Загружаем поля
		go func() {
			if *table1 != "" {
				fields1, err := getTableColumns(ctx, pool, *table1)
				if err == nil {
					field1Select.Options = fields1
				}
			}
			if *table2 != "" {
				fields2, err := getTableColumns(ctx, pool, *table2)
				if err == nil {
					field2Select.Options = fields2
				}
			}
		}()

		content.Add(widget.NewCard("Шаг 2: Условие соединения", "",
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("Поле первой таблицы", field1Select),
					widget.NewFormItem("Оператор", opSelect),
					widget.NewFormItem("Поле второй таблицы", field2Select),
				),
			)))

	case 3:
		prevButton.Enable()
		nextButton.Enable()
		nextButton.SetText("Выполнить JOIN")

		joinSelect := widget.NewSelect(
			[]string{"INNER", "LEFT", "RIGHT", "FULL"},
			func(s string) { *joinType = s },
		)
		joinSelect.SetSelected("INNER")

		desc := widget.NewRichTextFromMarkdown(`
**Типы соединений:**

- **INNER JOIN** - Возвращает только совпадающие записи из обеих таблиц
- **LEFT JOIN** - Все записи из левой таблицы + совпадающие из правой
- **RIGHT JOIN** - Все записи из правой таблицы + совпадающие из левой
- **FULL JOIN** - Все записи из обеих таблиц
`)

		content.Add(widget.NewCard("Шаг 3: Тип соединения", "",
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("Тип JOIN", joinSelect),
				),
				desc,
			)))
	}
}

// executeJoinQuery выполняет построенный JOIN запрос
func executeJoinQuery(ctx context.Context, pool *pgxpool.Pool, window fyne.Window,
	table1, table2, field1, field2, joinType string) {

	if table1 == "" || table2 == "" || field1 == "" || field2 == "" || joinType == "" {
		showError(window, "Заполните все параметры соединения")
		return
	}

	qb := operation.NewQueryBuilder(table1)
	onCondition := fmt.Sprintf("%s.%s = %s.%s", table1, field1, table2, field2)

	switch joinType {
	case "INNER":
		qb.InnerJoin(table2, onCondition)
	case "LEFT":
		qb.LeftJoin(table2, onCondition)
	case "RIGHT":
		qb.RightJoin(table2, onCondition)
	case "FULL":
		qb.FullJoin(table2, onCondition)
	}

	results, err := qb.Execute(ctx, pool)
	if err != nil {
		showError(window, "Ошибка JOIN: "+err.Error())
		return
	}

	if len(results) == 0 {
		showError(window, "Результатов не найдено")
		return
	}

	// Показываем результаты
	resultTable, err := CreateTable(results)
	if err != nil {
		showError(window, err.Error())
		return
	}

	resultWindow := fyne.CurrentApp().NewWindow(fmt.Sprintf("%s JOIN %s", joinType, table2))
	content := container.NewVBox(
		widget.NewCard("SQL Запрос", "", widget.NewLabel(qb.Build())),
		widget.NewCard("Результаты", "", container.NewScroll(resultTable)),
	)
	resultWindow.SetContent(content)
	resultWindow.Resize(fyne.NewSize(1000, 600))
	resultWindow.CenterOnScreen()
	resultWindow.Show()
}

// Вспомогательные функции
func getAllTables(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	var tables []string
	rows, err := pool.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}
	return tables, nil
}

func getTableColumns(ctx context.Context, pool *pgxpool.Pool, tableName string) ([]string, error) {
	var columns []string
	query := fmt.Sprintf(
		"SELECT column_name FROM information_schema.columns WHERE table_name = '%s'",
		tableName,
	)
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, err
		}
		columns = append(columns, colName)
	}
	return columns, nil
}
