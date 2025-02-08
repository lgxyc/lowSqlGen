package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/lowSqlGen/internal/gui"
)

func main() {
	// 创建应用实例
	application := app.New()
	
	// 创建主窗口
	mainWindow := application.NewWindow("LowSqlGen - SQL可视化生成器")
	mainWindow.Resize(fyne.NewSize(1024, 768))
	
	// 初始化GUI
	gui.InitMainWindow(mainWindow)
	
	// 显示并运行
	mainWindow.ShowAndRun()
} 