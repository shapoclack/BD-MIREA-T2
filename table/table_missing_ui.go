package table

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	operation "BD_Mirea/internal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ===== ТРЕБОВАНИЕ 1: UISearchDialog - ПОИСК С РЕГУЛЯРНЫМИ ВЫРАЖЕНИЯМИ И SIMILAR TO =====

// UISearchDialog создаёт интерфейс для поиска по строкам с поддержкой регулярных выражений и SIMILAR TO
func UISearchDialog(ctx context.Context, pool *pgxpool.Pool, window fyne.Window, data [][]string) {
	// Выбор таблицы
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Введите название таблицы (products, categories...)")

	// Выбор типа поиска
	searchTypeSelect := widget.NewSelect([]string{
		"LIKE (стандартный поиск)",
		"SIMILAR TO (SQL стандарт)",
		"POSIX ~ (регулярное выражение)",
		"POSIX ~* (регулярное выражение, без учета регистра)",
		"POSIX !~ (отрицание регулярного выражения)",
		"POSIX !~* (отрицание, без учета регистра)",
	}, nil)
	searchTypeSelect.PlaceHolder = "Выберите тип поиска"
	searchTypeSelect.SetSelected("LIKE (стандартный поиск)")

	// Столбец для поиска
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Название столбца (name, description...)")

	// Паттерн для поиска
	patternEntry := widget.NewEntry()
	patternEntry.SetPlaceHolder("Введите паттерн поиска")
	patternEntry.MultiLine = true
	patternEntry.SetMinRowsVisible(3)

	// Информационная подсказка
	infoLabel := widget.NewRichTextFromMarkdown(`
**Примеры паттернов:**

**LIKE:** 
- %товар% (содержит "товар")
- А% (начинается с "А")

**SIMILAR TO:**
- A%B (начинается с А, заканчивается на В)
- (A|B)% (начинается с А или В)

**POSIX Regex:**
- ^А.* (начинается с А)
- [0-9]{3} (ровно 3 цифры)
- (cat|dog)$ (заканчивается на cat или dog)
`)

	// Таблица результатов
	var resultsTable widget.Table
	var resultsData [][]string

	// Кнопка поиска
	executeButton := widget.NewButton("Выполнить поиск", func() {
		tableName := strings.TrimSpace(tableEntry.Text)
		column := strings.TrimSpace(columnEntry.Text)
		pattern := strings.TrimSpace(patternEntry.Text)
		searchType := searchTypeSelect.Selected

		if tableName == "" || column == "" || pattern == "" {
			showError(window, "Пожалуйста, заполните все поля")
			return
		}

		// Создаём QueryBuilder
		qb := operation.NewQueryBuilder(tableName)
		qb.Select("*")

		// Применяем поиск в зависимости от типа
		switch searchType {
		case "LIKE (стандартный поиск)":
			qb.WhereLike(column, pattern)
		case "SIMILAR TO (SQL стандарт)":
			qb.WhereSimilarTo(column, pattern)
		case "POSIX ~ (регулярное выражение)":
			qb.WhereRegex(column, pattern)
		case "POSIX ~* (регулярное выражение, без учета регистра)":
			qb.WhereRegexNoCase(column, pattern)
		case "POSIX !~ (отрицание регулярного выражения)":
			qb.WhereNotRegex(column, pattern)
		case "POSIX !~* (отрицание, без учета регистра)":
			qb.WhereNotRegexNoCase(column, pattern)
		}

		// Выполняем запрос
		results, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, fmt.Sprintf("Ошибка при выполнении поиска: %v", err))
			return
		}

		resultsData = results
		resultsTable.Refresh()

		showInfo(window, fmt.Sprintf("Найдено %d результатов", len(results)-1))
	})

	// Создание результативной таблицы
	resultsTable = *widget.NewTable(
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
			}
		},
	)

	// Формирование интерфейса
	form := container.NewVBox(
		widget.NewCard("Параметры поиска", "", container.NewVBox(
			widget.NewForm(
				widget.NewFormItem("Таблица", tableEntry),
				widget.NewFormItem("Столбец", columnEntry),
			),
			widget.NewCard("Тип поиска", "", container.NewVBox(searchTypeSelect)),
			widget.NewForm(
				widget.NewFormItem("Паттерн", patternEntry),
			),
		)),
		executeButton,
		widget.NewCard("Справка", "", infoLabel),
	)

	resultsContainer := container.NewScroll(container.NewVBox(&resultsTable))

	content := container.NewVBox(
		container.NewScroll(form),
		widget.NewCard("Результаты", "", resultsContainer),
	)

	searchWindow := fyne.CurrentApp().NewWindow("Window")
	searchWindow.SetTitle("Поиск по строкам с регулярными выражениями")
	searchWindow.SetContent(content)
	searchWindow.Resize(fyne.NewSize(1000, 700))
	searchWindow.CenterOnScreen()
	searchWindow.Show()
}

// ===== ТРЕБОВАНИЕ 1: UIStringFunctions - ФУНКЦИИ РАБОТЫ СО СТРОКАМИ =====

// UIStringFunctions создаёт интерфейс для работы со строками (UPPER, LOWER, TRIM, SUBSTRING и т.д.)
func UIStringFunctions(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Таблица")

	// Выбор функции
	functionSelect := widget.NewSelect([]string{
		"UPPER - в верхний регистр",
		"LOWER - в нижний регистр",
		"TRIM - удалить пробелы",
		"LTRIM - удалить пробелы слева",
		"RTRIM - удалить пробелы справа",
		"LENGTH - длина строки",
		"SUBSTRING - подстрока",
		"CONCAT - объединение строк",
	}, nil)
	functionSelect.PlaceHolder = "Выберите функцию"
	functionSelect.SetSelected("UPPER - в верхний регистр")

	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Название столбца")

	// Параметры (для SUBSTRING, CONCAT и т.д.)
	startEntry := widget.NewEntry()
	startEntry.SetPlaceHolder("Начало (для SUBSTRING)")

	lengthEntry := widget.NewEntry()
	lengthEntry.SetPlaceHolder("Длина (для SUBSTRING)")

	aliasEntry := widget.NewEntry()
	aliasEntry.SetPlaceHolder("Alias (название результирующего столбца)")

	// Кнопка выполнения
	executeButton := widget.NewButton("Применить функцию", func() {
		tableName := strings.TrimSpace(tableEntry.Text)
		column := strings.TrimSpace(columnEntry.Text)
		alias := strings.TrimSpace(aliasEntry.Text)
		fn := functionSelect.Selected

		if tableName == "" || column == "" || fn == "" {
			showError(window, "Заполните таблицу, столбец и функцию")
			return
		}

		if alias == "" {
			alias = column + "_result"
		}

		qb := operation.NewQueryBuilder(tableName)
		qb.Select("id") // Добавляем ID для связи

		// Применяем выбранную функцию
		switch fn {
		case "UPPER - в верхний регистр":
			qb.SelectUpper(column)
		case "LOWER - в нижний регистр":
			qb.SelectLower(column)
		case "TRIM - удалить пробелы":
			qb.SelectTrim(column)
		case "LTRIM - удалить пробелы слева":
			qb.SelectLTrim(column)
		case "RTRIM - удалить пробелы справа":
			qb.SelectRTrim(column)
		case "SUBSTRING - подстрока":
			start, _ := strconv.Atoi(startEntry.Text)
			length, _ := strconv.Atoi(lengthEntry.Text)
			if start == 0 {
				start = 1
			}
			if length == 0 {
				length = 50
			}
			qb.SelectSubstring(column, start, length)
		}

		qb.Limit(10)

		results, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, fmt.Sprintf("Ошибка: %v", err))
			return
		}

		// Показываем результаты
		resultTable, _ := CreateTable(results)
		resultWindow := fyne.CurrentApp().NewWindow("Window2")
		resultWindow.SetTitle(fmt.Sprintf("Результаты функции %s", fn))
		resultWindow.SetContent(container.NewScroll(resultTable))
		resultWindow.Resize(fyne.NewSize(600, 400))
		resultWindow.CenterOnScreen()
		resultWindow.Show()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Таблица", tableEntry),
			widget.NewFormItem("Функция", functionSelect),
			widget.NewFormItem("Столбец", columnEntry),
			widget.NewFormItem("Начало (SUBSTRING)", startEntry),
			widget.NewFormItem("Длина (SUBSTRING)", lengthEntry),
			widget.NewFormItem("Alias результата", aliasEntry),
		),
		executeButton,
		widget.NewRichTextFromMarkdown(`
**Доступные функции:**
- **UPPER/LOWER** - преобразование регистра
- **TRIM/LTRIM/RTRIM** - удаление пробелов
- **SUBSTRING** - извлечение подстроки (параметры: начало, длина)
- **LENGTH** - длина строки (можно использовать в WHERE)
- **CONCAT** - объединение нескольких столбцов
`),
	)

	stringWindow := fyne.CurrentApp().NewWindow("Window3")
	stringWindow.SetTitle("Функции работы со строками")
	stringWindow.SetContent(container.NewScroll(form))
	stringWindow.Resize(fyne.NewSize(600, 500))
	stringWindow.CenterOnScreen()
	stringWindow.Show()
}

// ===== ТРЕБОВАНИЕ 1: UIJoinWizard - МАСТЕР ДЛЯ СОЗДАНИЯ JOIN ЗАПРОСОВ =====

// UIJoinWizard создаёт интерфейс для создания JOIN запросов
func UIJoinWizard(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	mainTableEntry := widget.NewEntry()
	mainTableEntry.SetPlaceHolder("Основная таблица")
	mainTableEntry.SetText("products")

	joinTableEntry := widget.NewEntry()
	joinTableEntry.SetPlaceHolder("Таблица для JOIN")
	joinTableEntry.SetText("categories")

	joinTypeSelect := widget.NewSelect([]string{
		"INNER JOIN",
		"LEFT JOIN",
		"RIGHT JOIN",
		"FULL OUTER JOIN",
	}, nil)
	joinTypeSelect.PlaceHolder = "Тип JOIN"
	joinTypeSelect.SetSelected("INNER JOIN")

	onConditionEntry := widget.NewEntry()
	onConditionEntry.SetPlaceHolder("Условие ON (например: products.category_id = categories.id)")
	onConditionEntry.SetText("products.category_id = categories.id")

	columnsEntry := widget.NewEntry()
	columnsEntry.SetPlaceHolder("Столбцы для выбора (через запятую)")
	columnsEntry.SetText("products.name, categories.name as category_name, products.price")
	columnsEntry.MultiLine = true
	columnsEntry.SetMinRowsVisible(3)

	limitEntry := widget.NewEntry()
	limitEntry.SetPlaceHolder("Лимит результатов")
	limitEntry.SetText("10")

	// Кнопка выполнения
	executeButton := widget.NewButton("Выполнить JOIN запрос", func() {
		mainTable := strings.TrimSpace(mainTableEntry.Text)
		joinTable := strings.TrimSpace(joinTableEntry.Text)
		onCondition := strings.TrimSpace(onConditionEntry.Text)
		columns := strings.TrimSpace(columnsEntry.Text)
		limitStr := strings.TrimSpace(limitEntry.Text)

		if mainTable == "" || joinTable == "" || onCondition == "" {
			showError(window, "Заполните основную таблицу, таблицу JOIN и условие ON")
			return
		}

		limit := 10
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}

		qb := operation.NewQueryBuilder(mainTable)

		if columns != "" {
			columnsSlice := strings.Split(columns, ",")
			for i, col := range columnsSlice {
				columnsSlice[i] = strings.TrimSpace(col)
			}
			qb.Select(columnsSlice...)
		} else {
			qb.Select("*")
		}

		// Добавляем JOIN в зависимости от типа
		switch joinTypeSelect.Selected {
		case "INNER JOIN":
			qb.InnerJoin(joinTable, onCondition)
		case "LEFT JOIN":
			qb.LeftJoin(joinTable, onCondition)
		case "RIGHT JOIN":
			qb.RightJoin(joinTable, onCondition)
		case "FULL OUTER JOIN":
			qb.FullJoin(joinTable, onCondition)
		}

		qb.Limit(limit)

		results, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, fmt.Sprintf("Ошибка выполнения JOIN: %v", err))
			return
		}

		// Показываем результаты
		resultTable, _ := CreateTable(results)
		resultWindow := fyne.CurrentApp().NewWindow("Window4")
		resultWindow.SetTitle(fmt.Sprintf("%s - Результаты", joinTypeSelect.Selected))
		resultWindow.SetContent(container.NewVBox(
			widget.NewCard("SQL запрос", "", widget.NewLabel(qb.Build())),
			container.NewScroll(resultTable),
		))
		resultWindow.Resize(fyne.NewSize(900, 600))
		resultWindow.CenterOnScreen()
		resultWindow.Show()

		showInfo(window, fmt.Sprintf("Найдено %d результатов", len(results)-1))
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Основная таблица", mainTableEntry),
			widget.NewFormItem("Таблица для JOIN", joinTableEntry),
			widget.NewFormItem("Тип JOIN", joinTypeSelect),
			widget.NewFormItem("Условие ON", onConditionEntry),
		),
		widget.NewForm(
			widget.NewFormItem("Столбцы", columnsEntry),
			widget.NewFormItem("Лимит", limitEntry),
		),
		executeButton,
		widget.NewRichTextFromMarkdown(`
**Типы JOIN:**
- **INNER JOIN** - пересечение (только совпадающие строки)
- **LEFT JOIN** - все из левой таблицы + совпадающие из правой
- **RIGHT JOIN** - совпадающие из левой + все из правой
- **FULL OUTER JOIN** - все из обеих таблиц

**Пример условия ON:**
- products.category_id = categories.id
- t1.id = t2.parent_id AND t1.status = t2.status
`),
	)

	joinWindow := fyne.CurrentApp().NewWindow("Window5")
	joinWindow.SetTitle("Мастер создания JOIN")
	joinWindow.SetContent(container.NewScroll(form))
	joinWindow.Resize(fyne.NewSize(700, 600))
	joinWindow.CenterOnScreen()
	joinWindow.Show()
}

// Вспомогательные функции для UI
func showError(window fyne.Window, message string) {
	dialog.ShowError(fmt.Errorf(message), window)
}

func showInfo(window fyne.Window, message string) {
	dialog.ShowInformation("Информация", message, window)
}

// CreateTable создаёт таблицу из данных
func CreateTable(data [][]string) (*widget.Table, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("нет данных для отображения")
	}

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
		func(i widget.TableCellID, o fyne.CanvasObject) {
			if i.Row < len(data) && i.Col < len(data[i.Row]) {
				o.(*widget.Label).SetText(data[i.Row][i.Col])
			}
		},
	)

	// Установка ширины колонок
	for col := 0; col < len(data[0]); col++ {
		table.SetColumnWidth(col, 150)
	}

	return table, nil
}
