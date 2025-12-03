package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"wcediter/wcsave"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/ini.v1"
)

var (
	// 全局状态
	fyneApp         fyne.App
	editor          *wcsave.SaveEditor
	currentSave     string
	currentWindow   fyne.Window
	characterWindow fyne.Window // 角色属性编辑窗口
	// 保存每个角色的属性输入框
	characterPropertyInputs map[int][]*propertyInput

	// 存档进度相关
	progressNames = []string{"进度一", "进度二", "进度三", "进度四", "进度五"}

	// 配置文件相关
	configFile  = "./wcediter.ini"
	fileRecords []FileRecordItem
	// 固定的默认记录
	defaultRecords = map[string]string{
		"原版":    "./Save0.dat",
		"无名原版":  "./Sald0.dat",
		"无名简单版": "./Save0.dat",
		"无名困难版": "./Sav00.dat",
	}
	// 保存默认记录的顺序
	defaultRecordOrder = []string{"原版", "无名原版", "无名简单版", "无名困难版"}
)

// FileRecordItem 表示选择记录项
type FileRecordItem struct {
	Tag       string
	Path      string
	IsDefault bool
}

// 属性输入框结构体
type propertyInput struct {
	label    *widget.Label
	input    *widget.Entry
	property string
}

// 自定义文件过滤器，只显示以0.dat结尾的文件
type zeroDatFileFilter struct{}

// 实现FileFilter接口的Matches方法
func (f *zeroDatFileFilter) Matches(uri fyne.URI) bool {
	// 获取文件名
	fileName := filepath.Base(uri.Path())
	// 检查文件名是否以0.dat结尾
	return strings.HasSuffix(fileName, "0.dat")
}

// 进度选择回调函数
type progressSelectCallback func(int)

// 文件选择回调函数类型
type fileSelectCallback func(string, string)

// 选择存档文件的独立方法 - 使用新窗口展示
func selectSaveFile(parentWindow fyne.Window, onSelect fileSelectCallback) {
	// 加载选择记录
	loadFileRecords()

	// 创建一个新的窗口
	fileWindow := fyneApp.NewWindow("选择存档文件")
	fileWindow.Resize(fyne.NewSize(800, 600))
	fileWindow.SetFixedSize(false)

	// 设置窗口关闭时的行为
	fileWindow.SetOnClosed(func() {
		// 保存选择记录
		saveFileRecords()
		log.Println("文件选择窗口已关闭，配置已保存")
		// 重新显示主窗口
		currentWindow.Show()
	})

	// 创建文件列表容器
	fileList := widget.NewFileIcon(nil)

	// 创建选中文件信息
	selectedPath := ""
	selectedTag := ""
	pathLabel := widget.NewLabel("选中的文件: ")
	pathLabel.Wrapping = fyne.TextWrapWord
	pathLabel.Truncation = fyne.TextTruncateOff
	tagLabel := widget.NewLabel("文件标签: ")
	tagLabel.Wrapping = fyne.TextWrapWord
	tagLabel.Truncation = fyne.TextTruncateOff

	// 创建确认按钮
	confirmButton := widget.NewButton("确认选择", func() {
		if selectedPath != "" {
			// 保存选择记录
			saveFileRecords()
			log.Println("确认选择，配置已保存")
			// 调用回调函数，传递选中的标签和文件路径
			if onSelect != nil {
				onSelect(selectedTag, selectedPath)
			}
			// 关闭文件选择窗口
			fileWindow.Close()
		}
	})
	confirmButton.Disable()

	// 声明recordList变量
	var recordList *widget.List

	// 创建文件选择区域
	filePicker := widget.NewButton("浏览文件系统", func() {
		// 创建文件选择对话框
		fileDialog := dialog.NewFileOpen(
			func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, fileWindow)
					return
				}
				if reader != nil {
					filePath := reader.URI().Path()
					// 检查文件是否以0.dat结尾
					if filepath.Ext(filePath) == ".dat" && strings.HasSuffix(filePath[:len(filePath)-4], "0") {
						log.Printf("用户选择了存档文件：%s", filePath)
						// 更新文件图标显示
						fileList.SetURI(reader.URI())

						// 生成默认标签（文件名）
						tag := getFileName(filePath)
						baseTag := tag

						// 检查是否已存在相同路径的记录
						found := false
						recordIndex := -1
						for i, item := range fileRecords {
							if item.Path == filePath {
								found = true
								recordIndex = i
								// 更新选中信息
								selectedTag = item.Tag
								selectedPath = item.Path
								break
							}
						}

						if !found {
							// 确保标签唯一
							uniqueTag := tag
							localCounter := 1

							// 检查标签是否已被使用，如果是则添加数字后缀
							for {
								isUnique := true
								for _, item := range fileRecords {
									if item.Tag == uniqueTag {
										isUnique = false
										break
									}
								}

								if isUnique {
									break
								}

								// 生成新的标签
								uniqueTag = fmt.Sprintf("%s(%d)", baseTag, localCounter)
								localCounter++
							}

							// 添加新记录
							newItem := FileRecordItem{
								Tag:       uniqueTag,
								Path:      filePath,
								IsDefault: false,
							}
							fileRecords = append(fileRecords, newItem)
							// 更新选中信息
							selectedTag = uniqueTag
							selectedPath = filePath
							// 刷新记录列表，使新记录显示在左侧
							if recordList != nil {
								recordList.Refresh()
								// 选中新添加的记录
								recordList.Select(len(fileRecords) - 1)
							}
						} else {
							// 选中已存在的记录
							if recordList != nil && recordIndex != -1 {
								recordList.Select(recordIndex)
							}
						}

						// 更新右侧显示
						tagLabel.SetText("文件标签: " + selectedTag)
						pathLabel.SetText("选中的文件: " + selectedPath)
						confirmButton.Enable()
					} else {
						dialog.ShowInformation("提示", "请选择以0.dat结尾的存档文件", fileWindow)
					}
					reader.Close()
				}
			},
			fileWindow,
		)

		// 设置默认目录为程序当前工作目录
		currentDir, err := os.Getwd()
		if err != nil {
			log.Printf("获取当前工作目录失败: %v", err)
		} else {
			log.Printf("当前工作目录: %s", currentDir)
			// 使用storage包创建文件URI
			uri := storage.NewFileURI(currentDir)
			// 获取可列出内容的URI（目录）
			if lister, err := storage.ListerForURI(uri); err == nil {
				// 设置文件选择器的默认位置
				fileDialog.SetLocation(lister)
			}
		}
		// 创建自定义文件过滤器实例
		fileFilter := &zeroDatFileFilter{}
		// 设置文件过滤器
		fileDialog.SetFilter(fileFilter)
		// 设置文件对话框的大小与父窗口相同
		fileDialog.Resize(fileWindow.Canvas().Size())
		// 设置文件对话框默认以列表形式展示
		fileDialog.SetView(dialog.ListView)
		fileDialog.Show()
	})

	// 创建记录列表
	recordList = widget.NewList(
		func() int {
			return len(fileRecords)
		},
		func() fyne.CanvasObject {
			// 创建记录项的UI组件
			tagEntry := widget.NewEntry()
			pathLabel := widget.NewLabel("")
			pathLabel.Wrapping = fyne.TextWrapOff
			pathLabel.Truncation = fyne.TextTruncateOff
			deleteBtn := widget.NewButton("删除", func() {})

			// 创建Grid布局
			grid := container.NewGridWithColumns(4,
				container.NewStack(tagEntry),  // 标签列
				container.NewStack(pathLabel), // 路径列
				layout.NewSpacer(),            // 占位符
				deleteBtn,                     // 删除按钮
			)

			// 创建一个带有边框的容器，用于区分不同类型的记录
			return container.NewBorder(nil, nil, nil, nil, grid)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			// 获取记录项
			item := fileRecords[i]

			// 获取Border容器
			borderContainer := o.(*fyne.Container)
			// 获取Grid容器（只有中间组件，所以是Objects[0]）
			grid := borderContainer.Objects[0].(*fyne.Container)
			// 获取标签输入框（嵌套在container.NewStack中）
			tagContainer := grid.Objects[0].(*fyne.Container)
			tagEntry := tagContainer.Objects[0].(*widget.Entry)
			// 获取路径标签（嵌套在container.NewStack中）
			pathContainer := grid.Objects[1].(*fyne.Container)
			pathLabel := pathContainer.Objects[0].(*widget.Label)
			// 获取删除按钮
			deleteBtn := grid.Objects[3].(*widget.Button)

			// 设置标签
			tagEntry.SetText(item.Tag)

			// 设置路径（转换为相对路径）
			relPath := getRelativePath(item.Path)
			pathLabel.SetText(relPath)

			// 根据是否为默认记录设置样式
			if item.IsDefault {
				// 默认记录，正常样式显示，不可编辑
				tagEntry.Disable()
				tagEntry.TextStyle = fyne.TextStyle{}
				pathLabel.TextStyle = fyne.TextStyle{}
				deleteBtn.Disable()
			} else {
				// 普通记录，正常显示，可编辑
				tagEntry.Enable()
				tagEntry.TextStyle = fyne.TextStyle{}
				pathLabel.TextStyle = fyne.TextStyle{}
				deleteBtn.Enable()
			}

			// 设置标签编辑事件
			tagEntry.OnChanged = func(s string) {
				// 更新记录标签
				fileRecords[i].Tag = s
			}

			// 保存当前记录的路径，用于后续删除操作
			currentPath := item.Path

			// 设置删除按钮事件
			deleteBtn.OnTapped = func() {
				// 动态获取当前标签值
				var currentTag string
				for _, record := range fileRecords {
					if record.Path == currentPath {
						currentTag = record.Tag
						break
					}
				}

				// 创建确认对话框
				confirmDialog := dialog.NewConfirm(
					"确认删除",
					fmt.Sprintf("确定要删除记录 '%s' 吗？", currentTag),
					func(confirmed bool) {
						if confirmed {
							// 查找要删除的记录索引
							for idx, record := range fileRecords {
								if record.Path == currentPath && !record.IsDefault {
									// 删除记录
									fileRecords = append(fileRecords[:idx], fileRecords[idx+1:]...)
									// 刷新记录列表
									recordList.Refresh()
									break
								}
							}
						}
					},
					fileWindow,
				)
				confirmDialog.Show()
			}
		},
	)

	// 设置记录列表的选中事件
	recordList.OnSelected = func(id widget.ListItemID) {
		// 选中该记录
		selectedTag = fileRecords[id].Tag
		selectedPath = fileRecords[id].Path
		// 更新右侧显示
		tagLabel.SetText("文件标签: " + selectedTag)
		pathLabel.SetText("选中的文件: " + selectedPath)
		confirmButton.Enable()
	}

	// 设置记录列表的取消选中事件
	recordList.OnUnselected = func(id widget.ListItemID) {
		// 取消选中时清空显示
		tagLabel.SetText("文件标签: ")
		pathLabel.SetText("选中的文件: ")
		confirmButton.Disable()
	}

	// 创建右侧面板
	rightPanel := container.NewVBox(
		widget.NewLabelWithStyle("请选择风云存档文件（以0.dat结尾）", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		container.NewCenter(fileList),
		tagLabel,
		pathLabel,
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			filePicker,
			confirmButton,
			widget.NewButton("取消", func() {
				// 保存选择记录
				saveFileRecords()
				log.Println("取消选择，配置已保存")
				// 执行onSelect回调函数，传递空的路径
				if onSelect != nil {
					onSelect("", "")
				}
				// 关闭窗口
				fileWindow.Close()
			}),
			layout.NewSpacer(),
		),
	)

	// 创建左侧面板，确保占据整个左边窗口
	leftPanel := container.NewBorder(
		widget.NewLabelWithStyle("选择记录", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil,
		nil,
		nil,
		recordList,
	)

	// 创建主布局（左侧记录列表，右侧文件选择）
	mainLayout := container.NewHSplit(
		leftPanel,
		rightPanel,
	)
	mainLayout.SetOffset(0.5) // 设置左侧宽度为50%

	// 设置初始窗口内容
	fileWindow.SetContent(mainLayout)

	// 显示窗口
	fileWindow.CenterOnScreen()
	fileWindow.Show()
}

// 创建进度选择界面
func createProgressSelectUI(onSelect progressSelectCallback) *fyne.Container {
	log.Println("创建进度选择界面，包含文件选择功能")
	// 创建文件选择相关组件
	fileLabel := widget.NewLabel("请选择存档文件：")

	// 加载选择记录
	loadFileRecords()

	// 准备选择记录的标签和路径映射
	tagPathMap := make(map[string]string)
	tags := []string{}

	// 如果没有记录，使用默认记录
	if len(fileRecords) == 0 {
		log.Println("没有配置文件，使用默认记录")
		// 使用默认记录
		for _, tag := range defaultRecordOrder {
			path := defaultRecords[tag]
			tagPathMap[tag] = path
			tags = append(tags, tag)
		}
	} else {
		// 使用配置文件中的记录
		for _, item := range fileRecords {
			tagPathMap[item.Tag] = item.Path
			tags = append(tags, item.Tag)
		}
	}

	// 创建选择框
	var selectedTag string
	if len(tags) > 0 {
		selectedTag = tags[0]
	}
	fileSelect := widget.NewSelect(tags, func(tag string) {
		selectedTag = tag
	})
	if len(tags) > 0 {
		fileSelect.SetSelected(selectedTag)
	}

	// 创建浏览按钮，用于添加新记录
	browseButton := widget.NewButton("浏览...", func() {
		// 隐藏主窗口
		currentWindow.Hide()
		selectSaveFile(currentWindow, func(selectedTagFromDialog, filePath string) {
			// 重新加载记录
			loadFileRecords()
			// 重新准备选择记录的标签和路径映射
			tagPathMap = make(map[string]string)
			tags = []string{}
			for _, item := range fileRecords {
				tagPathMap[item.Tag] = item.Path
				tags = append(tags, item.Tag)
			}
			// 更新选择框
			fileSelect.Options = tags
			// 只有当selectedTagFromDialog不为空时，才更新选中的项目
			if selectedTagFromDialog != "" {
				// 选中用户刚刚选择的文件
				selectedTag = selectedTagFromDialog
				fileSelect.SetSelected(selectedTag)
			}
			fileSelect.Refresh()
		})
	})

	// 创建标题
	title := widget.NewLabel("请选择欲修改的进度名：")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	// 创建进度名称标签
	progressTitle := widget.NewLabel("进度名（1-5）")
	progressTitle.Alignment = fyne.TextAlignLeading // 左对齐

	// 创建单选按钮组
	radioGroup := widget.NewRadioGroup(progressNames, func(value string) {
		// 这个回调在单选按钮变化时触发，但我们只在点击确定按钮时处理
	})
	radioGroup.SetSelected(progressNames[0]) // 默认选择第一个进度

	// 创建确定按钮
	confirmButton := widget.NewButton("确定", func() {
		// 检查是否选择了文件
		if selectedTag == "" {
			dialog.ShowInformation("提示", "请先选择一个存档文件", currentWindow)
			return
		}

		// 获取选中的文件路径
		filePath := tagPathMap[selectedTag]
		if filePath == "" {
			dialog.ShowInformation("提示", "请先选择一个存档文件", currentWindow)
			return
		}

		selected := radioGroup.Selected
		if selected != "" {
			// 找到选中的进度索引
			for i, name := range progressNames {
				if name == selected {
					// 保存选择的文件路径
					currentSave = filePath
					onSelect(i)
					return
				}
			}
		}
	})

	// 创建取消按钮
	cancelButton := widget.NewButton("取消", func() {
		// 取消时退出窗口
		log.Println("用户取消了进度选择，正在退出应用")
		fyneApp.Quit() // 退出应用程序
	})

	// 创建按钮容器
	buttonBox := container.NewHBox(
		layout.NewSpacer(),
		confirmButton,
		cancelButton,
		layout.NewSpacer(),
	)

	// 创建单选按钮容器 - 居中显示
	radioCenterContainer := container.NewHBox(
		layout.NewSpacer(), // 左侧占位
		container.NewVBox(
			progressTitle,
			radioGroup,
		),
		layout.NewSpacer(), // 右侧占位
	)

	// 创建作者信息标签
	authorLabel := widget.NewLabel("作者: switch.st@gmail.com")
	authorLabel.Alignment = fyne.TextAlignCenter
	authorLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 创建文件选择区域的容器
	fileSelectionContainer := container.NewBorder(
		nil, nil, fileLabel, browseButton,
		fileSelect,
	)

	// 整个内容容器
	content := container.NewVBox(
		fileSelectionContainer,
		title,
		layout.NewSpacer(),
		radioCenterContainer,
		layout.NewSpacer(),
		buttonBox,
		authorLabel,
	)

	return content
}

// 创建属性输入框
func createPropertyInput(property, labelText string, initialValue string) *propertyInput {
	label := widget.NewLabel(labelText)
	input := widget.NewEntry()
	input.SetText(initialValue)
	input.SetPlaceHolder("请输入数字值")
	// 增加输入框宽度使其更长
	input.Resize(fyne.NewSize(200, 30))

	return &propertyInput{
		label:    label,
		input:    input,
		property: property,
	}
}

// 加载进度文件并打开角色属性窗口
func loadProgressAndOpenCharacterUI(progressIndex int, progressWindow fyne.Window) {
	if progressIndex < 0 || progressIndex >= len(progressNames) {
		log.Printf("无效的进度索引: %d", progressIndex)
		dialog.ShowError(fmt.Errorf("无效的进度索引"), progressWindow)
		return
	}

	// 如果没有选择基础存档文件，显示错误
	if currentSave == "" {
		log.Printf("错误: 没有选择存档文件")
		dialog.ShowError(fmt.Errorf("请先选择存档文件"), progressWindow)
		return
	}

	// 根据用户选择的基础文件路径和进度索引生成对应的存档文件路径
	// 例如: 如果基础文件是 "xxx0.dat"，进度1对应 "xxx1.dat"
	ext := filepath.Ext(currentSave)
	baseName := currentSave[:len(currentSave)-len(ext)]
	// 替换末尾的0为对应的进度索引+1
	baseName = baseName[:len(baseName)-1] + strconv.Itoa(progressIndex+1)
	filePath := baseName + ext

	log.Printf("准备加载进度 %d 的文件: %s", progressIndex+1, filePath)

	// 加载找到的存档文件
	err := loadSaveFile(filePath)
	if err != nil {
		log.Printf("加载存档失败: %v", err)
		dialog.ShowError(fmt.Errorf("加载存档失败: %v", err), progressWindow)
	} else {
		log.Printf("成功加载进度: %s, 文件: %s", progressNames[progressIndex], filePath)

		// 隐藏进度选择窗口
		progressWindow.Hide()

		// 创建角色属性窗口，传递进度索引
		openCharacterWindow(progressIndex)
	}
}

// 打开角色属性窗口
func openCharacterWindow(progressIndex int) {
	// 创建新的角色属性窗口
	log.Println("创建角色属性窗口...")
	// 在标题后添加进度信息
	title := fmt.Sprintf("角色属性编辑 - %s", progressNames[progressIndex])
	characterWindow = fyneApp.NewWindow(title)

	// 设置更大的窗口大小以确保所有角色属性都能完整显示
	characterWindow.Resize(fyne.NewSize(650, 800))

	// 设置窗口关闭时的行为
	characterWindow.SetCloseIntercept(func() {
		log.Println("角色属性窗口关闭中...")
		// 重新显示进度选择窗口
		log.Println("重新显示进度选择窗口...")
		currentWindow.Show()
		// 关闭窗口
		characterWindow.Close()
	})

	// 创建角色属性UI内容
	log.Println("创建角色属性UI内容...")
	content := createMainUI()
	characterWindow.SetContent(content)

	// 显示角色属性窗口
	log.Println("显示角色属性窗口...")
	characterWindow.Canvas().Focus(nil)
	characterWindow.CenterOnScreen() // 居中显示窗口
	characterWindow.Show()
}

// 加载存档文件
func loadSaveFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("错误: 存档文件不存在: %s", filePath)
		return fmt.Errorf("存档文件不存在: %s\n请确认文件路径是否正确", filePath)
	}

	// 初始化编辑器
	editor = wcsave.NewSaveEditor()

	// 读取存档
	err := editor.ReadSave(filePath)
	if err != nil {
		return fmt.Errorf("读取存档失败: %v", err)
	}

	currentSave = filePath
	log.Printf("成功加载存档: %s", filePath)
	log.Printf("发现 %d 个角色", editor.GetCharacterCount())

	return nil
}

// 创建角色选择下拉框
func createCharacterTabs(propertyInputs []*propertyInput) *container.AppTabs {
	// 初始化角色属性输入框映射
	if characterPropertyInputs == nil {
		characterPropertyInputs = make(map[int][]*propertyInput)
	}
	// 创建标签页容器，使用底部标签样式以便更好地显示角色信息
	tabs := container.NewAppTabs()
	// 设置标签页位置在顶部，这是更常见的标签页布局
	tabs.SetTabLocation(container.TabLocationTop)

	if editor != nil {
		for i := 0; i < editor.GetCharacterCount(); i++ {
			if char, ok := editor.GetCharacterByIndex(i); ok {
				// 为每个角色创建新的属性输入框副本
				charPropertyInputs := make([]*propertyInput, len(propertyInputs))
				for j, input := range propertyInputs {
					charPropertyInputs[j] = createPropertyInput(input.property, input.label.Text, "0")
					// 为输入框设置固定宽度，确保在不同平台上的一致性
					charPropertyInputs[j].input.Wrapping = fyne.TextWrapOff
					charPropertyInputs[j].input.Resize(fyne.NewSize(150, 30))
				}

				// 创建角色属性的网格布局
				inputGrid := container.New(layout.NewGridLayout(2))
				for _, input := range charPropertyInputs {
					inputGrid.Add(input.label)
					inputGrid.Add(input.input)
				}

				// 更新角色数据到输入框
				updateCharacterUI(i, charPropertyInputs)

				// 创建角色标签页，移除保存按钮，简化内容结构
				tabContent := container.NewPadded(container.NewVBox(inputGrid))

				// 保存角色属性输入框到全局映射
				characterPropertyInputs[i] = charPropertyInputs

				tabs.Append(container.NewTabItem(fmt.Sprintf("%d. %s", i+1, char.Name), tabContent))
			}
		}
	}

	return tabs
}

// 更新角色数据界面
func updateCharacterUI(charIndex int, propertyInputs []*propertyInput) {
	if editor == nil {
		return
	}

	char, ok := editor.GetCharacterByIndex(charIndex)
	if !ok {
		return
	}

	// 更新属性输入框的值
	for _, input := range propertyInputs {
		switch input.property {
		case "CurrentExp":
			input.input.SetText(strconv.FormatInt(int64(char.Data.CurrentExp), 10))
		case "NextLevelExp":
			input.input.SetText(strconv.FormatInt(int64(char.Data.NextLevelExp), 10))
		case "CurrentHP":
			input.input.SetText(strconv.FormatInt(int64(char.Data.CurrentHP), 10))
		case "CurrentMP":
			input.input.SetText(strconv.FormatInt(int64(char.Data.CurrentMP), 10))
		case "MaxHP":
			input.input.SetText(strconv.FormatInt(int64(char.Data.MaxHP), 10))
		case "MaxMP":
			input.input.SetText(strconv.FormatInt(int64(char.Data.MaxMP), 10))
		case "Strength":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Strength), 10))
		case "Reaction":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Reaction), 10))
		case "Constitution":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Constitution), 10))
		case "Speed":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Speed), 10))
		case "Attack":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Attack), 10))
		case "Defense":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Defense), 10))
		case "Luck":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Luck), 10))
		case "Level":
			input.input.SetText(strconv.FormatInt(int64(char.Data.Level), 10))
		}
	}
}

// 保存角色数据更改
func saveCharacterChanges(charIndex int, propertyInputs []*propertyInput) error {
	if editor == nil {
		return fmt.Errorf("编辑器未初始化")
	}

	char, ok := editor.GetCharacterByIndex(charIndex)
	if !ok {
		return fmt.Errorf("角色不存在")
	}

	// 更新角色数据
	for _, input := range propertyInputs {
		valueStr := input.input.Text
		if valueStr == "" {
			continue
		}

		// 根据属性类型转换值
		switch input.property {
		case "CurrentExp":
			val, err := strconv.ParseInt(valueStr, 10, 32)
			if err != nil {
				return fmt.Errorf("当前经验值格式错误: %v", err)
			}
			char.Data.CurrentExp = int32(val)
		case "NextLevelExp":
			val, err := strconv.ParseInt(valueStr, 10, 32)
			if err != nil {
				return fmt.Errorf("升级经验值格式错误: %v", err)
			}
			char.Data.NextLevelExp = int32(val)
		case "CurrentHP":
			val, err := strconv.ParseInt(valueStr, 10, 32)
			if err != nil {
				return fmt.Errorf("当前生命值格式错误: %v", err)
			}
			char.Data.CurrentHP = int32(val)
		case "CurrentMP":
			val, err := strconv.ParseInt(valueStr, 10, 32)
			if err != nil {
				return fmt.Errorf("当前内力值格式错误: %v", err)
			}
			char.Data.CurrentMP = int32(val)
		case "MaxHP":
			val, err := strconv.ParseInt(valueStr, 10, 32)
			if err != nil {
				return fmt.Errorf("最大生命值格式错误: %v", err)
			}
			char.Data.MaxHP = int32(val)
		case "MaxMP":
			val, err := strconv.ParseInt(valueStr, 10, 32)
			if err != nil {
				return fmt.Errorf("最大内力值格式错误: %v", err)
			}
			char.Data.MaxMP = int32(val)
		case "Strength":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("力量格式错误: %v", err)
			}
			char.Data.Strength = int16(val)
		case "Reaction":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("反应格式错误: %v", err)
			}
			char.Data.Reaction = int16(val)
		case "Constitution":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("体质格式错误: %v", err)
			}
			char.Data.Constitution = int16(val)
		case "Speed":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("速度格式错误: %v", err)
			}
			char.Data.Speed = int16(val)
		case "Attack":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("攻击格式错误: %v", err)
			}
			char.Data.Attack = int16(val)
		case "Defense":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("防御格式错误: %v", err)
			}
			char.Data.Defense = int16(val)
		case "Luck":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("运气格式错误: %v", err)
			}
			char.Data.Luck = int16(val)
		case "Level":
			val, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				return fmt.Errorf("等级格式错误: %v", err)
			}
			char.Data.Level = int16(val)
		}
	}

	// 更新编辑器中的角色数据
	result := editor.UpdateCharacter(charIndex, char.Data)
	if !result {
		return fmt.Errorf("更新角色数据失败")
	}

	return nil
}

// 创建主界面
func createMainUI() *fyne.Container {
	// 初始化默认值
	moneyValue := "0"

	// 如果editor为nil，创建一个空的编辑器实例
	if editor == nil {
		editor = wcsave.NewSaveEditor()
	}

	// 创建属性输入框模板
	propertyInputs := []*propertyInput{
		createPropertyInput("CurrentExp", "当前经验值:", "0"),
		createPropertyInput("NextLevelExp", "升级经验值:", "0"),
		createPropertyInput("CurrentHP", "当前生命值:", "0"),
		createPropertyInput("MaxHP", "最大生命值:", "0"),
		createPropertyInput("CurrentMP", "当前内力值:", "0"),
		createPropertyInput("MaxMP", "最大内力值:", "0"),
		createPropertyInput("Strength", "力量:", "0"),
		createPropertyInput("Reaction", "反应:", "0"),
		createPropertyInput("Constitution", "体质:", "0"),
		createPropertyInput("Speed", "速度:", "0"),
		createPropertyInput("Attack", "攻击:", "0"),
		createPropertyInput("Defense", "防御:", "0"),
		createPropertyInput("Luck", "运气:", "0"),
		createPropertyInput("Level", "等级:", "0"),
	}

	// 创建角色标签页
	characterTabs := createCharacterTabs(propertyInputs)

	// 获取银两值
	if editor != nil {
		moneyValue = strconv.FormatInt(int64(editor.MoneyInfo.Value), 10)
	}

	// 创建银两相关的输入框和按钮
	moneyLabel := widget.NewLabel("银两数量:")
	moneyInput := widget.NewEntry()
	moneyInput.SetText(moneyValue)
	moneyInput.SetPlaceHolder("请输入银两数量")

	// 创建保存修改按钮
	saveFileButton := widget.NewButton("保存修改", func() {
		if currentSave == "" || editor == nil {
			dialog.ShowError(fmt.Errorf("没有加载的存档文件"), characterWindow)
			return
		}

		// 修改银两
		if !updateMoneyValue(moneyInput.Text) {
			return
		}

		// 保存所有角色的数据
		savedCount := 0
		for charIndex, inputs := range characterPropertyInputs {
			err := saveCharacterChanges(charIndex, inputs)
			if err != nil {
				log.Printf("保存角色%d数据失败: %v", charIndex+1, err)
				continue
			}
			savedCount++
		}
		log.Printf("成功保存%d个角色的数据", savedCount)

		// 保存更改（直接修改源文件）
		var err error
		err = editor.SaveChanges(currentSave, currentSave)
		if err != nil {
			dialog.ShowError(fmt.Errorf("保存文件失败: %v", err), characterWindow)
			return
		}

		dialog.ShowInformation("成功", "保存修改成功！", characterWindow)
	})

	// 创建取消按钮
	cancelButton := widget.NewButton("取消", func() {
		log.Println("用户点击了取消按钮")
		// 直接显示主窗口
		currentWindow.Show()
		// 关闭角色属性窗口，不保存任何修改
		characterWindow.Close()
	})

	// 创建银两容器
	moneyContainer := container.NewGridWithColumns(2, moneyLabel, moneyInput)

	// 创建按钮容器，包含保存和取消按钮
	buttonContainer := container.NewHBox(
		layout.NewSpacer(),
		saveFileButton,
		cancelButton,
		layout.NewSpacer(),
	)

	// 创建主容器
	mainContainer := container.NewVBox(
		// 隐藏标题和状态信息
		widget.NewSeparator(),
		widget.NewLabel("角色属性管理:"),
		// 使用角色标签页替代选择器和属性网格
		characterTabs,
		widget.NewSeparator(),
		moneyContainer,
		layout.NewSpacer(),
		// 添加按钮容器，包含保存和取消按钮
		buttonContainer,
	)

	// 标签页已在createCharacterTabs中初始化了所有角色的数据

	return mainContainer
}

func main() {
	// 解析命令行参数（暂时保留flag包以支持未来扩展）
	flag.Parse()

	// 打印环境信息用于调试
	log.Println("风云存档编辑器启动中...")
	log.Printf("操作系统: macOS")
	log.Printf("当前工作目录: %s", getCurrentDir())

	// 已移除未使用的characterStats初始化

	// 检查DISPLAY环境变量（对macOS X11很重要）
	display := os.Getenv("DISPLAY")
	if display == "" {
		log.Println("警告: DISPLAY环境变量未设置，可能会影响GUI显示")
		log.Println("如果使用X11，请确保已安装XQuartz并正确配置")
	} else {
		log.Printf("DISPLAY环境变量: %s", display)
	}

	// 创建Fyne应用，使用NewWithID提供唯一标识符以避免Preferences API错误
	log.Println("正在创建Fyne应用实例...")
	// 为macOS添加显示配置选项
	fyneApp = app.NewWithID("wang.switch.wcediter")
	log.Println("Fyne应用实例创建成功")

	// 创建窗口
	log.Println("正在创建主窗口...")
	window := fyneApp.NewWindow("风云存档编辑器 V1.0")
	currentWindow = window // 设置全局窗口变量
	log.Println("主窗口创建成功")

	// 设置窗口属性
	window.Resize(fyne.NewSize(800, 600))
	window.SetFixedSize(false) // 允许调整窗口大小
	window.SetOnClosed(func() {
		log.Println("窗口已关闭")
	})
	log.Println("窗口属性设置完成")

	// 创建进度选择界面内容
	log.Println("正在创建进度选择界面...")
	content := createProgressSelectUI(func(progressIndex int) {
		log.Printf("用户选择了进度: %s (索引: %d)", progressNames[progressIndex], progressIndex)
		// 加载选择的进度文件并打开角色属性窗口
		loadProgressAndOpenCharacterUI(progressIndex, window)
	})

	// 设置窗口内容
	window.SetContent(content)

	// 设置窗口大小为适合进度选择界面的尺寸
	window.Resize(fyne.NewSize(350, 250))

	// 显示窗口
	log.Println("正在显示进度选择窗口...")
	window.Canvas().Focus(nil) // 重置焦点
	window.CenterOnScreen()    // 居中显示窗口
	window.Show()
	log.Println("窗口显示命令已发送")
	log.Println("请注意：在macOS环境下，如果使用XQuartz，窗口可能会在单独的X11会话中显示")
	log.Println("如果看不到窗口，请检查是否有XQuartz窗口打开")

	// 不再单独显示进度选择对话框，因为已经设置为窗口内容

	// 运行应用
	log.Println("应用已启动，进入主事件循环")
	fyneApp.Run()
}

// 获取当前工作目录的辅助函数
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "无法获取当前目录"
	}
	return dir
}

// initConfigFile 初始化配置文件
func initConfigFile() {
	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 创建配置文件
		cfg := ini.Empty()

		// 添加默认记录
		defaultSection, err := cfg.NewSection("default_Records")
		if err != nil {
			log.Printf("创建default_Records节失败: %v", err)
			return
		}

		// 添加代码中固定的默认记录
		for _, tag := range defaultRecordOrder {
			path := defaultRecords[tag]
			_, err := defaultSection.NewKey(tag, path)
			if err != nil {
				log.Printf("添加默认记录失败: %v", err)
			}
		}

		// 创建普通记录节
		_, err = cfg.NewSection("Records")
		if err != nil {
			log.Printf("创建Records节失败: %v", err)
			return
		}

		// 保存配置文件
		err = cfg.SaveTo(configFile)
		if err != nil {
			log.Printf("保存配置文件失败: %v", err)
			return
		}

		log.Printf("创建配置文件成功: %s", configFile)
	} else {
		// 检查配置文件是否为空
		info, err := os.Stat(configFile)
		if err != nil {
			log.Printf("获取配置文件信息失败: %v", err)
			return
		}

		if info.Size() == 0 {
			// 配置文件为空，生成默认内容
			cfg := ini.Empty()

			// 添加默认记录
			defaultSection, err := cfg.NewSection("default_Records")
			if err != nil {
				log.Printf("创建default_Records节失败: %v", err)
				return
			}

			// 添加代码中固定的默认记录
			for _, tag := range defaultRecordOrder {
				path := defaultRecords[tag]
				_, err := defaultSection.NewKey(tag, path)
				if err != nil {
					log.Printf("添加默认记录失败: %v", err)
				}
			}

			// 创建普通记录节
			_, err = cfg.NewSection("Records")
			if err != nil {
				log.Printf("创建Records节失败: %v", err)
				return
			}

			// 保存配置文件
			err = cfg.SaveTo(configFile)
			if err != nil {
				log.Printf("保存配置文件失败: %v", err)
				return
			}

			log.Printf("配置文件为空，生成默认内容: %s", configFile)
		}
	}
}

// 更新银两值
func updateMoneyValue(valueStr string) bool {
	if editor == nil {
		dialog.ShowError(fmt.Errorf("没有加载的存档文件"), characterWindow)
		return false
	}
	val, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		dialog.ShowError(fmt.Errorf("银两格式错误: %v", err), characterWindow)
		return false
	}
	editor.UpdateMoney(int32(val))
	return true
}

// loadFileRecords 加载选择记录
func loadFileRecords() {
	// 确保配置文件存在
	initConfigFile()

	// 清空现有记录
	fileRecords = []FileRecordItem{}

	// 读取配置文件
	cfg, err := ini.Load(configFile)
	if err != nil {
		log.Printf("读取配置文件失败: %v", err)
		return
	}

	// 加载默认记录
	defaultSection := cfg.Section("default_Records")
	for _, key := range defaultSection.Keys() {
		item := FileRecordItem{
			Tag:       key.Name(),
			Path:      key.String(),
			IsDefault: true,
		}
		fileRecords = append(fileRecords, item)
	}

	// 加载普通记录
	recordsSection := cfg.Section("Records")
	for _, key := range recordsSection.Keys() {
		item := FileRecordItem{
			Tag:       key.Name(),
			Path:      key.String(),
			IsDefault: false,
		}
		fileRecords = append(fileRecords, item)
	}

	log.Printf("加载选择记录成功，共 %d 条记录", len(fileRecords))
}

// saveFileRecords 保存选择记录
func saveFileRecords() {
	// 创建新的配置文件
	cfg := ini.Empty()

	// 添加默认记录节
	defaultSection, err := cfg.NewSection("default_Records")
	if err != nil {
		log.Printf("创建default_Records节失败: %v", err)
		return
	}

	// 添加普通记录节
	recordsSection, err := cfg.NewSection("Records")
	if err != nil {
		log.Printf("创建Records节失败: %v", err)
		return
	}

	// 遍历所有记录，根据类型写入不同的节
	// 创建一个map来跟踪已使用的Tag
	usedTags := make(map[string]bool)
	for _, item := range fileRecords {
		// 检查Tag是否已被使用，如果是则添加数字后缀
		tag := item.Tag
		baseTag := tag
		counter := 1

		// 生成唯一的Tag
		for usedTags[tag] {
			tag = fmt.Sprintf("%s(%d)", baseTag, counter)
			counter++
		}

		// 根据记录类型写入不同的节
		if item.IsDefault {
			// 默认记录写入default_Records节
			_, err := defaultSection.NewKey(tag, item.Path)
			if err != nil {
				log.Printf("添加默认记录失败: %v", err)
			} else {
				// 标记Tag为已使用
				usedTags[tag] = true
			}
		} else {
			// 普通记录写入Records节
			_, err := recordsSection.NewKey(tag, item.Path)
			if err != nil {
				log.Printf("添加普通记录失败: %v", err)
			} else {
				// 标记Tag为已使用
				usedTags[tag] = true
			}
		}
	}

	// 保存配置文件
	err = cfg.SaveTo(configFile)
	if err != nil {
		log.Printf("保存配置文件失败: %v", err)
		return
	}

	log.Printf("保存配置文件成功: %s", configFile)
}

// getRelativePath 将绝对路径转换为相对路径（如果在当前目录下）
func getRelativePath(absPath string) string {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("获取当前工作目录失败: %v", err)
		return absPath
	}

	// 转换为绝对路径
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		log.Printf("转换为绝对路径失败: %v", err)
		return absPath
	}

	// 检查是否在当前目录下
	relPath, err := filepath.Rel(currentDir, absPath)
	if err != nil {
		log.Printf("转换为相对路径失败: %v", err)
		return absPath
	}

	// 如果相对路径以".."开头，说明不在当前目录下，返回绝对路径
	if strings.HasPrefix(relPath, "..") {
		return absPath
	}

	return relPath
}

// getFileName 获取文件名（不含路径）
func getFileName(filePath string) string {
	return filepath.Base(filePath)
}
