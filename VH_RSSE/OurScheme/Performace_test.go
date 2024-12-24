package OurScheme

import (
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

// TestCombineCSVFile 合并多个 CSV 文件，生成新的 CSV 文件
func TestCombineCSVFile(t *testing.T) {
	// 文件列表
	files := []string{
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_1_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_2_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_3_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_4_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_5_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_6_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_7_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_8_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_9_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_10_d_10.csv",
	}

	// 遍历 1 到 10 个文件进行合并
	for i := 1; i <= 10; i++ {
		// 打开输出文件
		outputFile := fmt.Sprintf("C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_%d_d_10_combined.csv", i)
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

// TestPerformanceAdvanced 测试性能
func TestPerformanceAdvanced(t *testing.T) {
	// 文件列表
	files := []string{
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_1_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_2_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_3_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_4_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_5_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_6_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_7_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_8_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_9_d_10.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_10_d_10.csv",
	}

	// 参数范围
	LValues := []int{6424, 6424 * 2, 6424 * 3, 6424 * 4, 6424 * 5}
	k := 10

	// 结果存储目录
	resultsDir := "C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\results"

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
				queryRange, rangeWidth := generateQueryRangeWithWidth(keywords)

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

// generateQueryRangeWithWidth 根据关键词生成查询区间并返回区间宽度
func generateQueryRangeWithWidth(keywords []int) ([2]string, int) {
	n := len(keywords)
	if n < 2 {
		return [2]string{"0", "0"}, 0
	}

	leftIndex := rand.Intn(n)
	rightIndex := rand.Intn(n)
	if leftIndex > rightIndex {
		leftIndex, rightIndex = rightIndex, leftIndex
	}

	//// 扰动边界值
	//if rand.Float32() < 0.5 {
	//	leftIndex = max(0, leftIndex-3)
	//}
	//if rand.Float32() < 0.5 {
	//	rightIndex = min(n-1, rightIndex+3)
	//}

	rangeWidth := keywords[rightIndex] - keywords[leftIndex]
	return [2]string{
		strconv.Itoa(keywords[leftIndex]),
		strconv.Itoa(keywords[rightIndex]),
	}, rangeWidth
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

// loadInvertedIndex 从 CSV 文件中加载倒排索引
func loadInvertedIndex(filePath string) (map[string][]int, error) {
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
				return nil, fmt.Errorf("无法解析行号: %s", rowIDStr)
			}

			// 将行号加入 rowIDs
			rowIDs = append(rowIDs, rowID)
		}

		// 更新倒排索引
		invertedIndex[keyword] = rowIDs
	}

	return invertedIndex, nil
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
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_1.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_2.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_3.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_4.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_5.csv",
	}

	// 随机选择一个文件
	rand.Seed(time.Now().UnixNano())
	selectedFile := files[rand.Intn(len(files))]
	fmt.Printf("随机选择文件: %s\n", selectedFile)

	// 加载文件内容到 invertedIndex
	invertedIndex, err := loadInvertedIndex(selectedFile)
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
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_1.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_2.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_3.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_4.csv",
		"C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\DB_1288579_m_5.csv",
	}
	useBoundary := true
	// 参数范围
	LValues := []int{6424, 6424 * 2, 6424 * 3, 6424 * 4, 6424 * 5}
	rangeValues := []int{1, 10, 100, 1000, 10000, 100000}
	k := 10

	// 结果存储目录
	resultsDir := "C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\results"

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
	filePath := "C:\\Users\\Admin\\Desktop\\GoPros\\VolumeHidingSSE\\dataset\\split\\DB_1_d_10.csv"

	// 调用函数加载倒排索引
	invertedIndex, err := loadInvertedIndex(filePath)
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
