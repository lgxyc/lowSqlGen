package gui

import (
	"fmt"

	"fyne.io/fyne/v2"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/lowSqlGen/internal/service"
)

type JoinDialog struct {
	dialog      dialog.Dialog
	sourceTable string
	dbService   service.DatabaseService
	dbName      string
	onConfirm   func(targetTable, sourceColumn, targetColumn string)

	// 选中的列
	selectedSourceColumn string
	selectedTargetColumn string
	selectedTable        string
	sourceRadio          *widget.RadioGroup
	targetRadio          *widget.RadioGroup
}

func NewJoinDialog(window fyne.Window, sourceTable string, dbService service.DatabaseService, dbName string) *JoinDialog {
	j := &JoinDialog{
		sourceTable: sourceTable,
		dbService:   dbService,
		dbName:      dbName,
	}

	// 创建源表列表（左侧）
	sourceColumns, _ := dbService.GetColumns(dbName, sourceTable)
	var sourceListVar *widget.List
	sourceListVar = widget.NewList(
		func() int { return len(sourceColumns) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			check.Checked = false
			return container.NewHBox(check, widget.NewLabel("template"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			check := obj.(*fyne.Container).Objects[0].(*widget.Check)
			label := obj.(*fyne.Container).Objects[1].(*widget.Label)
			label.SetText(sourceColumns[id])

			// 设置选中状态
			check.Checked = (j.selectedSourceColumn == sourceColumns[id])
			check.OnChanged = func(checked bool) {
				if checked {
					// 取消其他选中状态
					for i := 0; i < sourceListVar.Length(); i++ {
						if i != id {
							itemContainer := sourceListVar.CreateItem()
							itemCheck := itemContainer.(*fyne.Container).Objects[0].(*widget.Check)
							itemCheck.Checked = false
						}
					}
					j.selectedSourceColumn = sourceColumns[id]
				} else {
					j.selectedSourceColumn = ""
				}
				sourceListVar.Refresh()
			}

		},
	)

	// 创建表列表（中间）
	tables, _ := dbService.GetTables(dbName)
	var targetColumns []string
	var targetListVar *widget.List
	targetListVar = widget.NewList(
		func() int { return len(targetColumns) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			check.Checked = false
			return container.NewHBox(check, widget.NewLabel("template"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			check := obj.(*fyne.Container).Objects[0].(*widget.Check)
			label := obj.(*fyne.Container).Objects[1].(*widget.Label)
			label.SetText(targetColumns[id])

			// 设置选中状态
			check.Checked = (j.selectedTargetColumn == targetColumns[id])
			check.OnChanged = func(checked bool) {
				if checked {
					// 取消其他选中状态
					for i := 0; i < targetListVar.Length(); i++ {
						if i != id {
							itemContainer := targetListVar.CreateItem()
							itemCheck := itemContainer.(*fyne.Container).Objects[0].(*widget.Check)
							itemCheck.Checked = false
						}
					}
					j.selectedTargetColumn = targetColumns[id]
				} else {
					j.selectedTargetColumn = ""
				}
				targetListVar.Refresh()
			}
		},
	)

	// 创建目标表列表（右侧）
	tablesList := widget.NewList(
		func() int { return len(tables) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(tables[id])
		},
	)

	// 表选择事件
	tablesList.OnSelected = func(id widget.ListItemID) {
		j.selectedTable = tables[id]
		targetColumns, _ = dbService.GetColumns(dbName, j.selectedTable)
		// 重置目标列选择
		j.selectedTargetColumn = ""
		targetListVar.Refresh()
	}

	// 修改确认按钮的检查逻辑
	confirmBtn := widget.NewButton("Confirm", func() {
		if j.selectedSourceColumn == "" {
			dialog.ShowError(fmt.Errorf("Please select a source column"), window)
			return
		}
		if j.selectedTable == "" {
			dialog.ShowError(fmt.Errorf("Please select a target table"), window)
			return
		}
		if j.selectedTargetColumn == "" {
			dialog.ShowError(fmt.Errorf("Please select a target column"), window)
			return
		}

		if j.onConfirm != nil {
			j.onConfirm(j.selectedTable, j.selectedSourceColumn, j.selectedTargetColumn)
		}
		j.dialog.Hide()
	})

	// 设置列表的最小大小
	sourceListVar.Resize(fyne.NewSize(200, 400))
	tablesList.Resize(fyne.NewSize(200, 400))
	targetListVar.Resize(fyne.NewSize(200, 400))

	// 创建滚动容器来包装列表
	sourceScroll := container.NewVScroll(sourceListVar)
	tablesScroll := container.NewVScroll(tablesList)
	targetScroll := container.NewVScroll(targetListVar)

	// 设置滚动容器的最小大小
	sourceScroll.SetMinSize(fyne.NewSize(200, 400))
	tablesScroll.SetMinSize(fyne.NewSize(200, 400))
	targetScroll.SetMinSize(fyne.NewSize(200, 400))

	// 修改对话框内容布局
	content := container.NewHBox(
		container.NewVBox(
			widget.NewLabel("Source Columns"),
			sourceScroll,
		),
		container.NewVBox(
			widget.NewLabel("Tables"),
			tablesScroll,
		),
		container.NewVBox(
			widget.NewLabel("Target Columns"),
			targetScroll,
		),
	)

	// 创建对话框，使用更大的尺寸
	j.dialog = dialog.NewCustom("Join Tables", "Cancel",
		container.NewVBox(content, confirmBtn), window)
	j.dialog.Resize(fyne.NewSize(700, 500))

	return j
}

func (j *JoinDialog) Show() {
	j.dialog.Show()
}

func (j *JoinDialog) SetOnConfirm(callback func(targetTable, sourceColumn, targetColumn string)) {
	j.onConfirm = callback
}
