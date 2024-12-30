package OurScheme

import (
	"EfficientAndLowStroageSSE/config"
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func sortKeywords(invertedIndex map[string][]int) []string {
	// 获取所有关键词
	keywords := make([]string, 0, len(invertedIndex))
	for keyword := range invertedIndex {
		keywords = append(keywords, keyword)
	}

	// 按照关键词的数值顺序排序
	sort.Slice(keywords, func(i, j int) bool {
		ki, _ := strconv.ParseInt(keywords[i], 10, 64)
		kj, _ := strconv.ParseInt(keywords[j], 10, 64)
		return ki < kj
	})

	//log.Printf("min: %s, max: %s", keywords[0], keywords[len(invertedIndex)-1])
	return keywords
}
func TestPerformanceAdvanced_valid(t *testing.T) {
	// 文件列表
	files := []string{
		config.FilePath_txt,
		// 其他文件路径...
	}

	// 参数范围
	LValues := []int{6424, 6424 * 2, 6424 * 3, 6424 * 4, 6424 * 5}
	k := 999999 // 设置最大查询次数
	resultCounts := 500
	// 结果存储目录
	resultsDir := "results"

	// 遍历每个文件
	for fileIndex, file := range files {
		// 加载倒排索引
		invertedIndex, err := loadInvertedIndex(file)
		if err != nil {
			t.Fatalf("无法加载文件 %s: %v", file, err)
		}

		// 提取文件中的 keywords
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			t.Fatalf("文件 %s 中未找到关键词", file)
		}

		// 测量关键词排序时间
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex) // 排序函数
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 Setup 参数
			sp := Setup(L)

			// 测量 BuildIndex 时间
			startTime := time.Now()
			err = sp.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("BuildIndex 返回错误: %v", err)
			}
			buildIndexDuration := time.Since(startTime).Nanoseconds()

			// 结果文件路径（CSV格式）
			resultFilePath := fmt.Sprintf("%s/result_m_%d_L_%d.csv", resultsDir, fileIndex+1, L)
			resultFile, err := os.Create(resultFilePath)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePath, err)
			}
			defer resultFile.Close()

			// 写入结果文件表头
			writer := csv.NewWriter(resultFile)
			defer writer.Flush()
			writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)"})

			// 初始化有效查询计数器
			validCount := 0

			// 开始测试
			for i := 0; i < k; i++ {
				// 生成查询区间并计算区间宽度
				queryRange, rangeWidth := generateQueryRangeWithWidth(sortedKeywords)

				// 测量 GenToken 时间
				startTime = time.Now()
				tokens, err := sp.GenToken(queryRange)
				if err != nil {
					t.Fatalf("GenToken 返回错误: %v", err)
				}
				genTokenDuration := time.Since(startTime).Nanoseconds()

				// 如果 tokens 为空，设置后续耗时为 0 并跳过
				if len(tokens) == 0 {
					searchTokensDuration := 0
					localSearchDuration := 0
					clientTimeCost := searchTokensDuration + localSearchDuration
					writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDuration), fmt.Sprintf("%d", genTokenDuration), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
					t.Logf("Tokens are empty, skipping iteration")
					continue
				}

				// 计数有效查询
				validCount++

				// 测量 SearchTokens 时间
				startTime = time.Now()
				searchResult := sp.SearchTokens(tokens)
				searchTokensDuration := time.Since(startTime).Nanoseconds()

				// 测量 LocalSearch 时间
				startTime = time.Now()
				_, err = sp.LocalSearch(searchResult, tokens)
				if err != nil {
					t.Fatalf("LocalSearch 返回错误: %v", err)
				}
				localSearchDuration := time.Since(startTime).Nanoseconds()

				// 计算 ClientTimeCost
				clientTimeCost := genTokenDuration + localSearchDuration

				// 写入每次实验的耗时记录
				writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDuration), fmt.Sprintf("%d", genTokenDuration), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})

				// 如果有效查询次数达到 100 次，停止循环
				if validCount >= resultCounts {
					t.Logf("达到 100 次有效查询，停止测试")
					break
				}
			}

			// 打印完成信息
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePath)
		}
	}
}

func TestPerformanceAdvanced(t *testing.T) {
	// 文件列表
	files := []string{
		config.FilePath_txt,
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_1_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_2_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_3_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_4_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_5_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_6_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_7_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_8_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_9_d_10_combined.csv",
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_10_d_10_combined.csv",
	}

	// 参数范围
	LValues := []int{6424}
	k := 30

	// 结果存储目录
	resultsDir := "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\results"

	// 遍历每个文件
	for fileIndex, file := range files {
		// 加载倒排索引
		invertedIndex, err := loadInvertedIndex(file)
		if err != nil {
			t.Fatalf("无法加载文件 %s: %v", file, err)
		}

		// 提取文件中的 keywords
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			t.Fatalf("文件 %s 中未找到关键词", file)
		}
		// 测量关键词排序时间
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex) // 排序函数
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 Setup 参数
			sp := Setup(L)

			// 测量 BuildIndex 时间
			startTime := time.Now()
			err = sp.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("BuildIndex 返回错误: %v", err)
			}
			buildIndexDuration := time.Since(startTime).Nanoseconds()

			// 结果文件路径（CSV格式）
			resultFilePath := fmt.Sprintf("%s/result_m_%d_L_%d.csv", resultsDir, fileIndex+1, L)
			resultFile, err := os.Create(resultFilePath)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePath, err)
			}
			defer resultFile.Close()

			// 写入结果文件表头
			writer := csv.NewWriter(resultFile)
			defer writer.Flush()
			writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)"})

			// 开始测试
			for i := 0; i < k; i++ {
				// 生成查询区间并计算区间宽度
				queryRange, rangeWidth := generateQueryRangeWithWidth(sortedKeywords)

				// 测量 GenToken 时间
				startTime = time.Now()
				tokens, err := sp.GenToken(queryRange)
				if err != nil {
					t.Fatalf("GenToken 返回错误: %v", err)
				}
				genTokenDuration := time.Since(startTime).Nanoseconds()

				// 如果 tokens 为空，设置后续耗时为 0 并跳过
				if len(tokens) == 0 {
					searchTokensDuration := 0
					localSearchDuration := 0
					clientTimeCost := searchTokensDuration + localSearchDuration
					writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDuration), fmt.Sprintf("%d", genTokenDuration), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
					t.Logf("Tokens are empty, skipping iteration")
					continue
				}

				// 测量 SearchTokens 时间
				startTime = time.Now()
				searchResult := sp.SearchTokens(tokens)
				searchTokensDuration := time.Since(startTime).Nanoseconds()

				// 测量 LocalSearch 时间
				startTime = time.Now()
				_, err = sp.LocalSearch(searchResult, tokens)
				if err != nil {
					t.Fatalf("LocalSearch 返回错误: %v", err)
				}
				localSearchDuration := time.Since(startTime).Nanoseconds()

				// 计算 ClientTimeCost
				clientTimeCost := genTokenDuration + localSearchDuration

				// 写入每次实验的耗时记录
				writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDuration), fmt.Sprintf("%d", genTokenDuration), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
			}

			// 打印完成信息
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePath)
		}
	}
}

func generateQueryRangeWithWidth(keywords []string) ([2]string, int) {
	n := len(keywords)
	if n < 2 {
		return [2]string{"0", "0"}, 0
	}

	// 随机选择 i 和 j (确保 j > i)
	i := rand.Intn(n)
	j := rand.Intn(n-i-1) + i + 1 // 生成 j > i

	// 获取第 i 和 j 个关键词作为查询区间的左右边界
	left := keywords[i]
	right := keywords[j]

	// 转为整数并计算区间宽度
	leftInt, _ := strconv.Atoi(left)
	rightInt, _ := strconv.Atoi(right)
	rangeWidth := rightInt - leftInt

	return [2]string{left, right}, rangeWidth
}

// extractKeywordsFromIndex 提取文件中的所有关键词
func extractKeywordsFromIndex(invertedIndex map[string][]int) []int {
	keywords := []int{}
	for k := range invertedIndex {
		key, err := strconv.Atoi(k)
		if err == nil {
			keywords = append(keywords, key)
		}
	}
	sort.Ints(keywords)
	return keywords
}

func TestLoadInvertedIndex_txt(t *testing.T) {
	filePath := "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\Gowalla_invertedIndex.txt"

	// 调用函数加载倒排索引
	invertedIndex, err := loadInvertedIndex(filePath)
	if err != nil {
		t.Fatalf("加载倒排索引失败: %v", err)
	}

	// 打印加载的倒排索引
	for key, rowIDs := range invertedIndex {
		t.Logf("Key: %s, Row IDs: %v", key, rowIDs)
	}
}

func TestLoadInvertedIndexDistribution(t *testing.T) {
	filePath := config.FilePath_txt

	// 加载倒排索引
	invertedIndex, err := loadInvertedIndex(filePath)
	if err != nil {
		t.Fatalf("加载倒排索引失败: %v", err)
	}

	// 提取并统计 convertedKey 的分布
	distribution := analyzeKeyDistribution(invertedIndex, -900000, 900000, 10000)

	// 打印分布结果
	fmt.Println("ConvertedKey 分布情况：")
	for rangeStart, count := range distribution {
		fmt.Printf("区间 [%d, %d): %d\n", rangeStart, rangeStart+10000, count)
	}
}

// analyzeKeyDistribution 统计 convertedKey 在 [-900000, 900000] 范围上的分布
func analyzeKeyDistribution(invertedIndex map[string][]int, minRange, maxRange, interval int) map[int]int {
	distribution := make(map[int]int)

	for key := range invertedIndex {
		// 将 key 转换为整数
		keyInt, err := strconv.Atoi(key)
		if err != nil {
			fmt.Printf("警告: 无法解析 key '%s' 为整数，跳过\n", key)
			continue
		}

		// 检查是否在范围内
		if keyInt >= minRange && keyInt < maxRange {
			// 确定区间起点
			rangeStart := (keyInt / interval) * interval
			if keyInt < 0 && keyInt%interval != 0 {
				rangeStart -= interval
			}
			distribution[rangeStart]++
		}
	}

	return distribution
}

func loadInvertedIndex_float(filePath string) (map[string][]int, error) {
	invertedIndex := make(map[string][]int)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // 去掉行首尾的空白字符
		if line == "" {
			continue // 跳过空行
		}

		// 按空格分割每行
		parts := strings.Fields(line)
		if len(parts) < 2 { // 至少需要一列 key 和一列值
			return nil, fmt.Errorf("无效的行格式: %s", line)
		}

		// 解析关键词（第一列）
		keyword := parts[0]
		floatKeyword, err := strconv.ParseFloat(keyword, 64) // 转换为浮点数
		if err != nil {
			fmt.Printf("警告: 无法解析关键词 '%s' 为浮点数，跳过此行\n", keyword)
			continue
		}

		// 乘以 10000，并将其转为 string
		convertedKey := strconv.FormatInt(int64(floatKeyword*10000), 10)

		// 解析行号（第二列及后续列）
		rowIDs := []int{}
		for _, part := range parts[1:] {
			rowID, err := strconv.Atoi(part) // 转换为整数
			if err != nil {
				fmt.Printf("警告: 无法解析行号 '%s'，跳过该值\n", part)
				continue
			}
			rowIDs = append(rowIDs, rowID)
		}

		// 将关键词和对应的行号存入倒排索引
		invertedIndex[convertedKey] = rowIDs
	}

	// 检查文件读取是否出错
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件出错: %v", err)
	}

	return invertedIndex, nil
}
func loadInvertedIndex(filePath string) (map[string][]int, error) {
	invertedIndex := make(map[string][]int)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // 去掉行首尾的空白字符
		if line == "" {
			continue // 跳过空行
		}

		// 按空格分割每行
		parts := strings.Fields(line)
		if len(parts) < 2 { // 至少需要一列 key 和一列值
			return nil, fmt.Errorf("无效的行格式: %s", line)
		}

		// 第一列作为关键词（key）
		keyword := parts[0]

		// 第二列及后续列作为行号（Row IDs）
		rowIDs := []int{}
		for _, part := range parts[1:] {
			rowID, err := strconv.Atoi(part) // 转换为整数
			if err != nil {
				fmt.Printf("警告: 无法解析行号 '%s'，跳过该值\n", part)
				continue
			}
			rowIDs = append(rowIDs, rowID)
		}

		// 将关键词和对应的行号存入倒排索引
		invertedIndex[keyword] = rowIDs
	}

	// 检查文件读取是否出错
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件出错: %v", err)
	}

	return invertedIndex, nil
}

// loadInvertedIndex_csv 从 CSV 文件中加载倒排索引
func loadInvertedIndex_csv(filePath string) (map[string][]int, error) {
	invertedIndex := make(map[string][]int)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 创建 CSV 读取器
	reader := csv.NewReader(file)

	// 跳过第一行（标题行）
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("无法读取标题行: %v", err)
	}

	// 逐行读取 CSV 内容
	for {
		record, err := reader.Read()
		if err != nil {
			break // 读取完成
		}

		if len(record) < 2 {
			return nil, fmt.Errorf("无效的 CSV 行: %v", record)
		}

		// 解析关键词
		keyword := record[0]

		// 解析 Row IDs
		rowIDs := []int{}
		for _, rowIDStr := range record[1:] { // 从第二列开始是 Row IDs
			rowIDStr = strings.TrimSpace(rowIDStr) // 去掉多余的空格
			if rowIDStr == "" {
				continue // 跳过空值
			}

			// 将 Row ID 转换为整数
			rowID, err := strconv.Atoi(rowIDStr)
			if err != nil {
				// 记录警告信息，跳过当前行，继续处理下一个 Row ID
				//fmt.Printf("警告: 无法解析行号 '%s'，跳过该行\n", rowIDStr)
				continue // 跳过当前行，继续下一行
			}

			// 将行号加入 rowIDs
			rowIDs = append(rowIDs, rowID)
		}

		// 更新倒排索引
		invertedIndex[keyword] = rowIDs
	}

	return invertedIndex, nil
}

// TestCombineCSVFile 合并多个 CSV 文件，生成新的 CSV 文件
func TestCombineCSVFile(t *testing.T) {
	// 文件列表
	files := []string{
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_1_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_2_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_3_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_4_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_5_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_6_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_7_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_8_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_9_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_10_d_10.csv",
	}

	// 遍历 1 到 10 个文件进行合并
	for i := 1; i <= 10; i++ {
		// 打开输出文件
		outputFile := fmt.Sprintf("C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_%d_d_10_combined.csv", i)
		outFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Printf("无法创建文件 %s: %v\n", outputFile, err)
			continue
		}
		defer outFile.Close()

		// 创建 CSV 写入器
		writer := csv.NewWriter(outFile)
		defer writer.Flush()

		// 存储表头
		var header []string
		isFirstFile := true

		// 合并前 i 个文件
		for j := 0; j < i; j++ {
			filePath := files[j]
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("无法打开文件 %s: %v\n", filePath, err)
				continue
			}
			defer file.Close()

			// 创建 CSV 读取器
			reader := csv.NewReader(file)

			// 读取文件内容并写入到新文件中
			for {
				record, err := reader.Read()
				if err != nil {
					break // 读取完成
				}

				// 如果是第一个文件，保存表头
				if isFirstFile {
					header = record
					isFirstFile = false
					writer.Write(header) // 写入表头
				} else {
					writer.Write(record) // 写入数据
				}
			}
		}

		fmt.Printf("已将前 %d 个文件合并并写入到文件: %s\n", i, outputFile)
	}
}

// TestPerformance 测试搜索性能
func TestPerformance(t *testing.T) {
	// 初始化 Setup 参数
	L := 6424

	// 初始化查询范围的上下界
	minKey := 1
	maxKey := 5
	k := 10 // 每次随机生成 10 个查询范围
	// 文件列表
	files := []string{
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_1.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_2.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_3.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_4.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_5.csv",
	}

	// 随机选择一个文件
	rand.Seed(time.Now().UnixNano())
	selectedFile := files[rand.Intn(len(files))]
	fmt.Printf("随机选择文件: %s\n", selectedFile)

	// 加载文件内容到 invertedIndex
	invertedIndex, err := loadInvertedIndex_csv(selectedFile)
	if err != nil {
		t.Fatalf("无法加载文件 %s: %v", selectedFile, err)
	}

	sp := Setup(L)

	// 测试性能
	for i := 0; i < k; i++ {
		// 随机生成查询范围
		startKey := rand.Intn(maxKey-minKey+1) + minKey
		endKey := rand.Intn(maxKey-startKey+1) + startKey

		queryRange := [2]string{strconv.Itoa(startKey), strconv.Itoa(endKey)}
		t.Logf("查询范围 %d: %v", i+1, queryRange)
		// 测量关键词排序时间
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex) // 排序函数
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
		// 测量 BuildIndex 时间
		startTime = time.Now()
		err = sp.BuildIndex(invertedIndex, sortedKeywords)
		if err != nil {
			t.Fatalf("BuildIndex 返回错误: %v", err)
		}
		buildIndexDuration := time.Since(startTime).Seconds() * 1000 // 毫秒
		t.Logf("BuildIndex 耗时: %.4f ms", buildIndexDuration)
		startTime = time.Now()
		tokens, err := sp.GenToken(queryRange)
		if err != nil {
			t.Fatalf("GenToken 返回错误: %v", err)
		}

		// 如果 tokens 为空，直接返回结果为空
		if len(tokens) == 0 {
			t.Logf("Tokens are empty, returning empty result")
			return
		}

		genTokenDuration := time.Since(startTime).Nanoseconds() // 纳秒
		t.Logf("GenToken 耗时: %d ns", genTokenDuration)

		// 测量 SearchTokens 时间
		startTime = time.Now()
		searchResult := sp.SearchTokens(tokens)
		searchTokensDuration := time.Since(startTime).Nanoseconds() // 纳秒
		t.Logf("SearchTokens 耗时: %d ns", searchTokensDuration)

		// 测量 LocalSearch 时间
		startTime = time.Now()
		actualSearchResult, err := sp.LocalSearch(searchResult, tokens)
		if err != nil {
			t.Fatalf("LocalSearch 返回错误: %v", err)
		}
		localSearchDuration := time.Since(startTime).Nanoseconds() // 纳秒
		t.Logf("LocalSearch 耗时: %d ns", localSearchDuration)
		maxResults := 5
		if len(actualSearchResult) > maxResults {
			fmt.Printf("查询范围 %d: %v, 解密结果（前 %d 个）: %v\n", i+1, queryRange, maxResults, actualSearchResult[:maxResults])
		} else {
			fmt.Printf("查询范围 %d: %v, 解密结果: %v\n", i+1, queryRange, actualSearchResult)
		}

	}
}

// TestPerformanceAdvanced 测试性能
func TestPerformanceAdvanced_Fix_Width(t *testing.T) {
	// 文件列表
	files := []string{
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_1.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_2.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_3.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_4.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\DB_1288579_m_5.csv",
	}
	useBoundary := true
	// 参数范围
	LValues := []int{6424, 6424 * 2, 6424 * 3, 6424 * 4, 6424 * 5}
	rangeValues := []int{1, 10, 100, 1000, 10000, 100000}
	k := 10

	// 结果存储目录
	resultsDir := "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\results"

	// 遍历每个文件
	for fileIndex, file := range files {
		// 加载倒排索引
		invertedIndex, err := loadInvertedIndex_csv(file)
		if err != nil {
			t.Fatalf("无法加载文件 %s: %v", file, err)
		}

		// 提取文件中的 keywords
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			t.Fatalf("文件 %s 中未找到关键词", file)
		}

		// 遍历每个 L 值
		for _, L := range LValues {
			// 遍历每个查询范围
			for _, rangeValue := range rangeValues {
				// 初始化 Setup 参数
				sp := Setup(L)

				// 结果文件路径
				resultFilePath := fmt.Sprintf("%s/result_m_%d_L_%d_range_%d.txt", resultsDir, fileIndex+1, L, rangeValue)
				resultFile, err := os.Create(resultFilePath)
				if err != nil {
					t.Fatalf("无法创建结果文件 %s: %v", resultFilePath, err)
				}
				defer resultFile.Close()

				// 写入结果文件表头
				writer := bufio.NewWriter(resultFile)
				writer.WriteString("Iteration,BuildIndex(ns),GenToken(ns),SearchTokens(ns),LocalSearch(ns)\n")

				// 开始测试
				for i := 0; i < k; i++ {
					// 随机生成查询范围
					queryRange := generateQueryRange_random_but_fixWidth(keywords, rangeValue, useBoundary)
					t.Logf("Query Range:%s", queryRange)
					// 测量关键词排序时间
					startTime := time.Now()
					sortedKeywords := sortKeywords(invertedIndex) // 排序函数
					sortDuration := time.Since(startTime).Nanoseconds()
					t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
					// 测量 BuildIndex 时间
					startTime = time.Now()
					err = sp.BuildIndex(invertedIndex, sortedKeywords)
					if err != nil {
						t.Fatalf("BuildIndex 返回错误: %v", err)
					}
					buildIndexDuration := time.Since(startTime).Nanoseconds()

					// 测量 GenToken 时间
					startTime = time.Now()
					tokens, err := sp.GenToken(queryRange)
					if err != nil {
						t.Fatalf("GenToken 返回错误: %v", err)
					}
					genTokenDuration := time.Since(startTime).Nanoseconds()

					// 如果 tokens 为空，设置后续耗时为 0 并跳过
					if len(tokens) == 0 {
						searchTokensDuration := 0
						localSearchDuration := 0
						writer.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n", i+1, buildIndexDuration, genTokenDuration, searchTokensDuration, localSearchDuration))
						t.Logf("Tokens are empty, skipping iteration")
						continue
					}

					// 测量 SearchTokens 时间
					startTime = time.Now()
					searchResult := sp.SearchTokens(tokens)
					searchTokensDuration := time.Since(startTime).Nanoseconds()

					// 测量 LocalSearch 时间
					startTime = time.Now()
					_, err = sp.LocalSearch(searchResult, tokens)
					if err != nil {
						t.Fatalf("LocalSearch 返回错误: %v", err)
					}
					localSearchDuration := time.Since(startTime).Nanoseconds()

					// 写入每次实验的耗时记录
					writer.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n", i+1, buildIndexDuration, genTokenDuration, searchTokensDuration, localSearchDuration))
				}

				writer.Flush()
				resultFile.Close()

				// 打印完成信息
				fmt.Printf("完成文件: %s, L: %d, range: %d, 结果存储于: %s\n", file, L, rangeValue, resultFilePath)
			}
		}
	}
}

// generateQueryRange 根据关键词和范围生成查询区间
func generateQueryRange_random_but_fixWidth(keywords []int, rangeValue int, useExactKeywords bool) [2]string {
	if useExactKeywords {
		// 完全选择 keywords 作为边界
		n := len(keywords)
		leftIndex := rand.Intn(n)
		rightIndex := min(leftIndex+rangeValue, n-1)
		return [2]string{
			strconv.Itoa(keywords[leftIndex]),
			strconv.Itoa(keywords[rightIndex]),
		}
	}

	// 保持随机选择机制
	n := len(keywords)
	leftIndex := rand.Intn(n)
	rightIndex := leftIndex

	// 三种情况生成查询范围
	caseType := rand.Intn(3)
	switch caseType {
	case 0: // 两个边界值都是关键词
		rightIndex = min(leftIndex+rangeValue, n-1)
	case 1: // 一个边界是关键词，另一个不是
		leftIndex = max(leftIndex-1, 0)
	case 2: // 两个边界都不是关键词
		leftIndex = max(leftIndex-1, 0)
		rightIndex = min(leftIndex+rangeValue, n-1)
	}

	return [2]string{
		strconv.Itoa(keywords[leftIndex]),
		strconv.Itoa(keywords[rightIndex]),
	}
}

func TestLoadInvertedIndex(t *testing.T) {
	// 定义文件路径
	filePath := "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_1_d_10.csv"

	// 调用函数加载倒排索引
	invertedIndex, err := loadInvertedIndex_csv(filePath)
	if err != nil {
		fmt.Printf("加载倒排索引失败: %v\n", err)
		return
	}

	// 打印倒排索引的前 100 行
	count := 0
	for keyword, rowIDs := range invertedIndex {
		count++
		if count > 100 {
			break
		}
		fmt.Printf("Keyword: %s, Row IDs: %v\n", keyword, rowIDs)
	}
}

func loadInvertedIndex_pre(filePath string) (map[string][]int, error) {
	invertedIndex := make(map[string][]int)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 创建 CSV 读取器
	reader := csv.NewReader(file)

	// 跳过第一行（标题行）
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("无法读取标题行: %v", err)
	}

	// 逐行读取 CSV 内容
	for {
		record, err := reader.Read()
		if err != nil {
			break // 读取完成
		}

		if len(record) < 2 {
			return nil, fmt.Errorf("无效的 CSV 行: %v", record)
		}

		// 解析关键词和行号
		keyword := record[0]
		rowIDsStr := record[1]
		rowIDsStr = strings.Trim(rowIDsStr, "[]")    // 去掉可能的方括号
		rowIDsSlice := strings.Split(rowIDsStr, ",") // 按逗号分割
		rowIDs := []int{}
		for _, rowIDStr := range rowIDsSlice {
			rowID, err := strconv.Atoi(strings.TrimSpace(rowIDStr))
			if err != nil {
				return nil, fmt.Errorf("无法解析行号: %s", rowIDStr)
			}
			rowIDs = append(rowIDs, rowID)
		}

		// 更新倒排索引
		invertedIndex[keyword] = rowIDs
	}

	return invertedIndex, nil
}

// TestPerformanceAdvanced 测试性能
func TestPerformanceAdvanced1(t *testing.T) {
	// 文件列表
	files := []string{
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_1_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_2_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_3_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_4_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_5_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_6_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_7_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_8_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_9_d_10_combined.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split\\DB_10_d_10_combined.csv",
	}

	// 参数范围
	LValues := []int{6424, 6424 * 2, 6424 * 3, 6424 * 4, 6424 * 5}
	k := 10

	// 结果存储目录
	resultsDir := "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\results"

	// 遍历每个文件
	for fileIndex, file := range files {
		// 加载倒排索引
		invertedIndex, err := loadInvertedIndex_csv(file)
		if err != nil {
			t.Fatalf("无法加载文件 %s: %v", file, err)
		}

		// 提取文件中的 keywords
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			t.Fatalf("文件 %s 中未找到关键词", file)
		}
		// 测量关键词排序时间
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex) // 排序函数
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 Setup 参数
			sp := Setup(L)

			// 测量 BuildIndex 时间
			startTime := time.Now()
			err = sp.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("BuildIndex 返回错误: %v", err)
			}
			buildIndexDuration := time.Since(startTime).Nanoseconds()
			// 结果文件路径
			resultFilePath := fmt.Sprintf("%s/result_m_%d_L_%d.txt", resultsDir, fileIndex+1, L)
			resultFile, err := os.Create(resultFilePath)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePath, err)
			}
			defer resultFile.Close()

			// 写入结果文件表头
			writer := bufio.NewWriter(resultFile)
			writer.WriteString("Iteration,RangeWidth,BuildIndex(ns),GenToken(ns),SearchTokens(ns),LocalSearch(ns)\n")

			// 开始测试
			for i := 0; i < k; i++ {
				// 随机生成查询范围
				queryRange, rangeWidth := generateQueryRangeWithWidth(sortedKeywords)

				// 测量 GenToken 时间
				startTime = time.Now()
				tokens, err := sp.GenToken(queryRange)
				if err != nil {
					t.Fatalf("GenToken 返回错误: %v", err)
				}
				genTokenDuration := time.Since(startTime).Nanoseconds()

				// 如果 tokens 为空，设置后续耗时为 0 并跳过
				if len(tokens) == 0 {
					searchTokensDuration := 0
					localSearchDuration := 0
					writer.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d\n", i+1, rangeWidth, buildIndexDuration, genTokenDuration, searchTokensDuration, localSearchDuration))
					t.Logf("Tokens are empty, skipping iteration")
					continue
				}

				// 测量 SearchTokens 时间
				startTime = time.Now()
				searchResult := sp.SearchTokens(tokens)
				searchTokensDuration := time.Since(startTime).Nanoseconds()

				// 测量 LocalSearch 时间
				startTime = time.Now()
				_, err = sp.LocalSearch(searchResult, tokens)
				if err != nil {
					t.Fatalf("LocalSearch 返回错误: %v", err)
				}
				localSearchDuration := time.Since(startTime).Nanoseconds()

				// 写入每次实验的耗时记录
				writer.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d\n", i+1, rangeWidth, buildIndexDuration, genTokenDuration, searchTokensDuration, localSearchDuration))
			}

			writer.Flush()
			resultFile.Close()

			// 打印完成信息
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePath)
		}
	}
}
