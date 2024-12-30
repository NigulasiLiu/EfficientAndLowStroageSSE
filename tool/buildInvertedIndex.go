package tool

import (
	"EfficientAndLowStroageSSE/config"
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// PreprocessResult 表示预处理结果
type PreprocessResult struct {
	TotalLines      int // 数据总行数
	UniqueUsers     int // 唯一用户数
	UniqueLocations int // 唯一位置数
	UniqueKeywords  int // 唯一纬度数量（latitude）
}

// roundToFourDecimalPlaces 四舍五入 float64 到小数点后四位
func roundToFourDecimalPlaces_string(value float64) string {
	return fmt.Sprintf("%.4f", value)
}

// roundToFourDecimalPlaces 将浮点数四舍五入到小数点后 4 位
func roundToFourDecimalPlaces(value string) (string, error) {
	latitude, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", fmt.Errorf("invalid latitude value: %s", value)
	}
	rounded := math.Round(latitude*config.Divide) / config.Divide
	return fmt.Sprintf("%.4f", rounded), nil
}

// roundToFourDecimalPlaces 将一个字符串表示的浮点数四舍五入到四位小数
func roundToFourDecimalPlaces2(value string) (float64, error) {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	// 使用 math.Round 将值四舍五入到四位小数
	roundedValue := math.Round(floatVal*10000) / 10000.0
	return roundedValue, nil
}

// PreprocessData 解析 Gowalla_totalCheckins.txt 文件并统计所需数据
func PreprocessData(filePath string) (*PreprocessResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	userSet := make(map[string]struct{})
	locationSet := make(map[string]struct{})
	keywordSet := make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	totalLines := 0

	for scanner.Scan() {
		line := scanner.Text()
		totalLines++

		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			continue
		}

		user := fields[0]
		location := fields[4]
		latitude := fields[2]

		roundedLatitude, err := roundToFourDecimalPlaces(latitude)
		if err != nil {
			continue
		}

		userSet[user] = struct{}{}
		locationSet[location] = struct{}{}
		keywordSet[roundedLatitude] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &PreprocessResult{
		TotalLines:      totalLines,
		UniqueUsers:     len(userSet),
		UniqueLocations: len(locationSet),
		UniqueKeywords:  len(keywordSet),
	}, nil
}

// InvertedIndex 倒排索引结构
type InvertedIndex map[string][]int // keyword -> list of row IDs
func BuildInvertedIndex(filePath string) (InvertedIndex, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建倒排索引
	invertedIndex := make(InvertedIndex)
	scanner := bufio.NewScanner(file)
	rowID := 0

	// 扫描每一行
	for scanner.Scan() {
		line := scanner.Text()
		rowID++

		// 分割行数据，获取 latitude（作为 keyword）和 rowID
		fields := strings.Split(line, ",")
		if len(fields) != 2 {
			continue // 如果一行数据格式不对，跳过
		}

		latitude := fields[0] // 第一个字段是 latitude（keyword）

		// 直接使用 latitude（作为 keyword）
		// 因为已经四舍五入过，所以直接将其作为字符串使用

		// 将 keyword 和 id 加入倒排索引
		invertedIndex[latitude] = append(invertedIndex[latitude], rowID-1)
	}

	// 错误处理
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return invertedIndex, nil
}
