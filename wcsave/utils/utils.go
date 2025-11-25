package utils

import (
	"io"
	"os"
)

// skipBytes 跳过n字节
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