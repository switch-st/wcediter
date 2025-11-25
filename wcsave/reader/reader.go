package reader

import (
	"encoding/binary"
	"fmt"
	"os"

	"wcediter/wcsave/models"
	"wcediter/wcsave/utils"

	"golang.org/x/text/encoding/traditionalchinese"
)

// readMoneyData 读取银两数据的函数
func ReadMoneyData(file *os.File, position int64) (models.MoneyInfo, error) {
	var moneyInfo models.MoneyInfo
	moneyInfo.Position = position

	// 检查文件大小是否足够
	fileInfo, err := file.Stat()
	if err != nil {
		return moneyInfo, fmt.Errorf("无法获取文件信息: %v", err)
	}

	if position >= fileInfo.Size() {
		return moneyInfo, fmt.Errorf("文件大小不足以读取银两数据位置")
	}

	// 定位到银两数据位置
	_, err = file.Seek(position, 0)
	if err != nil {
		return moneyInfo, fmt.Errorf("无法定位到银两数据位置: %v", err)
	}

	// 读取4字节银两数据
	money, moneyRawBytes, err := utils.ReadAndConvert(file, 4, func(b []byte) (int32, error) {
		return int32(binary.LittleEndian.Uint32(b)), nil
	})

	if err != nil {
		return moneyInfo, fmt.Errorf("读取银两数据失败: %v", err)
	}

	moneyInfo = models.MoneyInfo{
		Value:    money,
		RawBytes: moneyRawBytes,
		Position: position,
	}

	return moneyInfo, nil
}

// readCharacterProperties 读取角色属性的函数
func readCharacterProperties(file *os.File) (models.CharacterData, models.RawByteData, []byte, int64, bool, error) {
	// 记录当前位置
	position, err := file.Seek(0, 1)
	if err != nil {
		// 如果获取位置失败，使用-1作为默认值
		position = -1
	}
	var data models.CharacterData
	var rawBytes models.RawByteData
	var utf8Name []byte

	// 检查起始2字节是否全为0
	_, startCheckBuffer, err := utils.ReadAndConvert[struct{}](file, 2, nil)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取起始2字节失败: %v", err)
	}

	// 如果起始2字节全为0，表示遇到结束条件
	if len(startCheckBuffer) >= 2 && startCheckBuffer[0] == 0 && startCheckBuffer[1] == 0 {
		return data, rawBytes, utf8Name, position, false, nil
	}

	// 如果不是结束条件，需要将文件指针回退2字节
	_, err = file.Seek(-2, 1)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("文件指针回退失败: %v", err)
	}

	// 定义名字编码转换器
	nameConverter := func(b []byte) ([]byte, error) {
		// 保留原始字节内容，直接从Big5转换为UTF-8
		utf8Bytes, err := traditionalchinese.Big5.NewDecoder().Bytes(b)
		if err != nil {
			// 转换失败时返回原始字节
			return b, nil
		}
		return utf8Bytes, nil
	}

	// 使用泛型方法读取并转换名字
	utf8Name, _, err = utils.ReadAndConvert(file, 6, nameConverter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取名字失败: %v", err)
	}

	// 在名字后跳过2个字节
	err = utils.SkipBytes(file, 2)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("跳过名字后字节失败: %v", err)
	}

	// 使用泛型方法读取4字节字段
	uint32Converter := func(b []byte) (int32, error) {
		return int32(binary.LittleEndian.Uint32(b)), nil
	}

	// 读取当前经验值（4字节）
	val, bytes, err := utils.ReadAndConvert(file, 4, uint32Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取当前经验值失败: %v", err)
	}
	data.CurrentExp = val
	rawBytes.CurrentExp = bytes

	// 读取升级经验值（4字节）
	val, bytes, err = utils.ReadAndConvert(file, 4, uint32Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取升级经验值失败: %v", err)
	}
	data.NextLevelExp = val
	rawBytes.NextLevelExp = bytes

	// 读取最大生命值（4字节）
	val, bytes, err = utils.ReadAndConvert(file, 4, uint32Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取最大生命值失败: %v", err)
	}
	data.MaxHP = val
	rawBytes.MaxHP = bytes

	// 读取最大内力值（4字节）
	val, bytes, err = utils.ReadAndConvert(file, 4, uint32Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取最大内力值失败: %v", err)
	}
	data.MaxMP = val
	rawBytes.MaxMP = bytes

	// 读取当前生命值（4字节）
	val, bytes, err = utils.ReadAndConvert(file, 4, uint32Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取当前生命值失败: %v", err)
	}
	data.CurrentHP = val
	rawBytes.CurrentHP = bytes

	// 读取当前内力值（4字节）
	val, bytes, err = utils.ReadAndConvert(file, 4, uint32Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取当前内力值失败: %v", err)
	}
	data.CurrentMP = val
	rawBytes.CurrentMP = bytes

	// 使用泛型方法读取2字节字段
	uint16Converter := func(b []byte) (int16, error) {
		return int16(binary.LittleEndian.Uint16(b)), nil
	}

	// 读取力量（2字节）
	val16, bytes, err := utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取力量失败: %v", err)
	}
	data.Strength = val16
	rawBytes.Strength = bytes

	// 读取反应（2字节）
	val16, bytes, err = utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取反应失败: %v", err)
	}
	data.Reaction = val16
	rawBytes.Reaction = bytes

	// 读取体质（2字节）
	val16, bytes, err = utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取体质失败: %v", err)
	}
	data.Constitution = val16
	rawBytes.Constitution = bytes

	// 读取速度（2字节）
	val16, bytes, err = utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取速度失败: %v", err)
	}
	data.Speed = val16
	rawBytes.Speed = bytes

	// 读取攻击（2字节）
	val16, bytes, err = utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取攻击失败: %v", err)
	}
	data.Attack = val16
	rawBytes.Attack = bytes

	// 读取防御（2字节）
	val16, bytes, err = utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取防御失败: %v", err)
	}
	data.Defense = val16
	rawBytes.Defense = bytes

	// 跳过8字节
	err = utils.SkipBytes(file, 8)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("跳过8字节失败: %v", err)
	}

	// 读取运气（2字节）
	val16, bytes, err = utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取运气失败: %v", err)
	}
	data.Luck = val16
	rawBytes.Luck = bytes

	// 跳过16字节
	err = utils.SkipBytes(file, 16)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("跳过16字节失败: %v", err)
	}

	// 读取等级（2字节）
	val16, bytes, err = utils.ReadAndConvert(file, 2, uint16Converter)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("读取等级失败: %v", err)
	}
	data.Level = val16
	rawBytes.Level = bytes

	// 跳过12字节
	err = utils.SkipBytes(file, 12)
	if err != nil {
		return data, rawBytes, utf8Name, position, false, fmt.Errorf("跳过12字节失败: %v", err)
	}

	return data, rawBytes, utf8Name, position, true, nil
}

// ReadCharacters 读取所有角色数据
func ReadCharacters(file *os.File) ([]models.CharacterInfo, error) {
	// 定位到指定位置
	targetPosition := int64(202618)
	_, err := file.Seek(targetPosition, 0)
	if err != nil {
		return nil, fmt.Errorf("无法定位到指定位置: %v", err)
	}

	// 存储所有角色信息
	characters := make([]models.CharacterInfo, 0)

	// 循环读取多个角色数据
	for {
		// 调用函数读取角色属性
		characterData, rawBytes, utf8Name, position, continueReading, err := readCharacterProperties(file)
		if err != nil {
			return characters, fmt.Errorf("读取角色属性时出错: %v", err)
		}

		// 检查是否遇到结束条件
		if !continueReading {
			break
		}

		// 保存角色信息
		characterName := "未知"
		if len(utf8Name) > 0 {
			characterName = string(utf8Name)
		}

		characters = append(characters, models.CharacterInfo{
			Name:     characterName,
			Data:     characterData,
			RawBytes: rawBytes,
			Position: position,
		})
	}

	return characters, nil
}