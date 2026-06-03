package utils

import (
	"encoding/binary"
	"io"
	"os"

	"wceditor/wcsave/models"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// SkipBytes 跳过n字节
func SkipBytes(file *os.File, n int) error {
	_, err := file.Seek(int64(n), 1) // 从当前位置跳过n字节
	if err != nil {
		// 如果Seek失败，尝试通过读取来跳过
		buffer := make([]byte, n)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
	}
	return nil
}

// Converter 泛型转换函数类型
type Converter[T any] func([]byte) (T, error)

// ReadAndConvert 泛型读取方法，支持自定义转换函数
func ReadAndConvert[T any](file *os.File, n int, converter Converter[T]) (T, []byte, error) {
	var zero T
	buffer := make([]byte, n)
	readCount, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return zero, nil, err
	}

	rawBytes := make([]byte, readCount)
	copy(rawBytes, buffer[:readCount])

	// 如果converter为nil，直接返回零值
	if converter == nil {
		return zero, rawBytes, nil
	}

	// 调用converter进行转换
	value, err := converter(buffer[:readCount])
	if err != nil {
		return zero, rawBytes, err
	}

	return value, rawBytes, nil
}

// Big5ToUTF8Converter 将 Big5 编码字节转换为 UTF-8 字节。
// 转换失败时返回原始字节，保持向后兼容。
func Big5ToUTF8Converter(b []byte) ([]byte, error) {
	utf8Bytes, err := traditionalchinese.Big5.NewDecoder().Bytes(b)
	if err != nil {
		return b, nil
	}
	return utf8Bytes, nil
}

// GBKToUTF8Converter 将 GBK 编码字节转换为 UTF-8 字节。
// 转换失败时返回原始字节，保持向后兼容。
func GBKToUTF8Converter(b []byte) ([]byte, error) {
	utf8Bytes, err := simplifiedchinese.GBK.NewDecoder().Bytes(b)
	if err != nil {
		return b, nil
	}
	return utf8Bytes, nil
}

// ToUTF8Converter 按指定字符集将字节转换为 UTF-8。
// 简体版使用 GBK，繁体版使用 Big5。
func ToUTF8Converter(charset models.Charset, b []byte) ([]byte, error) {
	switch charset {
	case models.CharsetTraditional:
		return Big5ToUTF8Converter(b)
	case models.CharsetSimplified:
		return GBKToUTF8Converter(b)
	default:
		// 兼容未知值，默认按繁体版处理
		return Big5ToUTF8Converter(b)
	}
}

// NewUTF8Converter 根据字符集返回一个 UTF-8 转换器。
func NewUTF8Converter(charset models.Charset) Converter[[]byte] {
	return func(b []byte) ([]byte, error) {
		return ToUTF8Converter(charset, b)
	}
}

// Int32LittleEndianConverter 将 4 字节小端数据转换为 int32。
func Int32LittleEndianConverter(b []byte) (int32, error) {
	return int32(binary.LittleEndian.Uint32(b)), nil
}

// Int16LittleEndianConverter 将 2 字节小端数据转换为 int16。
func Int16LittleEndianConverter(b []byte) (int16, error) {
	return int16(binary.LittleEndian.Uint16(b)), nil
}
