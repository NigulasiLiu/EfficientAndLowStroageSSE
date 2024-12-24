package tool

import (
	"VolumeHidingSSE/config"
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestPreprocessData 测试 PreprocessData 函数
func TestPreprocessData(t *testing.T) {
	filePath := config.FilePath

	result, err := PreprocessData(filePath)
	if err != nil {
		t.Fatalf("PreprocessData failed: %v", err)
	}

	t.Logf("数据总行数: %d", result.TotalLines)
	t.Logf("唯一keywords数: %d", result.UniqueKeywords)

	if result.TotalLines == 0 {
		t.Errorf("TotalLines failed: expected non-zero, got %d", result.TotalLines)
	}
	if result.UniqueUsers == 0 {
		t.Errorf("UniqueUsers failed: expected non-zero, got %d", result.UniqueUsers)
	}
	if result.UniqueLocations == 0 {
		t.Errorf("UniqueLocations failed: expected non-zero, got %d", result.UniqueLocations)
	}
	if result.UniqueKeywords == 0 {
		t.Errorf("UniqueKeywords failed: expected non-zero, got %d", result.UniqueKeywords)
	}
}

// TestConvertFile 读取源文件，将 latitude 四舍五入到小数点后四位,然后去除小数点，输出到 invertedindex.csv 文件
func TestConvertFile(t *testing.T) {
	// 打开源文件
	filePath := config.FilePath_raw
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("无法打开文件 %s: %v", filePath, err)
	}
	defer file.Close()

	// 创建输出文件
	outputFile, err := os.Create("C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\origin.csv")
	if err != nil {
		t.Fatalf("无法创建 output 文件: %v", err)
	}
	defer outputFile.Close()

	// 创建文件写入器
	writer := bufio.NewWriter(outputFile)

	// 扫描文件内容
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	// 写入 CSV 表头
	_, err = writer.WriteString("Keyword,ID\n")
	if err != nil {
		t.Fatalf("写入 CSV 表头失败: %v", err)
	}

	// 遍历每一行
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		// 分割行数据，获取 latitude（第三列）作为 keyword
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			continue // 如果一行数据格式不对，跳过
		}

		latitude := fields[2] // 第三列数据，即 latitude

		// 将 latitude 转换为 float64
		latitudeFloat, err := strconv.ParseFloat(latitude, 64)
		if err != nil {
			continue // 如果无法转换为 float64，则跳过这一行
		}

		// 四舍五入 latitude 到小数点后四位
		roundedLatitude := roundToFourDecimalPlaces_string(latitudeFloat)

		// 去除小数点
		roundedLatitudeNoDot := strings.Replace(roundedLatitude, ".", "", -1)

		// 将 keyword（去除小数点后的 latitude）和 id 写入新文件
		_, err = writer.WriteString(fmt.Sprintf("%s,%d\n", roundedLatitudeNoDot, lineNumber-1))
		if err != nil {
			t.Fatalf("写入文件失败: %v", err)
		}
	}

	// 检查文件扫描错误
	if err := scanner.Err(); err != nil {
		t.Fatalf("扫描文件时发生错误: %v", err)
	}

	// 确保所有内容都被写入文件
	writer.Flush()

	// 打印成功日志
	t.Logf("已将 %d 行数据转换并写入到 origin.csv 文件", lineNumber)
}

// TestBuildInvertedIndex 测试 BuildInvertedIndex 函数并将倒排索引写入 CSV 文件
func TestBuildInvertedIndex(t *testing.T) {
	// 使用实际文件路径测试
	filePath := config.FilePath

	// 调用 BuildInvertedIndex 函数
	invertedIndex, err := BuildInvertedIndex(filePath)
	if err != nil {
		t.Fatalf("BuildInvertedIndex failed: %v", err)
	}

	// 打印倒排索引的数量
	totalKeywords := len(invertedIndex)
	t.Logf("倒排索引数量: %d", totalKeywords)

	// 创建 CSV 文件
	csvFile, err := os.Create("C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\InvertedIndex.csv")
	if err != nil {
		t.Fatalf("无法创建 CSV 文件: %v", err)
	}
	defer csvFile.Close()

	// 创建 CSV 写入器
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// 写入 CSV 表头
	csvWriter.Write([]string{"Keyword", "Row IDs"})

	// 将倒排索引写入 CSV 文件
	for keyword, rowIDs := range invertedIndex {
		// 将 rowIDs 转换为字符串，格式： rowID1, rowID2, ...
		rowIDsStr := fmt.Sprintf("%v", rowIDs)

		// 写入每个倒排索引的 keyword 和 rowIDs 到文件
		err := csvWriter.Write([]string{keyword, rowIDsStr})
		if err != nil {
			t.Fatalf("写入 CSV 文件失败: %v", err)
		}
	}

	// 打印倒排索引的前 20 个关键词和 Row IDs（如果总数大于 20）
	if totalKeywords <= 20 {
		for keyword, rows := range invertedIndex {
			t.Logf("Keyword: %s, Row IDs: %v", keyword, rows)
		}
		return
	}

	// 随机挑选 20 个倒排索引进行打印
	rand.Seed(time.Now().UnixNano())
	randomKeys := getRandomKeys(invertedIndex, 20)

	for _, keyword := range randomKeys {
		t.Logf("Keyword: %s, Row IDs: %v", keyword, invertedIndex[keyword])
	}
}

// getRandomKeys 从倒排索引中随机挑选指定数量的关键字
func getRandomKeys(invertedIndex map[string][]int, count int) []string {
	keys := make([]string, 0, len(invertedIndex))
	for key := range invertedIndex {
		keys = append(keys, key)
	}

	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	if len(keys) < count {
		return keys // 如果关键字数量少于所需数量，直接返回所有关键字
	}
	return keys[:count]
}

// 获取invertedIndex文件中的统计信息,第一行无效
func TestReadCSVAnalyze(t *testing.T) {
	// 打开CSV文件
	csvFile, err := os.Open("../dataset/InvertedIndex.csv")
	if err != nil {
		t.Fatalf("无法打开 CSV 文件: %v", err)
	}
	defer csvFile.Close()

	// 创建CSV读取器
	csvReader := csv.NewReader(csvFile)

	// 读取所有行
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("无法读取 CSV 文件内容: %v", err)
	}

	// 变量初始化
	var (
		minKeyword    float64
		maxKeyword    float64
		keywordExists bool
		duplicateIDs  map[string]map[int]bool // 存储重复的 Row IDs
	)

	duplicateIDs = make(map[string]map[int]bool)

	// 遍历 CSV 文件中的每一行，跳过表头
	for i, record := range records {
		// 跳过表头
		if i == 0 {
			continue
		}

		// 获取 Keyword 和 Row IDs 列
		keyword := record[0]
		rowIDsStr := record[1]

		// 将 Row IDs 从字符串转换为整数数组
		rowIDs := parseRowIDs(rowIDsStr)

		// 检查 Keyword 是否为有效数字
		keywordValue, err := strconv.ParseFloat(keyword, 64)
		if err != nil {
			t.Errorf("无效的 Keyword 值: %v", keyword)
			continue
		}

		// 初始化 min 和 max
		if !keywordExists {
			minKeyword = keywordValue
			maxKeyword = keywordValue
			keywordExists = true
		}

		// 更新 min 和 max
		if keywordValue < minKeyword {
			minKeyword = keywordValue
		}
		if keywordValue > maxKeyword {
			maxKeyword = keywordValue
		}

		// 检查是否有重复的 Row IDs
		if hasDuplicates(rowIDs) {
			duplicateIDs[keyword] = findDuplicates(rowIDs)
		}
	}

	// 输出最大值和最小值
	t.Logf("Keyword 最大值: %f", maxKeyword)
	t.Logf("Keyword 最小值: %f", minKeyword)

	// 打印重复的 Row IDs
	if len(duplicateIDs) > 0 {
		t.Log("以下 Keyword 对应的 Row IDs 存在重复：")
		for keyword, rows := range duplicateIDs {
			t.Logf("Keyword: %s, 重复的 Row IDs: %v", keyword, rows)
		}
	} else {
		t.Log("没有发现重复的 Row IDs。")
	}
}

// 解析 Row IDs 字符串为整数数组
func parseRowIDs(rowIDsStr string) []int {
	// 去除前后空格，并去除两边的方括号
	rowIDsStr = strings.TrimSpace(rowIDsStr)
	rowIDsStr = strings.Trim(rowIDsStr, "[]")

	// 将字符串拆分为多个 ID
	rowIDStrings := strings.Split(rowIDsStr, " ")

	// 将字符串转换为整数
	var rowIDs []int
	for _, idStr := range rowIDStrings {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			rowIDs = append(rowIDs, id)
		}
	}

	return rowIDs
}

// 检查是否有重复的 Row IDs
func hasDuplicates(rowIDs []int) bool {
	seen := make(map[int]bool)
	for _, id := range rowIDs {
		if seen[id] {
			return true
		}
		seen[id] = true
	}
	return false
}

// 找到重复的 Row IDs
func findDuplicates(rowIDs []int) map[int]bool {
	seen := make(map[int]bool)
	duplicates := make(map[int]bool)

	for _, id := range rowIDs {
		if seen[id] {
			duplicates[id] = true
		} else {
			seen[id] = true
		}
	}

	return duplicates
}
