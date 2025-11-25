package writer

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"wcediter/wcsave/models"
)

// 测试copyFile函数
func TestCopyFile(t *testing.T) {
	// 创建临时测试文件
	srcPath := filepath.Join(t.TempDir(), "source.dat")
	dstPath := filepath.Join(t.TempDir(), "dest.dat")
	
	// 写入测试数据到源文件
	srcData := []byte{1, 2, 3, 4, 5}
	if err := os.WriteFile(srcPath, srcData, 0644); err != nil {
		t.Fatalf("创建源测试文件失败: %v", err)
	}
	
	// 测试复制文件
	err := copyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyFile失败: %v", err)
	}
	
	// 验证目标文件是否创建成功
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Fatal("目标文件未创建")
	}
	
	// 验证目标文件内容是否正确
	dstData, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("读取目标文件失败: %v", err)
	}
	
	if len(dstData) != len(srcData) {
		t.Errorf("目标文件长度不匹配，预期%d，实际%d", len(srcData), len(dstData))
	}
}

// 测试writeToFilePosition函数
func TestWriteToFilePosition(t *testing.T) {
	// 创建临时测试文件
	filePath := filepath.Join(t.TempDir(), "test.dat")
	
	// 写入初始数据
	initData := make([]byte, 20)
	if err := os.WriteFile(filePath, initData, 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	
	// 测试写入指定位置
	testData := []byte{10, 20, 30, 40}
	err := writeToFilePosition(filePath, 5, testData)
	if err != nil {
		t.Fatalf("writeToFilePosition失败: %v", err)
	}
	
	// 验证数据是否正确写入
	resultData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取测试文件失败: %v", err)
	}
	
	for i, val := range testData {
		if resultData[5+i] != val {
			t.Errorf("位置%d的数据错误，预期%d，实际%d", 5+i, val, resultData[5+i])
		}
	}
}

// 测试SaveChanges函数（集成测试）
func TestSaveChanges(t *testing.T) {
	// 检查测试数据文件是否存在
	testFilePath := "../../data/Save4.dat"
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("测试数据文件不存在，跳过集成测试")
	}
	
	// 创建目标文件路径
	destFilePath := filepath.Join(t.TempDir(), "Save_modified.dat")
	
	// 准备测试数据
	characters := []models.CharacterInfo{
		{
			Name: "Test",
			Data: models.CharacterData{Strength: 10},
			RawBytes: models.RawByteData{
				Strength: []byte{0xA, 0x0}, // 10的小端字节序
			},
			Position: 0, // 这里应该是实际位置，为了测试使用0
		},
	}
	
	moneyInfo := models.MoneyInfo{
		Value:    1000,
		RawBytes: make([]byte, 4),
		Position: 203054,
	}
	binary.LittleEndian.PutUint32(moneyInfo.RawBytes, 1000)
	
	// 测试保存修改
	err := SaveChanges(testFilePath, destFilePath, characters, moneyInfo)
	if err != nil {
		t.Fatalf("SaveChanges失败: %v", err)
	}
	
	// 验证目标文件是否创建成功
	if _, err := os.Stat(destFilePath); os.IsNotExist(err) {
		t.Fatal("目标文件未创建")
	}
}