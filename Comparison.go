package main

import (
	"EfficientAndLowStroageSSE/FB_RSSE"
	"EfficientAndLowStroageSSE/VH_RSSE/OurScheme"
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// main 组合 FB_RSSE 和 VH_RSSE 两个测试的实验
func main() {
	// 设置参数
	files := []string{
		"dataset/Gowalla_invertedIndex_new_5000.txt",
		"dataset/Gowalla_invertedIndex_new_10000.txt",
		"dataset/Gowalla_invertedIndex_new_15000.txt",
		"dataset/Gowalla_invertedIndex_new_20000.txt",
	}

	indexNum := []int{5000, 10000, 15000, 20000}
	ranges := []int{600, 600 * 2, 600 * 3, 600 * 4, 600 * 5, 600 * 6, 600 * 7, 600 * 8}
	FB_BsLen := 1 << 15    // 设置 FB_RSSE 的参数
	LValues := []int{6424} // 设置 L 值范围
	k := 999999            // 设置最大查询次数
	resultCounts := 50     // 结果存储的有效查询次数

	// 结果存储目录
	resultsDir := "results"
	// 遍历每个文件进行测试
	for fileIndex, file := range files {
		// 加载倒排索引
		invertedIndex, err := loadInvertedIndex(file)
		if err != nil {
			fmt.Printf("无法加载文件 %s: %v\n", file, err)
			return
		}

		// 提取文件中的 keywords
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			fmt.Printf("文件 %s 中未找到关键词\n", file)
			return
		}

		// 测量关键词排序时间
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		fmt.Printf("关键词排序耗时: %d 纳秒\n", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 FB_RSSE 和 OurScheme 的对象
			ours := OurScheme.Setup(L)
			fb_rsse := FB_RSSE.Setup(FB_BsLen)

			// 测量 BuildIndex 时间（OurScheme）
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				fmt.Printf("OurScheme BuildIndex 返回错误: %v\n", err)
				return
			}
			buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

			// 测量 BuildIndex 时间（FB_RSSE）
			startTime = time.Now()
			err = fb_rsse.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				fmt.Printf("FB_RSSE BuildIndex 返回错误: %v\n", err)
				return
			}
			buildIndexDurationFB := time.Since(startTime).Nanoseconds()

			// 结果文件路径（CSV格式）
			resultFilePathOurs := fmt.Sprintf("%s/comparison_result_m_%d_L_%d_OurScheme.csv", resultsDir, indexNum[fileIndex], L)
			resultFilePathFB := fmt.Sprintf("%s/comparison_result_m_%d_L_%d_FB_RSSE.csv", resultsDir, indexNum[fileIndex], L)

			// 创建并写入结果文件（OurScheme）
			resultFile, err := os.Create(resultFilePathOurs)
			if err != nil {
				fmt.Printf("无法创建结果文件 %s: %v\n", resultFilePathOurs, err)
				return
			}
			defer resultFile.Close()

			writer := csv.NewWriter(resultFile)
			defer writer.Flush()
			writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)", "number of tokens"})

			// 创建并写入结果文件（FB_RSSE）
			resultFile, err = os.Create(resultFilePathFB)
			if err != nil {
				fmt.Printf("无法创建结果文件 %s: %v\n", resultFilePathFB, err)
				return
			}
			defer resultFile.Close()

			writerFB := csv.NewWriter(resultFile)
			defer writerFB.Flush()
			writerFB.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)", "number of tokens"})

			// 开始测试
			for _, r := range ranges {
				// 初始化有效查询计数器
				validCount := 0
				for i := 0; i < k; i++ { // 生成查询区间并计算区间宽度
					queryRange, rangeWidth := generateQueryRangeWithWidth(sortedKeywords, r)

					// 测量 GenToken 时间（OurScheme）
					startTime = time.Now()
					tokensOurs, err := ours.GenToken(queryRange)
					if err != nil {
						fmt.Printf("OurScheme GenToken 返回错误: %v\n", err)
						return
					}
					genTokenDurationOurs := time.Since(startTime).Nanoseconds()

					// 测量 GenToken 时间（FB_RSSE）
					startTime = time.Now()
					K_set, ST_set, c_set, err := fb_rsse.GenToken(queryRange, sortedKeywords)
					if err != nil {
						fmt.Printf("FB_RSSE GenToken 返回错误: %v\n", err)
						return
					}
					genTokenDurationFB := time.Since(startTime).Nanoseconds()

					// 如果 tokens 为空，跳过本次循环
					if len(tokensOurs) == 0 || len(c_set) == 0 {
						searchTokensDuration := 0
						localSearchDuration := 0
						clientTimeCost := searchTokensDuration + localSearchDuration
						writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost), fmt.Sprintf("%d", len(tokensOurs))})
						writerFB.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationFB), fmt.Sprintf("%d", genTokenDurationFB), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost), fmt.Sprintf("%d", len(c_set))})
						fmt.Println("Tokens are empty, skipping iteration")
						continue
					}

					// 计数有效查询
					validCount++

					// 测量 SearchTokens 时间（OurScheme）
					startTime = time.Now()
					searchResultOurs := ours.SearchTokens(tokensOurs)
					searchTokensDurationOurs := time.Since(startTime).Nanoseconds()

					// 测量 ServerSearch 时间（FB_RSSE）
					startTime = time.Now()
					searchResultFB, _ := fb_rsse.ServerSearch(K_set, ST_set, c_set)
					searchTokensDurationFB := time.Since(startTime).Nanoseconds()

					// 测量 LocalSearch 时间（OurScheme）
					startTime = time.Now()
					_, err = ours.LocalSearch(searchResultOurs, tokensOurs)
					if err != nil {
						fmt.Printf("OurScheme LocalSearch 返回错误: %v\n", err)
						return
					}
					localSearchDurationOurs := time.Since(startTime).Nanoseconds()

					// 测量 LocalParse 时间（FB_RSSE）
					startTime = time.Now()
					_, err = fb_rsse.LocalParse(K_set, c_set, searchResultFB)
					if err != nil {
						fmt.Printf("FB_RSSE LocalSearch 返回错误: %v\n", err)
						return
					}
					localSearchDurationFB := time.Since(startTime).Nanoseconds()

					// 计算 ClientTimeCost（OurScheme）
					clientTimeCostOurs := genTokenDurationOurs + localSearchDurationOurs

					// 计算 ClientTimeCost（FB_RSSE）
					clientTimeCostFB := genTokenDurationFB + localSearchDurationFB

					// 写入每次实验的耗时记录（OurScheme）
					writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDurationOurs), fmt.Sprintf("%d", localSearchDurationOurs), fmt.Sprintf("%d", clientTimeCostOurs), fmt.Sprintf("%d", len(tokensOurs))})

					// 写入每次实验的耗时记录（FB_RSSE）
					writerFB.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationFB), fmt.Sprintf("%d", genTokenDurationFB), fmt.Sprintf("%d", searchTokensDurationFB), fmt.Sprintf("%d", localSearchDurationFB), fmt.Sprintf("%d", clientTimeCostFB), fmt.Sprintf("%d", len(c_set))})

					// 如果有效查询次数达到 300 次，停止循环
					if validCount >= resultCounts {
						fmt.Printf("达到 %d 次有效查询，停止测试\n", resultCounts)
						break
					}
				}
			}

			// 打印完成信息
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathOurs)
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathFB)
		}
	}
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
func generateQueryRangeWithWidth(keywords []string, width int) ([2]string, int) {
	n := len(keywords)
	if n < 2 || width <= 0 {
		return [2]string{"0", "0"}, 0
	}

	// 随机选择一个左边界索引 i
	i := rand.Intn(n)

	// 将左边界转换为整数
	left, err := strconv.Atoi(keywords[i])
	if err != nil {
		// 错误处理：如果转换失败，返回默认值
		return [2]string{"0", "0"}, 0
	}

	// 计算右边界为 left + width
	right := left + width

	// 获取 keywords 中的最大值
	maxKeyword, err := strconv.Atoi(keywords[n-1])
	if err != nil {
		// 错误处理：如果转换失败，返回默认值
		return [2]string{"0", "0"}, 0
	}

	// 确保右边界不超出最大值
	for right > maxKeyword {
		i = rand.Intn(n)
		left, err = strconv.Atoi(keywords[i])
		if err != nil {
			// 错误处理：如果转换失败，返回默认值
			return [2]string{"0", "0"}, 0
		}
		right = left + width
	}

	// 返回生成的区间 [left, right]
	return [2]string{strconv.Itoa(left), strconv.Itoa(right)}, width
}
