package FB_RSSE

import (
	"EfficientAndLowStroageSSE/config"
	"bufio"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

//	func TestPerformanceAdvanced_valid(t *testing.T) {
//		// 文件列表
//		files := []string{
//			"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\Gowalla_invertedIndex_new.txt",
//			// 其他文件路径...
//		}
//
//		// 参数范围
//		LValues := []int{6424, 6424 * 2, 6424 * 3, 6424 * 4, 6424 * 5}
//		k := 999999 // 设置最大查询次数
//		resultCounts := 500
//		// 结果存储目录
//		resultsDir := "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\results\\FB_RSSE"
//
//		// 遍历每个文件
//		for fileIndex, file := range files {
//			// 加载倒排索引
//			invertedIndex, err := loadInvertedIndex(file)
//			if err != nil {
//				t.Fatalf("无法加载文件 %s: %v", file, err)
//			}
//
//			// 提取文件中的 keywords
//			keywords := extractKeywordsFromIndex(invertedIndex)
//			if len(keywords) == 0 {
//				t.Fatalf("文件 %s 中未找到关键词", file)
//			}
//
//			// 测量关键词排序时间
//			startTime := time.Now()
//			sortedKeywords := sortKeywords(invertedIndex) // 排序函数
//			sortDuration := time.Since(startTime).Nanoseconds()
//			t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
//
//			// 遍历每个 L 值
//			for _, L := range LValues {
//				// 初始化 Setup 参数
//				sp := Setup(L)
//
//				// 测量 BuildIndex 时间
//				startTime := time.Now()
//				buildIndexDuration := time.Since(startTime).Nanoseconds()
//
//				// 结果文件路径（CSV格式）
//				resultFilePath := fmt.Sprintf("%s/result_m_%d_L_%d.csv", resultsDir, fileIndex+1, L)
//				resultFile, err := os.Create(resultFilePath)
//				if err != nil {
//					t.Fatalf("无法创建结果文件 %s: %v", resultFilePath, err)
//				}
//				defer resultFile.Close()
//
//				// 写入结果文件表头
//				writer := csv.NewWriter(resultFile)
//				defer writer.Flush()
//				writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)"})
//
//				// 初始化有效查询计数器
//				validCount := 0
//
//				// 开始测试
//				for i := 0; i < k; i++ {
//					// 生成查询区间并计算区间宽度
//					queryRange, rangeWidth := generateQueryRangeWithWidth(sortedKeywords)
//
//					// 测量 GenToken 时间
//					startTime = time.Now()
//					tokens, err := sp.GenToken(queryRange)
//					if err != nil {
//						t.Fatalf("GenToken 返回错误: %v", err)
//					}
//					genTokenDuration := time.Since(startTime).Nanoseconds()
//
//					// 如果 tokens 为空，设置后续耗时为 0 并跳过
//					if len(tokens) == 0 {
//						searchTokensDuration := 0
//						localSearchDuration := 0
//						clientTimeCost := searchTokensDuration + localSearchDuration
//						writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDuration), fmt.Sprintf("%d", genTokenDuration), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
//						t.Logf("Tokens are empty, skipping iteration")
//						continue
//					}
//
//					// 计数有效查询
//					validCount++
//
//					// 测量 SearchTokens 时间
//					startTime = time.Now()
//					searchResult := sp.SearchTokens(tokens)
//					searchTokensDuration := time.Since(startTime).Nanoseconds()
//
//					// 测量 LocalSearch 时间
//					startTime = time.Now()
//					_, err = sp.LocalSearch(searchResult, tokens)
//					if err != nil {
//						t.Fatalf("LocalSearch 返回错误: %v", err)
//					}
//					localSearchDuration := time.Since(startTime).Nanoseconds()
//
//					// 计算 ClientTimeCost
//					clientTimeCost := genTokenDuration + localSearchDuration
//
//					// 写入每次实验的耗时记录
//					writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDuration), fmt.Sprintf("%d", genTokenDuration), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
//
//					// 如果有效查询次数达到 100 次，停止循环
//					if validCount >= resultCounts {
//						t.Logf("达到 100 次有效查询，停止测试")
//						break
//					}
//				}
//
//				// 打印完成信息
//				fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePath)
//			}
//		}
//	}
//

// TestBuildIndex 测试 BuildIndex 方法
func TestBuildIndex(t *testing.T) {
	sp := Setup(1 << 20)
	// 创建一个倒排索引（小范围的文档 ID）
	invertedIndex := map[string][]int{
		"1": {1, 2, 3},
		"2": {4, 5, 6},
		"3": {7, 8, 9},
		"4": {10, 11, 12},
		"5": {13, 14, 15},
	}
	// 创建一个倒排索引（小范围的文档 ID）
	invertedIndex = map[string][]int{
		"1": {1, 2, 3}, "2": {4, 5, 6}, "3": {7, 8, 9}, "4": {10, 11, 12}, "5": {13, 14, 15},
		"6": {16, 17}, "7": {18, 19, 20}, "8": {21, 22, 23}, "9": {24, 25}, "10": {26, 27, 28},
		"11": {29}, "12": {30, 31}, "13": {32, 33, 34}, "14": {35}, // 确保总文档ID数约为35
	}
	// 加载倒排索引
	invertedIndex, _ = loadInvertedIndex(config.FilePath_txt)
	queryRange := [2]string{"5", "10"}
	// 提取文件中的 keywords
	keywords := extractKeywordsFromIndex(invertedIndex)
	if len(keywords) == 0 {
		t.Fatalf("文件中未找到关键词")
	}

	// 测量关键词排序时间
	startTime := time.Now()
	sortedKeywords := sortKeywords(invertedIndex) // 排序函数
	sortDuration := time.Since(startTime).Nanoseconds()
	t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
	// 测试不同数量的倒排索引（1万，2万，...5万）
	//lineCounts := []int{10000, 20000, 30000, 40000, 50000}
	lineCounts := []int{len(sortedKeywords)}

	for _, count := range lineCounts {
		// 截取前 count 个关键词
		subsetIndex := make(map[string][]int)
		for _, keyword := range sortedKeywords[:count] {
			subsetIndex[keyword] = invertedIndex[keyword]
		}

		// 测量 BuildIndex 执行时间
		startTime = time.Now()
		err := sp.BuildIndex(subsetIndex, sortedKeywords[:count])
		if err != nil {
			t.Fatalf("BuildIndex 错误: %v", err)
		}
		buildDuration := time.Since(startTime).Milliseconds() // 转换为毫秒
		t.Logf("处理前 %d 个关键词，BuildIndex 耗时: %d 毫秒", count, buildDuration)

		//// 测量 BRC 时间
		//startTime = time.Now()
		//BRC, _ := sp.getBRC([2]string([]string{"1", "4"}), sortedKeywords)
		//getBRCTime := time.Since(startTime).Milliseconds()
		//t.Logf("BRC: %s ,耗时: %d 毫秒", BRC, getBRCTime)
		// 测量 GenToken 的执行时间
		startTime := time.Now()
		K_w_set, ST_w_set, c_set, err := sp.GenToken(queryRange, sortedKeywords[:count]) // 假设 sortedKeywords[:count] 已经是一个有效的切片
		if err != nil {
			fmt.Println("Error in GenToken:", err)
			return
		}
		genTokenDuration := time.Since(startTime).Milliseconds()
		fmt.Printf("GenToken execution time: %d ms\n", genTokenDuration)

		// 测量 ServerSearch 的执行时间
		startTime = time.Now()
		Sum_e, _ := sp.ServerSearch(K_w_set, ST_w_set, c_set)
		serverSearchDuration := time.Since(startTime).Milliseconds()
		fmt.Printf("ServerSearch execution time: %d ms\n", serverSearchDuration)

		// 测量 LocalParse 的执行时间
		startTime = time.Now()
		Sum_e, _ = sp.LocalParse(K_w_set, c_set, Sum_e)
		localParseDuration := time.Since(startTime).Milliseconds()
		fmt.Printf("LocalParse execution time: %d ms\n", localParseDuration)

		// 测量 PrintBitmap 的执行时间
		startTime = time.Now()
		PrintBitmap(Sum_e, sp.BsLength)
		printBitmapDuration := time.Since(startTime).Milliseconds()
		fmt.Printf("PrintBitmap execution time: %d ms\n", printBitmapDuration)

		// 总时间
		totalDuration := genTokenDuration + serverSearchDuration + localParseDuration + printBitmapDuration
		fmt.Printf("Total execution time: %d ms\n", totalDuration)
	}
	//// 打印每个关键词对应的位图
	//for keyword, data := range sp.EDB {
	//	fmt.Printf("Keyword: %s\n", keyword)
	//	// 假设位图长度是 20 位（或者根据需要调整）
	//	PrintBitmap(data.BigIntValue, sp.BsLength)
	//}
}

// PrintBitmap 输出大整数中所有为 1 的位的位置
func PrintBitmap_time(bitmap *big.Int, maxBits int) {
	fmt.Printf("Bitmap (big.Int): %s\n", bitmap.String())
	fmt.Println("Positions with 1 bits:")
	for i := 0; i < maxBits; i++ {
		if bitmap.Bit(i) == 1 {
			fmt.Printf("Bit %d is set to 1\n", i)
		}
	}
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
