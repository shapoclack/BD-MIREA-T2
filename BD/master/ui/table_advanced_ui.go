package table

import (
	operation "BD_Mirea/internal"
	"context"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateAdvancedUI создаёт расширенное UI с доступом ко всем функциям
func CreateAdvancedUI(window fyne.Window, ctx context.Context, pool *pgxpool.Pool) {
	// Создаем главное меню
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Таблицы",
			fyne.NewMenuItem("Создать таблицу", func() {
				UICreateTablesWithTypes(ctx, pool, window)
			}),
			fyne.NewMenuItem("Переименовать таблицу", func() {
				UIRenameTable(ctx, pool, window)
			}),
		),
		fyne.NewMenu("Столбцы",
			fyne.NewMenuItem("Добавить столбец", func() {
				UIAddColumn(ctx, pool, window)
			}),
			fyne.NewMenuItem("Удалить столбец", func() {
				UIDropColumn(ctx, pool, window)
			}),
			fyne.NewMenuItem("Изменить тип столбца", func() {
				UIAlterColumnType(ctx, pool, window)
			}),
			fyne.NewMenuItem("Переименовать столбец", func() {
				UIRenameColumn(ctx, pool, window)
			}),
		),
		fyne.NewMenu("Ограничения",
			fyne.NewMenuItem("Добавить CHECK", func() {
				UIAddCheck(ctx, pool, window)
			}),
			fyne.NewMenuItem("Добавить UNIQUE", func() {
				UIAddUnique(ctx, pool, window)
			}),
			fyne.NewMenuItem("Добавить FOREIGN KEY", func() {
				UIAddForeignKey(ctx, pool, window)
			}),
			fyne.NewMenuItem("Удалить ограничение", func() {
				UIDropConstraint(ctx, pool, window)
			}),
			fyne.NewMenuItem("Установить NOT NULL", func() {
				UISetNotNull(ctx, pool, window)
			}),
			fyne.NewMenuItem("Удалить NOT NULL", func() {
				UIDropNotNull(ctx, pool, window)
			}),
		),
		fyne.NewMenu("Запросы",
			fyne.NewMenuItem("Query Builder", func() {
				UIQueryBuilder(ctx, pool, window)
			}),
		),
		fyne.NewMenu("Поиск & Функции",
			fyne.NewMenuItem("Поиск по тексту (LIKE & REGEX)", func() {
				UISearchDialog(ctx, pool, window, "products")
			}),
			fyne.NewMenuItem("Функции преобразования строк", func() {
				UIStringFunctions(ctx, pool, window, "products")
			}),
			fyne.NewMenuItem("Мастер соединений (JOIN)", func() {
				UIJoinWizard(ctx, pool, window)
			}),
		),
		fyne.NewMenu("Подключение",
			fyne.NewMenuItem("Тест подключения", func() {
				UITestConnection(ctx, pool, window)
			}),
		),
	)

	window.SetMainMenu(mainMenu)

	// ===== НОВЫЙ КОД: Таблица при запуске =====
	currentTableName := "products"
	tableData, err := operation.GetAllProducts(ctx, pool)
	if err != nil {
		// Если ошибка, показываем старое приветственное сообщение
		welcomeCard := widget.NewCard(
			"Добро пожаловать в PostgreSQL UI Client!",
			"",
			container.NewVBox(
				widget.NewLabel("Выберите операцию из меню выше"),
				widget.NewLabel(""),
				widget.NewLabel("Ошибка загрузки данных: "+err.Error()),
			),
		)
		window.SetContent(container.NewCenter(welcomeCard))
		return
	}

	// Создаём виджет таблицы с обрезкой текста
	tableWidget := widget.NewTable(
		func() (int, int) {
			if len(tableData) == 0 {
				return 0, 0
			}
			return len(tableData), len(tableData[0])
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("cell")
			label.Truncation = fyne.TextTruncateEllipsis // Обрезка длинного текста
			return label
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row < len(tableData) && id.Col < len(tableData[id.Row]) {
				text := tableData[id.Row][id.Col]

				// Ограничиваем длину текста для предотвращения перекрытия
				maxLen := 60
				if id.Col == 2 { // description
					maxLen = 40
				} else if id.Col == 3 { // created_at
					maxLen = 30
				}

				if len(text) > maxLen {
					text = text[:maxLen-3] + "..."
				}

				label.SetText(text)
				if id.Row == 0 {
					label.TextStyle = fyne.TextStyle{Bold: true}
					label.Importance = widget.HighImportance
				}
			}
		},
	)

	// Обработчик клика для редактирования
	tableWidget.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return // Не редактируем заголовки
		}

		entry := widget.NewEntry()
		entry.SetText(tableData[id.Row][id.Col])

		dlg := dialog.NewCustomConfirm(
			"Редактировать значение",
			"Сохранить",
			"Отмена",
			entry,
			func(save bool) {
				if save {
					tableData[id.Row][id.Col] = entry.Text
					tableWidget.Refresh()
					showInfo(window, "Значение обновлено!")
				}
			},
			window,
		)
		dlg.Show()
	}

	// Автоматическая настройка ширины колонок
	setOptimalColumnWidths(tableWidget, tableData)

	infoLabel := widget.NewLabel(fmt.Sprintf("Таблица: %s | Строк: %d", currentTableName, len(tableData)-1))

	// ВАЖНО: Получаем список таблиц и создаём tableSelect ДО создания кнопок
	tablesList, _ := getTablesListFromDB(ctx, pool)
	tableSelect := widget.NewSelect(tablesList, func(selected string) {
		currentTableName = selected
		loadTableByName(ctx, pool, selected, &tableData, tableWidget, infoLabel)
	})
	if len(tablesList) > 0 {
		tableSelect.SetSelected(currentTableName)
	}

	// ТЕПЕРЬ можно создавать кнопки
	createTableBtn := widget.NewButton("➕ Создать таблицу", func() {
		UICreateTablesWithTypesButton(ctx, pool, window, &tableData, tableWidget, infoLabel, &currentTableName, tableSelect)
	})

	deleteTableBtn := widget.NewButton("🗑 Удалить таблицу", func() {
		showDeleteTableDialog(ctx, pool, window, &currentTableName, &tableData, tableWidget, infoLabel, tableSelect)
	})

	refreshBtn := widget.NewButton("🔄 Обновить", func() {
		loadTableByName(ctx, pool, currentTableName, &tableData, tableWidget, infoLabel)
	})

	addRowBtn := widget.NewButton("➕ Добавить строку", func() {
		showAddRowDialogAdvanced(ctx, pool, window, currentTableName, &tableData, tableWidget, infoLabel)
	})

	deleteRowBtn := widget.NewButton("🗑 Удалить строку", func() {
		showDeleteRowDialogAdvanced(ctx, pool, window, currentTableName, &tableData, tableWidget, infoLabel)
	})

	// Панель управления
	toolbar := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Выбрать таблицу:"),
			tableSelect,
			createTableBtn,
			deleteTableBtn,
		),
		container.NewHBox(
			refreshBtn,
			addRowBtn,
			deleteRowBtn,
		),
		infoLabel,
		widget.NewSeparator(),
	)

	// Основной контент
	mainContent := container.NewBorder(
		toolbar,
		nil, nil, nil,
		container.NewScroll(tableWidget),
	)

	window.SetContent(mainContent)
}

// setOptimalColumnWidths автоматически устанавливает оптимальную ширину колонок
func setOptimalColumnWidths(table *widget.Table, data [][]string) {
	if len(data) == 0 {
		return
	}

	// Минимальная и максимальная ширина
	const minWidth = 50.0
	const maxWidth = 400.0

	for col := 0; col < len(data[0]); col++ {
		maxLen := 0

		// Находим максимальную длину текста в столбце
		for row := 0; row < len(data) && row < 10; row++ { // Проверяем только первые 10 строк
			if col < len(data[row]) {
				textLen := len(data[row][col])
				if textLen > maxLen {
					maxLen = textLen
				}
			}
		}

		// Рассчитываем ширину: ~7 пикселей на символ
		width := float32(maxLen * 7)
		if width < minWidth {
			width = minWidth
		}
		if width > maxWidth {
			width = maxWidth
		}

		table.SetColumnWidth(col, width)
	}
}

//  Функции UI диалогов

// UITestConnection проверяет подключение к БД
func UITestConnection(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	err := operation.TestConnection(ctx, pool)
	if err != nil {
		showError(window, "Ошибка подключения: "+err.Error())
		return
	}
	showInfo(window, "Подключение к базе данных успешно!")
}

// UICreateTablesWithTypes создаёт диалог для создания таблицы с типами
func UICreateTablesWithTypes(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableNameEntry := widget.NewEntry()
	tableNameEntry.SetPlaceHolder("Имя таблицы")

	columnsEntry := widget.NewMultiLineEntry()
	columnsEntry.SetPlaceHolder("Столбцы (формат: имя тип ограничения)\nПример:\nid SERIAL PRIMARY KEY\nname VARCHAR(100) NOT NULL")
	columnsEntry.SetMinRowsVisible(8)

	form := container.NewVBox(
		widget.NewLabel("Создание таблицы с типами"),
		widget.NewForm(
			widget.NewFormItem("Имя таблицы", tableNameEntry),
		),
		widget.NewLabel("Определения столбцов:"),
		columnsEntry,
	)

	dlg := dialog.NewCustomConfirm("Создать таблицу", "Создать", "Отмена", form, func(ok bool) {
		if ok {
			tableName := strings.TrimSpace(tableNameEntry.Text)
			if tableName == "" {
				showError(window, "Укажите имя таблицы")
				return
			}

			lines := strings.Split(columnsEntry.Text, "\n")
			var columns []operation.ColumnDefinition
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) < 2 {
					showError(window, "Неверный формат столбца: "+line)
					return
				}
				col := operation.ColumnDefinition{
					Name: parts[0],
					Type: parts[1],
				}
				if len(parts) > 2 {
					col.Constraints = strings.Join(parts[2:], " ")
				}
				columns = append(columns, col)
			}

			if len(columns) == 0 {
				showError(window, "Необходимо указать хотя бы один столбец")
				return
			}

			err := operation.CreateTablesWithTypes(ctx, pool, tableName, columns)
			if err != nil {
				showError(window, "Ошибка создания таблицы: "+err.Error())
				return
			}

			showInfo(window, fmt.Sprintf("Таблица '%s' успешно создана!", tableName))
		}
	}, window)

	dlg.Resize(fyne.NewSize(600, 500))
	dlg.Show()
}

// UICreateTablesWithTypesButton - кнопка создания таблицы с обновлением UI
func UICreateTablesWithTypesButton(ctx context.Context, pool *pgxpool.Pool, window fyne.Window,
	dataPtr *[][]string, table *widget.Table, infoLabel *widget.Label,
	currentTable *string, tableSelect *widget.Select) {

	tableNameEntry := widget.NewEntry()
	tableNameEntry.SetPlaceHolder("Имя таблицы")

	columnsEntry := widget.NewMultiLineEntry()
	columnsEntry.SetPlaceHolder("Столбцы (формат: имя тип ограничения)\nПример:\nid SERIAL PRIMARY KEY\nname VARCHAR(100) NOT NULL\nemail VARCHAR(255) UNIQUE")
	columnsEntry.SetMinRowsVisible(8)

	form := container.NewVBox(
		widget.NewLabel("Создание новой таблицы"),
		widget.NewForm(
			widget.NewFormItem("Имя таблицы", tableNameEntry),
		),
		widget.NewLabel("Определения столбцов:"),
		columnsEntry,
	)

	dlg := dialog.NewCustomConfirm("Создать таблицу", "Создать", "Отмена", form, func(ok bool) {
		if ok {
			tableName := strings.TrimSpace(tableNameEntry.Text)
			if tableName == "" {
				showError(window, "Укажите имя таблицы")
				return
			}

			lines := strings.Split(columnsEntry.Text, "\n")
			var columns []operation.ColumnDefinition
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) < 2 {
					showError(window, "Неверный формат: "+line)
					return
				}
				col := operation.ColumnDefinition{
					Name: parts[0],
					Type: parts[1],
				}
				if len(parts) > 2 {
					col.Constraints = strings.Join(parts[2:], " ")
				}
				columns = append(columns, col)
			}

			if len(columns) == 0 {
				showError(window, "Необходимо указать хотя бы один столбец")
				return
			}

			err := operation.CreateTablesWithTypes(ctx, pool, tableName, columns)
			if err != nil {
				showError(window, "Ошибка: "+err.Error())
				return
			}

			showInfo(window, fmt.Sprintf("Таблица '%s' создана!", tableName))

			// Обновление списка таблиц
			tablesList, _ := getTablesListFromDB(ctx, pool)
			tableSelect.Options = tablesList
			*currentTable = tableName
			tableSelect.SetSelected(tableName)
			loadTableByName(ctx, pool, tableName, dataPtr, table, infoLabel)
		}
	}, window)

	dlg.Resize(fyne.NewSize(600, 500))
	dlg.Show()
}

// UIRenameTable переименование таблицы
func UIRenameTable(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	oldTableEntry := widget.NewEntry()
	oldTableEntry.SetPlaceHolder("Текущее имя таблицы")
	newTableEntry := widget.NewEntry()
	newTableEntry.SetPlaceHolder("Новое имя таблицы")

	form := widget.NewForm(
		widget.NewFormItem("Текущее имя", oldTableEntry),
		widget.NewFormItem("Новое имя", newTableEntry),
	)

	dialog.ShowCustomConfirm("Переименовать таблицу", "Переименовать", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.RenameTable(ctx, pool,
				strings.TrimSpace(oldTableEntry.Text),
				strings.TrimSpace(newTableEntry.Text))
			if err != nil {
				showError(window, "Ошибка переименования таблицы: "+err.Error())
				return
			}
			showInfo(window, "Таблица успешно переименована!")
		}
	}, window)

}

// UIAddColumn  добавление столбца
func UIAddColumn(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Имя столбца")
	typeEntry := widget.NewEntry()
	typeEntry.SetPlaceHolder("Тип данных (например, VARCHAR(100))")
	constraintsEntry := widget.NewEntry()
	constraintsEntry.SetPlaceHolder("Ограничения (например, NOT NULL)")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Столбец", columnEntry),
		widget.NewFormItem("Тип", typeEntry),
		widget.NewFormItem("Ограничения", constraintsEntry),
	)

	dialog.ShowCustomConfirm("Добавить столбец", "Добавить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.AddColumn(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(columnEntry.Text),
				strings.TrimSpace(typeEntry.Text),
				strings.TrimSpace(constraintsEntry.Text))
			if err != nil {
				showError(window, "Ошибка добавления столбца: "+err.Error())
				return
			}
			showInfo(window, "Столбец успешно добавлен!")
		}
	}, window)
}

// UIDropColumn удаление столбца
func UIDropColumn(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Имя столбца")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Столбец", columnEntry),
	)

	dialog.ShowCustomConfirm("Удалить столбец", "Удалить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.DropColumn(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(columnEntry.Text))
			if err != nil {
				showError(window, "Ошибка удаления столбца: "+err.Error())
				return
			}
			showInfo(window, "Столбец успешно удален!")
		}
	}, window)
}

// UIAlterColumnType изменения типа столбца
func UIAlterColumnType(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Имя столбца")
	newTypeEntry := widget.NewEntry()
	newTypeEntry.SetPlaceHolder("Новый тип (например, TEXT)")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Столбец", columnEntry),
		widget.NewFormItem("Новый тип", newTypeEntry),
	)

	dialog.ShowCustomConfirm("Изменить тип столбца", "Изменить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.AlterColumnType(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(columnEntry.Text),
				strings.TrimSpace(newTypeEntry.Text))
			if err != nil {
				showError(window, "Ошибка изменения типа: "+err.Error())
				return
			}
			showInfo(window, "Тип столбца успешно изменен!")
		}
	}, window)
}

// UIRenameColumn переименование столбца
func UIRenameColumn(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	oldColumnEntry := widget.NewEntry()
	oldColumnEntry.SetPlaceHolder("Текущее имя столбца")
	newColumnEntry := widget.NewEntry()
	newColumnEntry.SetPlaceHolder("Новое имя столбца")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Текущее имя", oldColumnEntry),
		widget.NewFormItem("Новое имя", newColumnEntry),
	)

	dialog.ShowCustomConfirm("Переименовать столбец", "Переименовать", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.RenameColumn(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(oldColumnEntry.Text),
				strings.TrimSpace(newColumnEntry.Text))
			if err != nil {
				showError(window, "Ошибка переименования столбца: "+err.Error())
				return
			}
			showInfo(window, "Столбец успешно переименован!")
		}
	}, window)
}

// ========== UI для операций с ограничениями ==========

// UIAddCheck  для добавление CHECK ограничения
func UIAddCheck(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	constraintNameEntry := widget.NewEntry()
	constraintNameEntry.SetPlaceHolder("Имя ограничения")
	expressionEntry := widget.NewEntry()
	expressionEntry.SetPlaceHolder("Условие (например, price > 0)")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Имя ограничения", constraintNameEntry),
		widget.NewFormItem("Условие", expressionEntry),
	)

	dialog.ShowCustomConfirm("Добавить CHECK", "Добавить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.AddCheck(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(constraintNameEntry.Text),
				strings.TrimSpace(expressionEntry.Text))
			if err != nil {
				showError(window, "Ошибка добавления CHECK: "+err.Error())
				return
			}
			showInfo(window, "CHECK ограничение успешно добавлено!")
		}
	}, window)
}

// UIDropConstraint удалениe ограничения
func UIDropConstraint(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	constraintNameEntry := widget.NewEntry()
	constraintNameEntry.SetPlaceHolder("Имя ограничения")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Имя ограничения", constraintNameEntry),
	)

	dialog.ShowCustomConfirm("Удалить ограничение", "Удалить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.DropConstraint(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(constraintNameEntry.Text))
			if err != nil {
				showError(window, "Ошибка удаления ограничения: "+err.Error())
				return
			}
			showInfo(window, "Ограничение успешно удалено!")
		}
	}, window)
}

// UISetNotNull для установка NOT NULL
func UISetNotNull(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Имя столбца")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Столбец", columnEntry),
	)

	dialog.ShowCustomConfirm("Установить NOT NULL", "Установить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.SetNotNull(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(columnEntry.Text))
			if err != nil {
				showError(window, "Ошибка установки NOT NULL: "+err.Error())
				return
			}
			showInfo(window, "NOT NULL успешно установлен!")
		}
	}, window)
}

// UIDropNotNull удаление NOT NULL
func UIDropNotNull(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Имя столбца")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Столбец", columnEntry),
	)

	dialog.ShowCustomConfirm("Удалить NOT NULL", "Удалить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.DropNotNull(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(columnEntry.Text))
			if err != nil {
				showError(window, "Ошибка удаления NOT NULL: "+err.Error())
				return
			}
			showInfo(window, "NOT NULL успешно удален!")
		}
	}, window)
}

// UIAddUnique для добавления UNIQUE ограничения
func UIAddUnique(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	constraintNameEntry := widget.NewEntry()
	constraintNameEntry.SetPlaceHolder("Имя ограничения")
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Имя столбца")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Имя ограничения", constraintNameEntry),
		widget.NewFormItem("Столбец", columnEntry),
	)

	dialog.ShowCustomConfirm("Добавить UNIQUE", "Добавить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.AddUnique(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(constraintNameEntry.Text),
				strings.TrimSpace(columnEntry.Text))
			if err != nil {
				showError(window, "Ошибка добавления UNIQUE: "+err.Error())
				return
			}
			showInfo(window, "UNIQUE ограничение успешно добавлено!")
		}
	}, window)
}

// UIAddForeignKey добавление FOREIGN KEY
func UIAddForeignKey(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	constraintNameEntry := widget.NewEntry()
	constraintNameEntry.SetPlaceHolder("Имя ограничения")
	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Столбец")
	refTableEntry := widget.NewEntry()
	refTableEntry.SetPlaceHolder("Ссылочная таблица")
	refColumnEntry := widget.NewEntry()
	refColumnEntry.SetPlaceHolder("Ссылочный столбец")

	form := widget.NewForm(
		widget.NewFormItem("Таблица", tableEntry),
		widget.NewFormItem("Имя ограничения", constraintNameEntry),
		widget.NewFormItem("Столбец", columnEntry),
		widget.NewFormItem("Ссылочная таблица", refTableEntry),
		widget.NewFormItem("Ссылочный столбец", refColumnEntry),
	)

	dialog.ShowCustomConfirm("Добавить FOREIGN KEY", "Добавить", "Отмена", form, func(ok bool) {
		if ok {
			err := operation.AddForeignKey(ctx, pool,
				strings.TrimSpace(tableEntry.Text),
				strings.TrimSpace(constraintNameEntry.Text),
				strings.TrimSpace(columnEntry.Text),
				strings.TrimSpace(refTableEntry.Text),
				strings.TrimSpace(refColumnEntry.Text))
			if err != nil {
				showError(window, "Ошибка добавления FOREIGN KEY: "+err.Error())
				return
			}
			showInfo(window, "FOREIGN KEY успешно добавлен!")
		}
	}, window)
}

// ========== UI для Query Builder ==========

// UIQueryBuilder создаёт интерактивный построитель запросов
func UIQueryBuilder(ctx context.Context, pool *pgxpool.Pool, window fyne.Window) {
	tableEntry := widget.NewEntry()
	tableEntry.SetPlaceHolder("Имя таблицы")
	tableEntry.SetText("products")

	columnsEntry := widget.NewEntry()
	columnsEntry.SetPlaceHolder("Столбцы через запятую (оставьте пустым для SELECT *)")

	whereEntry := widget.NewEntry()
	whereEntry.SetPlaceHolder("WHERE условие")

	limitEntry := widget.NewEntry()
	limitEntry.SetPlaceHolder("LIMIT")

	queryPreview := widget.NewMultiLineEntry()
	queryPreview.SetPlaceHolder("SQL запрос")
	queryPreview.Disable()
	queryPreview.SetMinRowsVisible(3)

	var resultsTable *widget.Table
	var resultsData [][]string

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
		func(i widget.TableCellID, o fyne.CanvasObject) {
			if i.Row < len(resultsData) && i.Col < len(resultsData[i.Row]) {
				o.(*widget.Label).SetText(resultsData[i.Row][i.Col])
			}
		},
	)

	executeButton := widget.NewButton("Выполнить", func() {
		tableName := strings.TrimSpace(tableEntry.Text)
		if tableName == "" {
			showError(window, "Укажите имя таблицы")
			return
		}

		qb := operation.NewQueryBuilder(tableName)

		if where := strings.TrimSpace(whereEntry.Text); where != "" {
			qb.Where(where)
		}

		if limitStr := strings.TrimSpace(limitEntry.Text); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				qb.Limit(limit)
			}
		}

		queryPreview.SetText(qb.Build())

		results, err := qb.Execute(ctx, pool)
		if err != nil {
			showError(window, "Ошибка: "+err.Error())
			return
		}

		resultsData = results
		resultsTable.Refresh()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Таблица", tableEntry),
			widget.NewFormItem("WHERE", whereEntry),
			widget.NewFormItem("LIMIT", limitEntry),
		),
		executeButton,
		widget.NewLabel("SQL:"),
		queryPreview,
		widget.NewLabel("Результаты:"),
		container.NewScroll(resultsTable),
	)

	qbWindow := fyne.CurrentApp().NewWindow("Query Builder")
	qbWindow.SetContent(container.NewScroll(form))
	qbWindow.Resize(fyne.NewSize(800, 600))
	qbWindow.CenterOnScreen()
	qbWindow.Show()
}

// ========== НОВЫЕ ФУНКЦИИ ==========

// Универсальное добавление строки для любой таблицы
func showAddRowDialogAdvanced(ctx context.Context, pool *pgxpool.Pool, window fyne.Window,
	tableName string, dataPtr *[][]string, table *widget.Table, infoLabel *widget.Label) {

	if len(*dataPtr) == 0 || len((*dataPtr)[0]) == 0 {
		showError(window, "Нет данных о структуре таблицы")
		return
	}

	headers := (*dataPtr)[0]

	var entries []*widget.Entry
	var formItems []*widget.FormItem
	var columnNames []string

	for _, colName := range headers {
		if strings.ToLower(colName) == "id" {
			continue
		}

		entry := widget.NewEntry()
		entry.SetPlaceHolder("Введите " + colName)
		entries = append(entries, entry)
		columnNames = append(columnNames, colName)
		formItems = append(formItems, widget.NewFormItem(colName, entry))
	}

	if len(formItems) == 0 {
		showError(window, "Нет полей для ввода")
		return
	}

	form := widget.NewForm(formItems...)

	dlg := dialog.NewCustomConfirm("Добавить строку", "Добавить", "Отмена", form, func(ok bool) {
		if ok {
			values := make([]string, len(entries))
			for i, entry := range entries {
				values[i] = entry.Text
			}

			err := insertRowGeneric(ctx, pool, tableName, columnNames, values)
			if err != nil {
				showError(window, "Ошибка добавления: "+err.Error())
				return
			}

			showInfo(window, "Строка успешно добавлена!")
			loadTableByName(ctx, pool, tableName, dataPtr, table, infoLabel)
		}
	}, window)

	dlg.Resize(fyne.NewSize(500, 400))
	dlg.Show()
}

// Универсальная функция вставки строки
func insertRowGeneric(ctx context.Context, pool *pgxpool.Pool, tableName string,
	columnNames []string, values []string) error {

	if len(columnNames) != len(values) {
		return fmt.Errorf("количество колонок (%d) не совпадает с количеством значений (%d)",
			len(columnNames), len(values))
	}

	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i := range values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = values[i]
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := pool.Exec(ctx, query, args...)
	return err
}

// Диалог удаления строки
func showDeleteRowDialogAdvanced(ctx context.Context, pool *pgxpool.Pool, window fyne.Window,
	tableName string, dataPtr *[][]string, table *widget.Table, infoLabel *widget.Label) {

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("ID строки для удаления")

	form := widget.NewForm(
		widget.NewFormItem("ID", idEntry),
	)

	dlg := dialog.NewCustomConfirm("Удалить строку", "Удалить", "Отмена", form, func(ok bool) {
		if ok {
			idStr := strings.TrimSpace(idEntry.Text)
			id, err := strconv.Atoi(idStr)
			if err != nil {
				showError(window, "Неверный формат ID")
				return
			}

			err = deleteRowGeneric(ctx, pool, tableName, id)
			if err != nil {
				showError(window, "Ошибка удаления: "+err.Error())
				return
			}

			showInfo(window, "Строка успешно удалена!")
			loadTableByName(ctx, pool, tableName, dataPtr, table, infoLabel)
		}
	}, window)

	dlg.Show()
}

// Универсальное удаление строки
func deleteRowGeneric(ctx context.Context, pool *pgxpool.Pool, tableName string, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)
	_, err := pool.Exec(ctx, query, id)
	return err
}

// Диалог удаления таблицы
func showDeleteTableDialog(ctx context.Context, pool *pgxpool.Pool, window fyne.Window,
	currentTable *string, dataPtr *[][]string, table *widget.Table,
	infoLabel *widget.Label, tableSelect *widget.Select) {

	protectedTables := []string{"products", "categories", "orders", "order_items"}

	tableNameEntry := widget.NewEntry()
	tableNameEntry.SetText(*currentTable)

	warningLabel := widget.NewLabel("⚠ ВНИМАНИЕ: Удаление таблицы необратимо!\nВсе данные будут потеряны навсегда.")
	warningLabel.Wrapping = fyne.TextWrapWord

	confirmEntry := widget.NewEntry()
	confirmEntry.SetPlaceHolder(fmt.Sprintf("Введите '%s' для подтверждения", *currentTable))

	form := container.NewVBox(
		widget.NewLabel("Удаление таблицы"),
		widget.NewForm(
			widget.NewFormItem("Имя таблицы", tableNameEntry),
		),
		widget.NewSeparator(),
		warningLabel,
		widget.NewForm(
			widget.NewFormItem("Подтверждение", confirmEntry),
		),
	)

	dlg := dialog.NewCustomConfirm(
		"Удалить таблицу",
		"Удалить навсегда",
		"Отмена",
		form,
		func(ok bool) {
			if !ok {
				return
			}

			tableName := strings.TrimSpace(tableNameEntry.Text)
			confirmation := strings.TrimSpace(confirmEntry.Text)

			if tableName == "" {
				showError(window, "Имя таблицы не может быть пустым")
				return
			}

			if confirmation != tableName {
				showError(window, "Подтверждение не совпадает с именем таблицы")
				return
			}

			for _, protected := range protectedTables {
				if tableName == protected {
					showError(window, fmt.Sprintf("Таблица '%s' защищена от удаления", tableName))
					return
				}
			}

			err := dropTable(ctx, pool, tableName)
			if err != nil {
				showError(window, "Ошибка удаления таблицы: "+err.Error())
				return
			}

			showInfo(window, fmt.Sprintf("Таблица '%s' успешно удалена!", tableName))

			tablesList, err := getTablesListFromDB(ctx, pool)
			if err != nil {
				showError(window, "Ошибка обновления списка таблиц")
				return
			}

			tableSelect.Options = tablesList

			if len(tablesList) > 0 {
				*currentTable = tablesList[0]
				tableSelect.SetSelected(*currentTable)
				loadTableByName(ctx, pool, *currentTable, dataPtr, table, infoLabel)
			} else {
				*dataPtr = [][]string{{"Нет таблиц"}}
				table.Refresh()
				infoLabel.SetText("Таблиц не найдено")
			}
		},
		window,
	)

	dlg.Resize(fyne.NewSize(500, 350))
	dlg.Show()
}

// Функция удаления таблицы из БД
func dropTable(ctx context.Context, pool *pgxpool.Pool, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("не удалось удалить таблицу: %w", err)
	}
	return nil
}

// Вспомогательные функции для работы с БД

func getTablesListFromDB(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	query := `SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}
	return tables, nil
}

func getGenericTableData(ctx context.Context, pool *pgxpool.Pool, tableName string) ([][]string, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 1000", tableName)
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	var result [][]string

	headers := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		headers[i] = string(fd.Name)
	}
	result = append(result, headers)

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		rowData := make([]string, len(values))
		for i, v := range values {
			if v == nil {
				rowData[i] = "NULL"
			} else {
				rowData[i] = fmt.Sprintf("%v", v)
			}
		}
		result = append(result, rowData)
	}

	return result, nil
}

func loadTableByName(ctx context.Context, pool *pgxpool.Pool, tableName string,
	dataPtr *[][]string, table *widget.Table, infoLabel *widget.Label) {

	var newData [][]string
	var err error

	if tableName == "products" {
		newData, err = operation.GetAllProducts(ctx, pool)
	} else {
		newData, err = getGenericTableData(ctx, pool, tableName)
	}

	if err == nil {
		*dataPtr = newData

		// Автоматически настраиваем ширину колонок
		setOptimalColumnWidths(table, newData)

		table.Refresh()
		rowCount := len(newData) - 1
		if rowCount < 0 {
			rowCount = 0
		}
		infoLabel.SetText(fmt.Sprintf("Таблица: %s | Строк: %d", tableName, rowCount))
	}
}
