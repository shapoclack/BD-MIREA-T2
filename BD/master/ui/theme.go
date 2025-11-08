package table

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme реализует кастомную темную тему на основе Material Design
type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

// Color возвращает цвета для различных элементов UI
func (ct CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Основные цвета из палитры
	primaryColor := color.NRGBA{R: 0xBB, G: 0x86, B: 0xFC, A: 0xFF}        // #BB86FC (светло-фиолетовый)
	primaryVariantColor := color.NRGBA{R: 0x37, G: 0x00, B: 0xB3, A: 0xFF} // #3700B3 (темно-фиолетовый)
	secondaryColor := color.NRGBA{R: 0x03, G: 0xDA, B: 0xC6, A: 0xFF}      // #03DAC6 (бирюзовый)

	// УЛУЧШЕННЫЕ цвета фона для более красочного вида
	backgroundColor := color.NRGBA{R: 0x1E, G: 0x1E, B: 0x2E, A: 0xFF} // #1E1E2E (темно-серый с фиолетовым)
	surfaceColor := color.NRGBA{R: 0x2A, G: 0x2A, B: 0x3E, A: 0xFF}    // #2A2A3E (светлее для панелей)

	errorColor := color.NRGBA{R: 0xCF, G: 0x66, B: 0x79, A: 0xFF}        // #CF6679 (розовый для ошибок)
	onBackgroundColor := color.NRGBA{R: 0xE8, G: 0xE8, B: 0xF0, A: 0xFF} // #E8E8F0 (мягкий белый для текста)

	switch name {
	// Основной цвет (используется для кнопок, акцентов)
	case theme.ColorNamePrimary:
		return primaryColor

	// Фон приложения
	case theme.ColorNameBackground:
		return backgroundColor

	// Фон для кнопок и виджетов
	case theme.ColorNameButton:
		return primaryVariantColor

	// Текст на фоне
	case theme.ColorNameForeground:
		return onBackgroundColor

	// Наведение курсора
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x37, G: 0x00, B: 0xB3, A: 0x40} // Полупрозрачный primary variant

	// Нажатие на элемент
	case theme.ColorNamePressed:
		return primaryVariantColor

	// Фокус (подсветка активного элемента)
	case theme.ColorNameFocus:
		return secondaryColor

	// Разделители - более видимые
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x3A, G: 0x3A, B: 0x4E, A: 0xFF}

	// Тени
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x66}

	// Поля ввода - с фиолетовым оттенком
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x25, G: 0x25, B: 0x38, A: 0xFF}

	// Граница полей ввода - фиолетовая
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0x4A, G: 0x4A, B: 0x6E, A: 0xFF}

	// Placeholder в полях ввода
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x99, G: 0x99, B: 0xAA, A: 0xFF}

	// Полоса прокрутки
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x55, G: 0x55, B: 0x6E, A: 0xFF}

	// Выделенный текст
	case theme.ColorNameSelection:
		return secondaryColor

	// Отключенные элементы
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x66, G: 0x66, B: 0x77, A: 0xFF}

	// Ошибки
	case theme.ColorNameError:
		return errorColor

	// Успех
	case theme.ColorNameSuccess:
		return secondaryColor

	// Предупреждения
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xFF, G: 0xBB, B: 0x33, A: 0xFF}

	// Гиперссылки
	case theme.ColorNameHyperlink:
		return primaryColor

	// Заголовки меню
	case theme.ColorNameMenuBackground:
		return surfaceColor

	// Оверлей (затемнение фона при диалогах)
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xCC}

	default:
		// Для всех остальных цветов используем темную тему по умолчанию
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font возвращает шрифты для различных стилей текста
func (ct CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon возвращает иконки
func (ct CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size возвращает размеры для различных элементов UI
func (ct CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
