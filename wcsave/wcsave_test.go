package wcsave

import (
	"os"
	"testing"

	"wcediter/wcsave/models"
)

// 测试NewSaveEditor函数
func TestNewSaveEditor(t *testing.T) {
	editor := NewSaveEditor()
	if editor == nil {
		t.Fatal("NewSaveEditor返回nil")
	}
	if len(editor.Characters) != 0 {
		t.Errorf("Characters长度应为0，实际为%d", len(editor.Characters))
	}
}

// 测试GetCharacterCount函数
func TestGetCharacterCount(t *testing.T) {
	editor := NewSaveEditor()
	editor.Characters = append(editor.Characters, models.CharacterInfo{Name: "Test"})
	editor.Characters = append(editor.Characters, models.CharacterInfo{Name: "Test2"})
	
	if count := editor.GetCharacterCount(); count != 2 {
		t.Errorf("GetCharacterCount应返回2，实际返回%d", count)
	}
}

// 测试GetCharacterByIndex函数
func TestGetCharacterByIndex(t *testing.T) {
	editor := NewSaveEditor()
	editor.Characters = append(editor.Characters, models.CharacterInfo{Name: "Test"})
	editor.Characters = append(editor.Characters, models.CharacterInfo{Name: "Test2"})
	
	// 测试有效索引
	char, found := editor.GetCharacterByIndex(0)
	if !found || char.Name != "Test" {
		t.Errorf("获取索引0的角色失败")
	}
	
	char, found = editor.GetCharacterByIndex(1)
	if !found || char.Name != "Test2" {
		t.Errorf("获取索引1的角色失败")
	}
	
	// 测试无效索引
	_, found = editor.GetCharacterByIndex(-1)
	if found {
		t.Errorf("索引-1应该返回false")
	}
	
	_, found = editor.GetCharacterByIndex(2)
	if found {
		t.Errorf("索引2应该返回false")
	}
}

// 测试UpdateMoney函数
func TestUpdateMoney(t *testing.T) {
	editor := NewSaveEditor()
	editor.MoneyInfo.Value = 100
	
	editor.UpdateMoney(200)
	if editor.MoneyInfo.Value != 200 {
		t.Errorf("UpdateMoney失败，预期200，实际%d", editor.MoneyInfo.Value)
	}
}

// 测试UpdateCharacter函数
func TestUpdateCharacter(t *testing.T) {
	editor := NewSaveEditor()
	char := models.CharacterInfo{
		Name: "Test",
		Data: models.CharacterData{Strength: 10},
	}
	editor.Characters = append(editor.Characters, char)
	
	// 更新有效索引
	newData := models.CharacterData{Strength: 20}
	result := editor.UpdateCharacter(0, newData)
	if !result {
		t.Errorf("UpdateCharacter应该返回true")
	}
	if editor.Characters[0].Data.Strength != 20 {
		t.Errorf("角色数据更新失败，预期20，实际%d", editor.Characters[0].Data.Strength)
	}
	
	// 更新无效索引
	result = editor.UpdateCharacter(1, newData)
	if result {
		t.Errorf("更新无效索引应该返回false")
	}
}

// 测试ReadSave和SaveChanges功能（集成测试）
func TestSaveEditorIntegration(t *testing.T) {
	// 这个测试需要一个有效的测试数据文件
	testFilePath := "../data/Save4.dat"
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("测试数据文件不存在，跳过集成测试")
	}
	
	editor := NewSaveEditor()
	err := editor.ReadSave(testFilePath)
	if err != nil {
		t.Fatalf("读取测试文件失败: %v", err)
	}
	
	// 验证是否读取到了数据
	if len(editor.Characters) == 0 {
		t.Error("没有读取到角色数据")
	}
}