package writer

import (
	"encoding/binary"
	"io"
	"os"

	"wceditor/wcsave/models"
)

// writeToFilePosition 写入数据到文件指定位置
func writeToFilePosition(filePath string, position int64, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = file.Seek(position, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()
	var destination *os.File
	destination, err = os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destination.Close() }()

	var copyErr error
	_, copyErr = io.Copy(destination, source)
	return copyErr
}

// SaveChanges 保存修改到新文件
func SaveChanges(sourceFilePath, destinationFilePath string,
	characters []models.CharacterInfo, moneyInfo models.MoneyInfo) error {
	var err error

	// 复制源文件到目标文件（如果路径不同）
	if sourceFilePath != destinationFilePath {
		err = copyFile(sourceFilePath, destinationFilePath)
		if err != nil {
			return err
		}
	}

	// 保存银两修改
	if moneyInfo.Position != 0 && len(moneyInfo.RawBytes) > 0 {
		// 准备4字节的银两数据
		moneyBuffer := make([]byte, 4)
		binary.LittleEndian.PutUint32(moneyBuffer, uint32(moneyInfo.Value))

		err = writeToFilePosition(destinationFilePath, moneyInfo.Position, moneyBuffer)
		if err != nil {
			return err
		}
	}

	// 保存每个角色的属性修改
	for _, char := range characters {
		// 当前经验值
		buffer := make([]byte, 4)
		binary.LittleEndian.PutUint32(buffer, uint32(char.Data.CurrentExp))
		err = writeToFilePosition(destinationFilePath, char.Position+8, buffer)
		if err != nil {
			return err
		}

		// 升级经验值
		binary.LittleEndian.PutUint32(buffer, uint32(char.Data.NextLevelExp))
		err = writeToFilePosition(destinationFilePath, char.Position+12, buffer)
		if err != nil {
			return err
		}

		// 最大生命值
		binary.LittleEndian.PutUint32(buffer, uint32(char.Data.MaxHP))
		err = writeToFilePosition(destinationFilePath, char.Position+16, buffer)
		if err != nil {
			return err
		}

		// 最大内力值
		binary.LittleEndian.PutUint32(buffer, uint32(char.Data.MaxMP))
		err = writeToFilePosition(destinationFilePath, char.Position+20, buffer)
		if err != nil {
			return err
		}

		// 当前生命值
		binary.LittleEndian.PutUint32(buffer, uint32(char.Data.CurrentHP))
		err = writeToFilePosition(destinationFilePath, char.Position+24, buffer)
		if err != nil {
			return err
		}

		// 当前内力值
		binary.LittleEndian.PutUint32(buffer, uint32(char.Data.CurrentMP))
		err = writeToFilePosition(destinationFilePath, char.Position+28, buffer)
		if err != nil {
			return err
		}

		// 2字节属性
		buffer16 := make([]byte, 2)

		// 力量
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Strength))
		err = writeToFilePosition(destinationFilePath, char.Position+32, buffer16)
		if err != nil {
			return err
		}

		// 反应
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Reaction))
		err = writeToFilePosition(destinationFilePath, char.Position+34, buffer16)
		if err != nil {
			return err
		}

		// 体质
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Constitution))
		err = writeToFilePosition(destinationFilePath, char.Position+36, buffer16)
		if err != nil {
			return err
		}

		// 速度
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Speed))
		err = writeToFilePosition(destinationFilePath, char.Position+38, buffer16)
		if err != nil {
			return err
		}

		// 攻击
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Attack))
		err = writeToFilePosition(destinationFilePath, char.Position+40, buffer16)
		if err != nil {
			return err
		}

		// 防御
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Defense))
		err = writeToFilePosition(destinationFilePath, char.Position+42, buffer16)
		if err != nil {
			return err
		}

		// 运气
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Luck))
		err = writeToFilePosition(destinationFilePath, char.Position+52, buffer16)
		if err != nil {
			return err
		}

		// 等级
		binary.LittleEndian.PutUint16(buffer16, uint16(char.Data.Level))
		err = writeToFilePosition(destinationFilePath, char.Position+70, buffer16)
		if err != nil {
			return err
		}
	}

	return nil
}
