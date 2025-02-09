package main

import (
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/flopp/go-findfont"
	"github.com/lowSqlGen/internal/gui"
)

func main() {
	// 设置中文字体
	setChineseFont()

	// 创建应用实例
	application := app.New()

	// 设置自定义主题
	application.Settings().SetTheme(gui.NewMyTheme())

	// 创建主窗口
	mainWindow := application.NewWindow("LowSqlGen - SQL可视化生成器")
	mainWindow.Resize(fyne.NewSize(1024, 768))

	// 初始化GUI
	gui.InitMainWindow(mainWindow)

	// 显示并运行
	mainWindow.ShowAndRun()
}

func setChineseFont() {
	// 寻找中文字体
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		// 优先使用这些字体
		if strings.Contains(path, "simkai.ttf") || // 楷体
			strings.Contains(path, "simhei.ttf") || // 黑体
			strings.Contains(path, "simsun.ttc") || // 宋体
			strings.Contains(path, "msyh.ttf") || // 微软雅黑
			strings.Contains(path, "msyh.ttc") { // 微软雅黑
			os.Setenv("FYNE_FONT", path) // 设置字体
			return
		}
	}
}
