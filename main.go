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
	"fyne.io/fyne/v2/widget"
)

var (
	// 命令行参数
	saveFilePath string
	debugMode    bool

	// 全局状态
	fyneApp        fyne.App
	editor         *wcsave.SaveEditor
	currentSave    string
	currentWindow  fyne.Window
	charIndex      = 0
	propertyInputs []*propertyInput
	characterStats map[string]int

	// 存档进度相关
	progressNames = []string{"进度一", "进度二", "进度三", "进度四", "进度五"}
	progressFiles = []string{"Save0.dat", "Save1.dat", "Save2.dat", "Save3.dat", "Save4.dat"}
)

// 属性输入框结构体
type propertyInput struct {
	label    *widget.Label
	input    *widget.Entry
	property string
}

// 角色选择回调函数
type characterSelectCallback func(int)

// 进度选择回调函数
type progressSelectCallback func(int)

// 创建进度选择界面
func createProgressSelectUI(onSelect progressSelectCallback) *fyne.Container {
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
		selected := radioGroup.Selected
		if selected != "" {
			// 找到选中的进度索引
			for i, name := range progressNames {
				if name == selected {
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

	// 整个内容容器
	content := container.NewVBox(
		title,
		layout.NewSpacer(),
		radioCenterContainer,
		layout.NewSpacer(),
		buttonBox,
		authorLabel,
	)

	return content
}

// 显示进度选择对话框
func showProgressDialog(window fyne.Window, onSelect progressSelectCallback) {
	// 创建单选按钮组
	radio := widget.NewRadioGroup(progressNames, func(selected string) {
		// 选择后的处理在对话框按钮中处理
	})

	// 默认选择第一个进度
	if len(progressNames) > 0 {
		radio.SetSelected(progressNames[0])
	}

	// 创建内容容器
	content := container.NewVBox(
		widget.NewLabel("请选择欲修改的进度名:"),
		container.NewPadded(radio),
	)

	// 创建对话框
	dialog.ShowCustomConfirm(
		"进度选择",
		"确定",
		"取消",
		content,
		func(confirm bool) {
			if confirm && radio.Selected != "" {
				// 查找选择的进度索引
				for i, name := range progressNames {
					if name == radio.Selected {
						if onSelect != nil {
							onSelect(i)
						}
						break
					}
				}
			}
		},
		window,
	)
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
	if progressIndex < 0 || progressIndex >= len(progressFiles) {
		log.Printf("无效的进度索引: %d", progressIndex)
		dialog.ShowError(fmt.Errorf("无效的进度索引"), progressWindow)
		return
	}

	// 构建存档文件路径
	// 首先尝试在data目录中查找
	dataDir := filepath.Join(getCurrentDir(), "data")
	filePath := filepath.Join(dataDir, progressFiles[progressIndex])

	// 如果data目录中的文件不存在，尝试其他可能的路径
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 尝试直接在当前目录查找
		filePath = progressFiles[progressIndex]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// 尝试在桌面查找
			home, err := os.UserHomeDir()
			if err == nil {
				desktopPath := filepath.Join(home, "Desktop")
				filePath = filepath.Join(desktopPath, progressFiles[progressIndex])
			}
		}
	}

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
	characterWindow := fyneApp.NewWindow(title)

	// 设置更大的窗口大小以确保所有角色属性都能完整显示
	characterWindow.Resize(fyne.NewSize(650, 800))

	// 设置窗口关闭时的行为
	characterWindow.SetCloseIntercept(func() {
		log.Println("角色属性窗口关闭中...")
		// 保存当前编辑的内容
		if len(propertyInputs) > 0 && charIndex >= 0 {
			saveCharacterChanges(charIndex, propertyInputs)
		}
		// 关闭窗口并重新显示进度选择窗口
		characterWindow.Close()
		// 重新显示进度选择窗口
		log.Println("重新显示进度选择窗口...")
		currentWindow.Show()
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

// 根据进度索引加载存档文件
func loadProgressFile(progressIndex int) {
	if progressIndex < 0 || progressIndex >= len(progressFiles) {
		log.Printf("无效的进度索引: %d", progressIndex)
		dialog.ShowError(fmt.Errorf("无效的进度索引"), nil)
		return
	}

	// 构建存档文件路径
	// 首先尝试在data目录中查找
	dataDir := filepath.Join(getCurrentDir(), "data")
	filePath := filepath.Join(dataDir, progressFiles[progressIndex])

	// 如果data目录中的文件不存在，尝试其他可能的路径
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 尝试直接在当前目录查找
		filePath = progressFiles[progressIndex]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// 尝试在桌面查找
			home, err := os.UserHomeDir()
			if err == nil {
				desktopPath := filepath.Join(home, "Desktop")
				filePath = filepath.Join(desktopPath, progressFiles[progressIndex])
			}
		}
	}

	// 加载找到的存档文件
	err := loadSaveFile(filePath)
	if err != nil {
		log.Printf("加载存档失败: %v", err)
		dialog.ShowError(fmt.Errorf("加载存档失败: %v", err), currentWindow)
	} else {
		log.Printf("成功加载进度: %s, 文件: %s", progressNames[progressIndex], filePath)
		// 重新创建并更新UI
		content := createMainUI()
		currentWindow.SetContent(content)
	}
}

// 加载存档文件
func loadSaveFile(filePath string) error {
	// 如果文件不存在，尝试使用默认路径
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 检查是否有默认存档文件
		defaultPath := filepath.Join(os.Getenv("HOME"), "Desktop", "风云存档.dat")
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			log.Printf("警告: 未找到存档文件: %s 和 %s", filePath, defaultPath)
			// 返回nil，允许程序继续运行
			return nil
		}
		filePath = defaultPath
		log.Printf("使用默认存档文件: %s", filePath)
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
func createCharacterSelector(onCharacterChange characterSelectCallback) *widget.Select {
	options := []string{}
	if editor != nil {
		for i := 0; i < editor.GetCharacterCount(); i++ {
			if char, ok := editor.GetCharacterByIndex(i); ok {
				options = append(options, fmt.Sprintf("%d. %s", i+1, char.Name))
			}
		}
	}

	selector := widget.NewSelect(options, func(value string) {
		if value != "" {
			// 提取角色索引
			parts := strings.Split(value, ".")
			if len(parts) > 0 {
				index, err := strconv.Atoi(strings.TrimSpace(parts[0]))
				if err == nil {
					charIndex = index - 1
					if onCharacterChange != nil {
						onCharacterChange(charIndex)
					}
				}
			}
		}
	})

	// 默认选择第一个角色
	if len(options) > 0 {
		selector.SetSelected(options[0])
	}

	return selector
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

	// 显示提示信息
	statusLabel := widget.NewLabel("")
	if currentSave == "" {
		statusLabel.SetText("提示: 请先准备风云存档文件并通过命令行参数指定")
		statusLabel.TextStyle = fyne.TextStyle{Italic: true}
	} else {
		statusLabel.SetText(fmt.Sprintf("已加载存档: %s", currentSave))
	}

	// 创建角色选择器
	characterSelector := createCharacterSelector(func(index int) {
		// 角色切换时的处理逻辑将在创建完propertyInputs后添加
	})

	// 创建属性输入框
	propertyInputs := []*propertyInput{
		createPropertyInput("CurrentExp", "当前经验值:", "0"),
		createPropertyInput("NextLevelExp", "升级经验值:", "0"),
		createPropertyInput("CurrentHP", "当前生命值:", "0"),
		createPropertyInput("CurrentMP", "当前内力值:", "0"),
		createPropertyInput("MaxHP", "最大生命值:", "0"),
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

	// 重新设置角色选择回调，确保可以访问propertyInputs
	characterSelector.OnChanged = func(value string) {
		if value != "" {
			parts := strings.Split(value, ".")
			if len(parts) > 0 {
				index, err := strconv.Atoi(strings.TrimSpace(parts[0]))
				if err == nil {
					charIndex = index - 1
					updateCharacterUI(charIndex, propertyInputs)
				}
			}
		}
	}

	// 获取银两值
	if editor != nil {
		moneyValue = strconv.FormatInt(int64(editor.MoneyInfo.Value), 10)
	}

	// 创建银两相关的输入框和按钮
	moneyLabel := widget.NewLabel("银两数量:")
	moneyInput := widget.NewEntry()
	moneyInput.SetText(moneyValue)
	moneyInput.SetPlaceHolder("请输入银两数量")

	// 创建银两修改按钮
	moneyButton := widget.NewButton("修改银两", func() {
		if editor == nil {
			dialog.ShowError(fmt.Errorf("没有加载的存档文件"), nil)
			return
		}
		valueStr := moneyInput.Text
		val, err := strconv.ParseInt(valueStr, 10, 32)
		if err != nil {
			dialog.ShowError(fmt.Errorf("银两格式错误: %v", err), nil)
			return
		}
		editor.UpdateMoney(int32(val))
		dialog.ShowInformation("成功", "银两修改成功", nil)
	})

	// 创建保存角色按钮
	saveButton := widget.NewButton("保存角色数据", func() {
		if editor == nil || editor.GetCharacterCount() == 0 {
			dialog.ShowError(fmt.Errorf("没有加载的存档文件或角色数据"), nil)
			return
		}
		err := saveCharacterChanges(charIndex, propertyInputs)
		if err != nil {
			dialog.ShowError(err, nil)
			return
		}
		dialog.ShowInformation("成功", "角色数据保存成功", nil)
	})

	// 创建保存文件按钮
	saveFileButton := widget.NewButton("保存更改到文件", func() {
		if currentSave == "" {
			dialog.ShowError(fmt.Errorf("没有加载的存档文件"), nil)
			return
		}

		// 创建备份文件
		backupPath := currentSave + ".bak"
		// 复制当前文件到备份
		sourceFile, err := os.Open(currentSave)
		if err != nil {
			dialog.ShowError(fmt.Errorf("打开源文件失败: %v", err), nil)
			return
		}
		defer sourceFile.Close()

		destFile, err := os.Create(backupPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("创建备份文件失败: %v", err), nil)
			return
		}
		defer destFile.Close()

		// 直接复制文件内容
		buffer := make([]byte, 1024)
		for {
			n, err := sourceFile.Read(buffer)
			if err != nil {
				break
			}
			destFile.Write(buffer[:n])
		}

		// 保存更改
		newFilePath := currentSave
		err = editor.SaveChanges(currentSave, newFilePath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("保存文件失败: %v", err), nil)
			return
		}

		dialog.ShowInformation("成功", fmt.Sprintf("更改已保存到: %s\n备份已创建: %s", newFilePath, backupPath), nil)
	})

	// 创建属性网格 - 一行两列布局，输入框水平排列在属性文字后面
	propertyGrid := container.NewGridWithColumns(2)
	// 将每个属性（标签+输入框）组合成一个水平容器，然后添加到网格中
	for _, input := range propertyInputs {
		// 为输入框设置固定宽度
		input.input.Wrapping = fyne.TextWrapOff
		input.input.Resize(fyne.NewSize(200, 30))

		// 为每个属性创建一个水平容器，使用layout.NewSpacer()让输入框扩展
		propertyItem := container.NewHBox(
			input.label,
			layout.NewSpacer(), // 添加这个可以让输入框占据更多空间
			input.input,
		)
		// 为整个属性项设置最小宽度，确保输入框能够显示完整
		propertyItem.Resize(fyne.NewSize(300, 30))
		propertyGrid.Add(propertyItem)
	}

	// 创建按钮容器
	buttonContainer := container.NewGridWithColumns(3, saveButton, saveFileButton)

	// 创建银两容器
	moneyContainer := container.NewGridWithColumns(3, moneyLabel, moneyInput, moneyButton)

	// 创建主容器
	mainContainer := container.NewVBox(
		// 隐藏标题和状态信息
		widget.NewSeparator(),
		widget.NewLabel("选择角色:"),
		characterSelector,
		widget.NewSeparator(),
		widget.NewLabel("角色属性:"),
		// 为属性网格添加边框效果
		container.NewPadded(
			widget.NewCard(
				"", // 无边框标题
				"", // 无边框副标题
				propertyGrid,
			),
		),
		widget.NewSeparator(),
		moneyContainer,
		layout.NewSpacer(),
		buttonContainer,
	)

	// 初始化显示第一个角色的数据
	if editor != nil && editor.GetCharacterCount() > 0 {
		updateCharacterUI(charIndex, propertyInputs)
	}

	return mainContainer
}

func main() {
	// 解析命令行参数
	flag.StringVar(&saveFilePath, "file", "", "存档文件路径")
	flag.BoolVar(&debugMode, "debug", false, "调试模式")
	flag.Parse()

	// 打印环境信息用于调试
	log.Println("风云存档编辑器启动中...")
	log.Printf("操作系统: macOS")
	log.Printf("当前工作目录: %s", getCurrentDir())

	// 初始化角色属性映射
	characterStats = make(map[string]int)

	// 检查DISPLAY环境变量（对macOS X11很重要）
	display := os.Getenv("DISPLAY")
	if display == "" {
		log.Println("警告: DISPLAY环境变量未设置，可能会影响GUI显示")
		log.Println("如果使用X11，请确保已安装XQuartz并正确配置")
	} else {
		log.Printf("DISPLAY环境变量: %s", display)
	}

	// 创建Fyne应用
	log.Println("正在创建Fyne应用实例...")
	// 为macOS添加显示配置选项
	fyneApp = app.New()
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
