package wcsave

import (
	"os"

	"wcediter/wcsave/models"
	"wcediter/wcsave/reader"
	"wcediter/wcsave/writer"
)

// SaveEditor 是存档编辑器的主要接口
type SaveEditor struct {
	Characters []models.CharacterInfo
	MoneyInfo  models.MoneyInfo
}

// NewSaveEditor 创建一个新的存档编辑器实例
func NewSaveEditor() *SaveEditor {
	return &SaveEditor{
		Characters: make([]models.CharacterInfo, 0),
	}
}

// ReadSave 从文件中读取存档数据
func (e *SaveEditor) ReadSave(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 读取角色数据
	characters, err := reader.ReadCharacters(file)
	if err != nil {
		return err
	}
	e.Characters = characters

	// 读取银两数据
	moneyInfo, err := reader.ReadMoneyData(file, 203054)
	if err != nil {
		// 银两数据读取失败不会中断整体操作
		moneyInfo = models.MoneyInfo{}
	}
	e.MoneyInfo = moneyInfo

	return nil
}

// SaveChanges 将修改保存到新文件
func (e *SaveEditor) SaveChanges(sourceFilePath, destFilePath string) error {
	return writer.SaveChanges(sourceFilePath, destFilePath, e.Characters, e.MoneyInfo)
}

// GetCharacterCount 获取角色数量
func (e *SaveEditor) GetCharacterCount() int {
	return len(e.Characters)
}

// GetCharacterByIndex 通过索引获取角色信息
func (e *SaveEditor) GetCharacterByIndex(index int) (models.CharacterInfo, bool) {
	if index >= 0 && index < len(e.Characters) {
		return e.Characters[index], true
	}
	return models.CharacterInfo{}, false
}

// UpdateMoney 更新银两值
func (e *SaveEditor) UpdateMoney(value int32) {
	e.MoneyInfo.Value = value
}

// UpdateCharacter 更新角色信息
func (e *SaveEditor) UpdateCharacter(index int, data models.CharacterData) bool {
	if index >= 0 && index < len(e.Characters) {
		e.Characters[index].Data = data
		return true
	}
	return false
}