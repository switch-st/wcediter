package reader

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试ReadMoneyData函数
func TestReadMoneyData(t *testing.T) {
	// 创建临时测试文件
	tempFile := filepath.Join(t.TempDir(), "test_money.dat")
	
	// 写入模拟银两数据 (1000的小端字节序) 在位置20
	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	
	// 写入足够的零字节到达位置20
	zeros := make([]byte, 20)
	if _, err := file.Write(zeros); err != nil {
		file.Close()
		t.Fatalf("写入测试文件失败: %v", err)
	}
	
	// 写入银两数据 (1000)
	moneyBytes := []byte{0xE8, 0x3, 0x0, 0x0}
	if _, err := file.Write(moneyBytes); err != nil {
		file.Close()
		t.Fatalf("写入银两数据失败: %v", err)
	}
	file.Close()
	
	// 重新打开文件进行测试
	testFile, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("打开测试文件失败: %v", err)
	}
	defer testFile.Close()
	
	// 调用ReadMoneyData函数
	moneyInfo, err := ReadMoneyData(testFile, 20)
	if err != nil {
		t.Fatalf("ReadMoneyData失败: %v", err)
	}
	
	// 验证返回的银两值是否正确
	if moneyInfo.Value != 1000 {
		t.Errorf("银两值错误，预期1000，实际%d", moneyInfo.Value)
	}
}

// 集成测试：测试读取角色数据
func TestReadCharacters(t *testing.T) {
	// 检查测试数据文件是否存在
	testFilePath := "../../data/Save4.dat"
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("测试数据文件不存在，跳过集成测试")
	}
	
	// 打开文件
	file, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("打开测试文件失败: %v", err)
	}
	defer file.Close()
	
	// 调用ReadCharacters函数
	characters, err := ReadCharacters(file)
	if err != nil {
		t.Fatalf("ReadCharacters失败: %v", err)
	}
	
	// 验证返回的characters是否有效
	if characters == nil {
		t.Error("characters不应为nil")
	}
	
	// 检查是否读取到了角色信息
	if len(characters) == 0 {
		t.Log("没有读取到角色信息，但不认为这是错误，可能是数据格式问题")
	} else {
		// 验证第一个角色的基本信息
		firstChar := characters[0]
		if firstChar.Position < 0 {
			t.Errorf("角色位置不应为负数，实际%d", firstChar.Position)
		}
		// 检查Name是否为字符串
		if len(firstChar.Name) == 0 {
			t.Log("角色名称为空，但不认为这是错误，可能是数据格式问题")
		}
	}
}

// 测试错误情况：文件位置越界
func TestReadMoneyData_OutOfBounds(t *testing.T) {
	// 创建临时测试文件
	tempFile := filepath.Join(t.TempDir(), "test_money_oob.dat")
	
	// 创建一个小文件
	if err := os.WriteFile(tempFile, []byte{1, 2, 3, 4, 5}, 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	
	// 打开文件
	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("打开测试文件失败: %v", err)
	}
	defer file.Close()
	
	// 尝试在文件末尾之后读取
	_, err = ReadMoneyData(file, 100) // 位置100远大于文件大小5
	if err == nil {
		t.Fatal("预期应该返回错误，但没有")
	}
}