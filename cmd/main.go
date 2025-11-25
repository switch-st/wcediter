package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"wcediter/wcsave"
	"wcediter/wcsave/models"
)

func main() {
	// 命令行参数解析
	sourceFilePathFlag := flag.String("input", "", "输入存档文件路径")
	destFilePathFlag := flag.String("output", "", "输出存档文件路径")
	flag.Parse()

	// 使用命令行参数
	sourceFilePath := *sourceFilePathFlag
	destFilePath := *destFilePathFlag
	
	// 检查输入文件是否设置
	if sourceFilePath == "" {
		fmt.Println("错误: 必须使用 -input 参数指定输入存档文件路径")
		fmt.Println("使用示例: go run main.go -input Save.dat [-output Save_modified.dat]")
		os.Exit(1)
	}

	// 显示程序信息和使用说明
	fmt.Println("===================================")
	fmt.Println("游戏存档编辑器")
	fmt.Println("===================================")
	fmt.Printf("当前使用的输入文件: %s\n", sourceFilePath)
	if destFilePath != "" {
		fmt.Printf("当前使用的输出文件: %s\n", destFilePath)
	} else {
		fmt.Println("未指定输出文件，将以只读模式运行")
	}
	fmt.Println()
	fmt.Println("使用说明:")
	fmt.Println("  -input <文件路径>  指定输入存档文件路径 (必需)")
	fmt.Println("  -output <文件路径> 指定输出存档文件路径 (可选)")
	fmt.Println("例如:")
	fmt.Println("  go run main.go -input Save.dat -output Save_modified.dat\n")

	// 内部函数定义
	getUserInput := func(prompt string) string {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print(prompt)
		scanner.Scan()
		return strings.TrimSpace(scanner.Text())
	}

	getConfirmation := func(prompt string) bool {
		for {
			response := getUserInput(prompt)
			response = strings.ToLower(response)
			if response == "y" || response == "是" {
				return true
			} else if response == "n" || response == "否" {
				return false
			}
			fmt.Println("请输入 y/n 或 是/否")
		}
	}

	// 创建存档编辑器实例
	editor := wcsave.NewSaveEditor()

	// 读取存档数据
	err := editor.ReadSave(sourceFilePath)
	if err != nil {
		fmt.Printf("读取存档文件失败: %v\n", err)
		os.Exit(1)
	}

	// 输出总结信息
	fmt.Println("\n=== 读取总结 ===")
	fmt.Printf("总共成功读取了 %d 个角色的信息\n", editor.GetCharacterCount())

	// 统一输出所有角色的详细属性信息
	fmt.Println("\n=== 角色详细属性信息 ===")

	for i := 0; i < editor.GetCharacterCount(); i++ {
		char, _ := editor.GetCharacterByIndex(i)
		fmt.Printf("\n----- 角色 %d: %s -----\n\n", i+1, char.Name)

		// 统一输出所有属性
		fmt.Printf("当前经验值: %d\n", char.Data.CurrentExp)
		fmt.Printf("升级经验值: %d\n", char.Data.NextLevelExp)
		fmt.Printf("当前生命值: %d\n", char.Data.CurrentHP)
		fmt.Printf("当前内力值: %d\n", char.Data.CurrentMP)
		fmt.Printf("最大生命值: %d\n", char.Data.MaxHP)
		fmt.Printf("最大内力值: %d\n", char.Data.MaxMP)
		fmt.Printf("力量: %d\n", char.Data.Strength)
		fmt.Printf("反应: %d\n", char.Data.Reaction)
		fmt.Printf("体质: %d\n", char.Data.Constitution)
		fmt.Printf("速度: %d\n", char.Data.Speed)
		fmt.Printf("攻击: %d\n", char.Data.Attack)
		fmt.Printf("防御: %d\n", char.Data.Defense)
		fmt.Printf("运气: %d\n", char.Data.Luck)
		fmt.Printf("等级: %d\n", char.Data.Level)
	}

	// 角色列表概览
	fmt.Println("\n=== 角色列表概览 ===")
	for i := 0; i < editor.GetCharacterCount(); i++ {
		char, _ := editor.GetCharacterByIndex(i)
		fmt.Printf("角色 %d: %s\n", i+1, char.Name)
	}

	// 显示银两数据
	fmt.Println("\n=== 银两数据 ===")
	fmt.Printf("银两: %d\n", editor.MoneyInfo.Value)

	// 只有在指定了输出文件时才显示修改功能
	var needModifications bool = false
	if destFilePath != "" {
		// 开始命令行交互修改
		fmt.Println("\n=== 修改功能 ===")

		// 询问是否需要修改银两
		if editor.MoneyInfo.Position != 0 {
			if getConfirmation("是否需要修改银两字段？(y/n): ") {
				needModifications = true
				fmt.Printf("当前银两: %d\n", editor.MoneyInfo.Value)

				// 使用循环确保获取有效的输入
				for {
					newMoneyStr := getUserInput("请输入新的银两值: ")
					newMoney, moneyErr := strconv.ParseInt(newMoneyStr, 10, 32)
					if moneyErr != nil {
						fmt.Printf("无效的数字输入: %v\n", moneyErr)
						continue
					}

					// 保存修改前的值用于对比
					oldMoney := editor.MoneyInfo.Value
					editor.UpdateMoney(int32(newMoney))
					fmt.Printf("银两修改成功: %d -> %d\n", oldMoney, editor.MoneyInfo.Value)
					break
				}
			}
		}

		// 针对每个角色，询问是否需要修改属性
		for i := 0; i < editor.GetCharacterCount(); i++ {
			char, _ := editor.GetCharacterByIndex(i)
			fmt.Printf("\n==== 角色 %d: %s ====\n", i+1, char.Name)
			if getConfirmation("是否需要修改该角色的属性？(y/n): ") {
				needModifications = true

				for {
					fmt.Println("\n可用属性列表:")
					fmt.Println("1. 当前经验值")
					fmt.Println("2. 升级经验值")
					fmt.Println("3. 当前生命值")
					fmt.Println("4. 当前内力值")
					fmt.Println("5. 最大生命值")
					fmt.Println("6. 最大内力值")
					fmt.Println("7. 力量")
					fmt.Println("8. 反应")
					fmt.Println("9. 体质")
					fmt.Println("10. 速度")
					fmt.Println("11. 攻击")
					fmt.Println("12. 防御")
					fmt.Println("13. 运气")
					fmt.Println("14. 等级")
					fmt.Println("0. 完成该角色的修改")

					attrChoiceStr := getUserInput("请选择要修改的属性编号: ")
					attrChoice, err := strconv.Atoi(attrChoiceStr)
					if err != nil || attrChoice < 0 || attrChoice > 14 {
						fmt.Println("无效的属性编号，请重新选择")
						continue
					}

					if attrChoice == 0 {
						break
					}

					var valueName string
					var is32Bit bool

					switch attrChoice {
					case 1:
						valueName = "当前经验值"
						is32Bit = true
					case 2:
						valueName = "升级经验值"
						is32Bit = true
					case 3:
						valueName = "当前生命值"
						is32Bit = true
					case 4:
						valueName = "当前内力值"
						is32Bit = true
					case 5:
						valueName = "最大生命值"
						is32Bit = true
					case 6:
						valueName = "最大内力值"
						is32Bit = true
					case 7:
						valueName = "力量"
						is32Bit = false
					case 8:
						valueName = "反应"
						is32Bit = false
					case 9:
						valueName = "体质"
						is32Bit = false
					case 10:
						valueName = "速度"
						is32Bit = false
					case 11:
						valueName = "攻击"
						is32Bit = false
					case 12:
						valueName = "防御"
						is32Bit = false
					case 13:
						valueName = "运气"
						is32Bit = false
					case 14:
						valueName = "等级"
						is32Bit = false
					}

					newValueStr := getUserInput(fmt.Sprintf("请输入新的%s值: ", valueName))

					charData := char.Data
					var oldValue interface{}
					var newValue interface{}

					if is32Bit {
						newValueInt, parseErr := strconv.ParseInt(newValueStr, 10, 32)
						if parseErr != nil {
							fmt.Printf("无效的数字输入: %v\n", parseErr)
							continue
						}
						newValue = int32(newValueInt)

						switch attrChoice {
						case 1:
							oldValue = charData.CurrentExp
							charData.CurrentExp = newValue.(int32)
						case 2:
							oldValue = charData.NextLevelExp
							charData.NextLevelExp = newValue.(int32)
						case 3:
							oldValue = charData.CurrentHP
							charData.CurrentHP = newValue.(int32)
						case 4:
							oldValue = charData.CurrentMP
							charData.CurrentMP = newValue.(int32)
						case 5:
							oldValue = charData.MaxHP
							charData.MaxHP = newValue.(int32)
						case 6:
							oldValue = charData.MaxMP
							charData.MaxMP = newValue.(int32)
						}
					} else {
						newValueInt, parseErr := strconv.ParseInt(newValueStr, 10, 16)
						if parseErr != nil {
							fmt.Printf("无效的数字输入: %v\n", parseErr)
							continue
						}
						newValue = int16(newValueInt)

						switch attrChoice {
						case 7:
							oldValue = charData.Strength
							charData.Strength = newValue.(int16)
						case 8:
							oldValue = charData.Reaction
							charData.Reaction = newValue.(int16)
						case 9:
							oldValue = charData.Constitution
							charData.Constitution = newValue.(int16)
						case 10:
							oldValue = charData.Speed
							charData.Speed = newValue.(int16)
						case 11:
							oldValue = charData.Attack
							charData.Attack = newValue.(int16)
						case 12:
							oldValue = charData.Defense
							charData.Defense = newValue.(int16)
						case 13:
							oldValue = charData.Luck
							charData.Luck = newValue.(int16)
						case 14:
							oldValue = charData.Level
							charData.Level = newValue.(int16)
						}
					}

					editor.UpdateCharacter(i, charData)
					char = models.CharacterInfo{Name: char.Name, Data: charData, RawBytes: char.RawBytes, Position: char.Position}

					fmt.Printf("%s 修改成功: %v -> %v\n", valueName, oldValue, newValue)
				}
			}
		}

		// 如果有修改且指定了输出文件，保存修改
		if needModifications {
			fmt.Println("\n=== 保存修改 ===")
			err = editor.SaveChanges(sourceFilePath, destFilePath)
			if err != nil {
				fmt.Printf("保存修改失败: %v\n", err)
			} else {
				fmt.Printf("已创建修改后的文件: %s\n", destFilePath)

				// 添加修改前后的对比显示
				fmt.Println("\n=== 修改前后对比 ===")

				// 银两对比
				if editor.MoneyInfo.Position != 0 && len(editor.MoneyInfo.RawBytes) > 0 {
					oldMoney := int32(binary.LittleEndian.Uint32(editor.MoneyInfo.RawBytes))
					fmt.Printf("\n银两: [%d] -> [%d]", oldMoney, editor.MoneyInfo.Value)
					if oldMoney != editor.MoneyInfo.Value {
						fmt.Printf(" ✓ (已修改)")
					} else {
						fmt.Printf(" (未修改)")
					}
					fmt.Println()
				}

				// 角色属性对比
				fmt.Println("\n角色属性修改对比:")
				for i := 0; i < editor.GetCharacterCount(); i++ {
					char, _ := editor.GetCharacterByIndex(i)
					hasChanged := false

					// 先检查是否有属性被修改
					currentExpOld := int32(binary.LittleEndian.Uint32(char.RawBytes.CurrentExp))
					nextLevelExpOld := int32(binary.LittleEndian.Uint32(char.RawBytes.NextLevelExp))
					currentHPOld := int32(binary.LittleEndian.Uint32(char.RawBytes.CurrentHP))
					currentMPOld := int32(binary.LittleEndian.Uint32(char.RawBytes.CurrentMP))
					maxHPOld := int32(binary.LittleEndian.Uint32(char.RawBytes.MaxHP))
					maxMPOld := int32(binary.LittleEndian.Uint32(char.RawBytes.MaxMP))
					strengthOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Strength))
					reactionOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Reaction))
					constitutionOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Constitution))
					speedOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Speed))
					attackOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Attack))
					defenseOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Defense))
					luckOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Luck))
					levelOld := int16(binary.LittleEndian.Uint16(char.RawBytes.Level))

					hasChanged = currentExpOld != char.Data.CurrentExp ||
						nextLevelExpOld != char.Data.NextLevelExp ||
						currentHPOld != char.Data.CurrentHP ||
						currentMPOld != char.Data.CurrentMP ||
						maxHPOld != char.Data.MaxHP ||
						maxMPOld != char.Data.MaxMP ||
						strengthOld != char.Data.Strength ||
						reactionOld != char.Data.Reaction ||
						constitutionOld != char.Data.Constitution ||
						speedOld != char.Data.Speed ||
						attackOld != char.Data.Attack ||
						defenseOld != char.Data.Defense ||
						luckOld != char.Data.Luck ||
						levelOld != char.Data.Level

					// 只显示有修改的角色
					if hasChanged {
						fmt.Printf("\n----- 角色: %s -----\n", char.Name)

						// 显示修改的属性对比
						if currentExpOld != char.Data.CurrentExp {
							fmt.Printf("当前经验值: [%d] -> [%d] ✓\n", currentExpOld, char.Data.CurrentExp)
						}
						if nextLevelExpOld != char.Data.NextLevelExp {
							fmt.Printf("升级经验值: [%d] -> [%d] ✓\n", nextLevelExpOld, char.Data.NextLevelExp)
						}
						if currentHPOld != char.Data.CurrentHP {
							fmt.Printf("当前生命值: [%d] -> [%d] ✓\n", currentHPOld, char.Data.CurrentHP)
						}
						if currentMPOld != char.Data.CurrentMP {
							fmt.Printf("当前内力值: [%d] -> [%d] ✓\n", currentMPOld, char.Data.CurrentMP)
						}
						if maxHPOld != char.Data.MaxHP {
							fmt.Printf("最大生命值: [%d] -> [%d] ✓\n", maxHPOld, char.Data.MaxHP)
						}
						if maxMPOld != char.Data.MaxMP {
							fmt.Printf("最大内力值: [%d] -> [%d] ✓\n", maxMPOld, char.Data.MaxMP)
						}
						if strengthOld != char.Data.Strength {
							fmt.Printf("力量: [%d] -> [%d] ✓\n", strengthOld, char.Data.Strength)
						}
						if reactionOld != char.Data.Reaction {
							fmt.Printf("反应: [%d] -> [%d] ✓\n", reactionOld, char.Data.Reaction)
						}
						if constitutionOld != char.Data.Constitution {
							fmt.Printf("体质: [%d] -> [%d] ✓\n", constitutionOld, char.Data.Constitution)
						}
						if speedOld != char.Data.Speed {
							fmt.Printf("速度: [%d] -> [%d] ✓\n", speedOld, char.Data.Speed)
						}
						if attackOld != char.Data.Attack {
							fmt.Printf("攻击: [%d] -> [%d] ✓\n", attackOld, char.Data.Attack)
						}
						if defenseOld != char.Data.Defense {
							fmt.Printf("防御: [%d] -> [%d] ✓\n", defenseOld, char.Data.Defense)
						}
						if luckOld != char.Data.Luck {
							fmt.Printf("运气: [%d] -> [%d] ✓\n", luckOld, char.Data.Luck)
						}
						if levelOld != char.Data.Level {
							fmt.Printf("等级: [%d] -> [%d] ✓\n", levelOld, char.Data.Level)
						}
					}
				}
			}
		}
	}

	fmt.Println("\n操作完成！")
}