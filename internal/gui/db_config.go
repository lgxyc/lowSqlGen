package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/lowSqlGen/internal/model"
)

type DBConfigDialog struct {
	window   fyne.Window
	config   *model.DatabaseConfig
	onSubmit func(config *model.DatabaseConfig)
}

func NewDBConfigDialog(parent fyne.Window) *DBConfigDialog {
	dialog := &DBConfigDialog{
		window: fyne.CurrentApp().NewWindow("数据库连接配置"),
		config: &model.DatabaseConfig{},
	}

	// 创建输入框
	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("localhost")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("3306")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("root")

	passwordEntry := widget.NewPasswordEntry()

	// 创建表单
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "主机地址", Widget: hostEntry},
			{Text: "端口", Widget: portEntry},
			{Text: "用户名", Widget: usernameEntry},
			{Text: "密码", Widget: passwordEntry},
		},
		OnSubmit: func() {
			dialog.config.Host = hostEntry.Text
			dialog.config.Port = portEntry.Text
			dialog.config.Username = usernameEntry.Text
			dialog.config.Password = passwordEntry.Text

			if dialog.onSubmit != nil {
				dialog.onSubmit(dialog.config)
			}
			dialog.window.Close()
		},
		OnCancel: func() {
			dialog.window.Close()
		},
	}

	// 设置窗口内容
	dialog.window.SetContent(container.NewPadded(form))
	dialog.window.Resize(fyne.NewSize(300, 250))
	dialog.window.CenterOnScreen()

	return dialog
}

func (d *DBConfigDialog) Show() {
	d.window.Show()
}

func (d *DBConfigDialog) SetOnSubmit(callback func(config *model.DatabaseConfig)) {
	d.onSubmit = callback
}
