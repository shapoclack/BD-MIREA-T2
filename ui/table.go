package table

import (
	operation "BD_Mirea/internal"
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jackc/pgx/v5/pgxpool"
)

var TableData [][]string

func LimitEntry(ent *widget.Entry) {

}

// Создание таблицы с данными
func CreateTable(data [][]string) (*widget.Table, error) {
	if len(data) == 0 {
		return nil, errors.New("данные для таблицы пусты!")
	}

	TableData = data

	table := widget.NewTable(
		func() (int, int) {
			if len(TableData) == 0 {
				return 0, 0
			}
			return len(TableData), len(TableData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			if i.Row < len(TableData) && i.Col < len(TableData[i.Row]) {
				o.(*widget.Label).SetText(TableData[i.Row][i.Col])
			} else {
				o.(*widget.Label).SetText("")
			}
		})

	// Устанавливаем ширину колонок
	table.SetColumnWidth(0, 50)  // ID
	table.SetColumnWidth(1, 150) // Название
	table.SetColumnWidth(2, 200) // Описание
	table.SetColumnWidth(3, 100) // Цена
	table.SetColumnWidth(4, 80)  // Количество
	table.SetColumnWidth(5, 80)  // Активен
	table.SetColumnWidth(6, 120) // Категория

	return table, nil
}

// Обновление данных в таблице
func UpdateTableData(newData [][]string) {
	TableData = newData
}

// Получение текущих данных таблицы
func GetTableData() [][]string {
	return TableData
}
func showCopyDialog(ctx context.Context, pool *pgxpool.Pool, window fyne.Window, data [][]string, id widget.TableCellID, table *widget.Table) {
	// Создаем Entry widget для отображения данных
	entry := widget.NewMultiLineEntry()
	entry.SetText(data[id.Row][id.Col]) // ИСПРАВЛЕНО: Row и Col местами
	entry.Wrapping = fyne.TextWrapWord

	// Кнопки диалога
	copyButton := widget.NewButton("Копировать", func() {
		window.Clipboard().SetContent(data[id.Row][id.Col]) // ИСПРАВЛЕНО
		dialog.ShowInformation("Успех", "Данные скопированы в буфер обмена!", window)
	})

	// Кнопка для изменения
	updateTableButton := widget.NewButton("Изменить", func() {
		// Сначала обновляем данные в локальном массиве
		data[id.Row][id.Col] = entry.Text

		// ИСПРАВЛЕНО: Получаем данные из правильной строки
		idd, _ := strconv.Atoi(data[id.Row][0])             // ID из колонки 0
		name := data[id.Row][1]                             // Название из колонки 1
		description := data[id.Row][2]                      // Описание из колонки 2
		price, _ := strconv.ParseFloat(data[id.Row][3], 64) // Цена из колонки 3
		quantity, _ := strconv.Atoi(data[id.Row][4])        // Количество из колонки 4

		// ИСПРАВЛЕНО: Получаем название категории и находим её ID
		categoryName := data[id.Row][6] // Категория из колонки 6
		var categoryID *int

		if categoryName != "Без категории" {
			// Получаем список категорий для поиска ID
			categories, err := operation.GetCategories(ctx, pool)
			if err == nil {
				for _, cat := range categories {
					if cat[1] == categoryName { // cat[1] - название категории
						catID, _ := strconv.Atoi(cat[0]) // cat[0] - ID категории
						categoryID = &catID
						break
					}
				}
			}
		}

		// Обновляем в базе данных
		err := operation.UpdateProduct(ctx, pool, idd, name, description, price, quantity, categoryID)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		// Обновляем таблицу из базы данных
		refreshTable(ctx, pool, table, &data)
		dialog.ShowInformation("Успех", "Данные изменены!", window)
	})

	// Контейнер для кнопок
	buttons := container.NewVBox(copyButton, updateTableButton)

	// Контейнер для всего содержимого
	content := container.NewVBox(
		widget.NewLabel("Данные ячейки:"),
		entry,
		buttons,
	)

	// Создаем и показываем диалог
	dialog.ShowCustom("Копирование данных", "Закрыть", content, window)
}

func CreateUI(window fyne.Window, ctx context.Context, pool *pgxpool.Pool) {
	// Таблица для отображения данных
	data, _ := operation.GetAllProducts(ctx, pool)
	var table *widget.Table
	table = widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i.Row][i.Col])
			// Стилизация заголовков
			if i.Row == 0 {
				o.(*widget.Label).TextStyle = fyne.TextStyle{Bold: true}
			}
		},
	)
	//WARNING
	//Обработчик нажатия!
	table.OnSelected = func(id widget.TableCellID) {
		showCopyDialog(ctx, pool, window, data, id, table)
	}

	// Настройка размеров колонок
	table.SetColumnWidth(0, 60)  // ID
	table.SetColumnWidth(1, 150) // Название
	table.SetColumnWidth(2, 200) // Описание
	table.SetColumnWidth(3, 100) // Цена
	table.SetColumnWidth(4, 100) // Количество
	table.SetColumnWidth(5, 80)  // Активен
	table.SetColumnWidth(6, 120) // Категория

	// Форма добавления продукта
	nameEntry := widget.NewEntry()
	nameEntry.Validator = func(s string) error {
		if len(s) > 200 {
			return errors.New("Превышен лимит символов! (200 max)")
		}
		return nil
	}
	nameEntry.SetPlaceHolder("Введите название продукта")

	descEntry := widget.NewEntry()
	descEntry.Validator = func(s string) error {
		if len(s) > 200 {
			return errors.New("Превышен лимит символов! (200 max)")
		}
		return nil
	}
	descEntry.SetPlaceHolder("Введите описание продукта")
	descEntry.MultiLine = true

	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("0.00")

	quantityEntry := widget.NewEntry()
	quantityEntry.SetPlaceHolder("0")

	// Выпадающий список категорий
	categories, _ := operation.GetCategories(ctx, pool)
	categoryOptions := []string{}
	categoryMap := make(map[string]int)

	for _, cat := range categories {
		id, _ := strconv.Atoi(cat[0])
		categoryOptions = append(categoryOptions, cat[1])
		categoryMap[cat[1]] = id
	}

	categorySelect := widget.NewSelect(categoryOptions, nil)
	if len(categoryOptions) > 0 {
		categorySelect.SetSelected(categoryOptions[0])
	}

	// Кнопка добавления
	addButton := widget.NewButton("Добавить продукт", func() {
		name := strings.TrimSpace(nameEntry.Text)
		description := strings.TrimSpace(descEntry.Text)
		priceStr := strings.TrimSpace(priceEntry.Text)
		quantityStr := strings.TrimSpace(quantityEntry.Text)
		selectedCategory := categorySelect.Selected

		if name == "" || priceStr == "" || quantityStr == "" {
			showError(window, "Пожалуйста, заполните все обязательные поля")
			return
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			showError(window, "Неверный формат цены")
			return
		}

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			showError(window, "Неверный формат количества")
			return
		}

		var categoryID *int
		if selectedCategory != "" {
			if id, exists := categoryMap[selectedCategory]; exists {
				categoryID = &id
			}
		}

		err = operation.InsertProduct(ctx, pool, name, description, price, quantity, categoryID)
		if err != nil {
			showError(window, "Ошибка добавления: "+err.Error())
			return
		}

		// Очистка формы
		nameEntry.SetText("")
		descEntry.SetText("")
		priceEntry.SetText("")
		quantityEntry.SetText("")

		// Обновление таблицы
		refreshTable(ctx, pool, table, &data)

		showInfo(window, "Продукт успешно добавлен!")
	})

	// Форма удаления
	deleteEntry := widget.NewEntry()
	deleteEntry.SetPlaceHolder("ID продукта для удаления")

	deleteButton := widget.NewButton("Удалить продукт", func() {
		idStr := strings.TrimSpace(deleteEntry.Text)
		if idStr == "" {
			showError(window, "Введите ID продукта")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			showError(window, "Неверный формат ID")
			return
		}

		err = operation.DeleteProduct(ctx, pool, id)
		if err != nil {
			showError(window, "Ошибка удаления: "+err.Error())
			return
		}

		deleteEntry.SetText("")
		refreshTable(ctx, pool, table, &data)
		showInfo(window, "Продукт успешно удален!")
	})

	// Кнопка обновления данных
	refreshButton := widget.NewButton("Обновить данные", func() {
		refreshTable(ctx, pool, table, &data)
	})

	// Информация о категориях
	categoriesInfo := widget.NewRichTextFromMarkdown(`
	**Доступные категории:**
	• Электроника
	• Книги
	• Одежда
	• Продукты
	• Другое
	`)

	// Левая панель с формами
	leftPanel := container.NewVBox(
		widget.NewCard("Добавление нового продукта", "",
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("Название", nameEntry),
					widget.NewFormItem("Описание", descEntry),
					widget.NewFormItem("Цена", priceEntry),
					widget.NewFormItem("Количество", quantityEntry),
					widget.NewFormItem("Категория", categorySelect),
				),
				addButton,
			),
		),

		widget.NewSeparator(),

		widget.NewCard("Управление данными", "",
			container.NewVBox(
				refreshButton,
				widget.NewSeparator(),
				widget.NewForm(
					widget.NewFormItem("ID продукта", deleteEntry),
				),
				deleteButton,
			),
		),

		widget.NewSeparator(),
		categoriesInfo,
	)

	// Правая панель с таблицей
	rightPanel := container.NewBorder(
		widget.NewLabel("Список продуктов"),
		nil, nil, nil,
		container.NewScroll(table),
	)

	// Основной контейнер
	content := container.NewHSplit(leftPanel, rightPanel)
	content.SetOffset(0.35) // 35% для левой панели, 65% для таблицы

	window.SetContent(content)
}
func refreshTable(ctx context.Context, pool *pgxpool.Pool, table *widget.Table, data *[][]string) {
	newData, err := operation.GetAllProducts(ctx, pool)
	if err != nil {
		log.Printf("Ошибка обновления данных: %v", err)
		return
	}
	*data = newData
	table.Refresh()
}

// ИСПРАВЛЕННАЯ функция показа ошибок
func showError(window fyne.Window, message string) {
	// Создаем error из строки
	err := errors.New(message)
	errorDialog := dialog.NewError(err, window)
	errorDialog.Show()
}

// ИСПРАВЛЕННАЯ функция показа информации
func showInfo(window fyne.Window, message string) {
	infoDialog := dialog.NewInformation(
		"Информация",
		message,
		window,
	)
	infoDialog.Show()
}
