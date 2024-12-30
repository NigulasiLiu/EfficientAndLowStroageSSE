package main

import (
	"EfficientAndLowStroageSSE/VH_RSSE/OurScheme"
	"EfficientAndLowStroageSSE/config"
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func main1() {
	// 设置参数
	files := []string{
		config.FilePath_txt,
	}

	Lines := []int{67070, 134140, 201210, 268280, 335349}
	LValues := []int{1606, 3212, 1606 * 3, 6424, 1606 * 5} // 设置 L 值范围
	ranges := []int{6000, 6000 * 2, 6000 * 3, 6000 * 4, 6000 * 5, 6000 * 6, 6000 * 7, 6000 * 8}
	k := 999999         // 设置最大查询次数
	resultCounts := 100 // 结果存储的有效查询次数

	// 结果存储目录
	resultsDir := "results"
	// 遍历每个文件进行测试
	for _, file := range files {
		// 加载倒排索引
		invertedIndex, err := loadInvertedIndex(file)
		if err != nil {
			fmt.Printf("无法加载文件 %s: %v", file, err)
		}

		// 提取文件中的 keywords
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			fmt.Printf("文件 %s 中未找到关键词", file)
		}

		// 测量关键词排序时间
		startTime := time.Now()
		SortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		fmt.Printf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 FB_RSSE 和 OurScheme 的对象
			ours := OurScheme.Setup(L)
			for _, lines := range Lines {
				sortedKeywords := SortedKeywords[:lines]
				tempInvertedIndex := make(map[string][]int)
				// 遍历 sortedKeywords，对于每个关键词，检查 invertedIndex 是否存在该键
				for _, keyword := range sortedKeywords {
					if value, exists := invertedIndex[keyword]; exists {
						tempInvertedIndex[keyword] = value
					}
				}
				// 测量 BuildIndex 时间（OurScheme）
				startTime = time.Now()
				err = ours.BuildIndex(tempInvertedIndex, sortedKeywords)
				if err != nil {
					fmt.Printf("OurScheme BuildIndex 返回错误: %v", err)
				}
				buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

				// 结果文件路径（CSV格式）
				resultFilePathOurs := fmt.Sprintf("%s/self_result_m_%d_L_%d_OurScheme.csv", resultsDir, lines, L)

				// 创建并写入结果文件（OurScheme）
				resultFile, err := os.Create(resultFilePathOurs)
				if err != nil {
					fmt.Printf("无法创建结果文件 %s: %v", resultFilePathOurs, err)
				}
				defer resultFile.Close()

				writer := csv.NewWriter(resultFile)
				defer writer.Flush()
				writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)", "number of tokens"})

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
							fmt.Printf("OurScheme GenToken 返回错误: %v", err)
						}
						genTokenDurationOurs := time.Since(startTime).Nanoseconds()

						// 如果 tokens 为空，跳过本次循环
						if len(tokensOurs) == 0 {
							searchTokensDuration := 0
							localSearchDuration := 0
							clientTimeCost := searchTokensDuration + localSearchDuration
							writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost), fmt.Sprintf("%d", len(tokensOurs))})
							fmt.Printf("Tokens are empty, skipping iteration")
							continue
						}

						// 计数有效查询
						validCount++

						// 测量 SearchTokens 时间（OurScheme）
						startTime = time.Now()
						searchResultOurs := ours.SearchTokens(tokensOurs)
						searchTokensDurationOurs := time.Since(startTime).Nanoseconds()

						// 测量 LocalSearch 时间（OurScheme）
						startTime = time.Now()
						_, err = ours.LocalSearch(searchResultOurs, tokensOurs)
						if err != nil {
							fmt.Printf("OurScheme LocalSearch 返回错误: %v", err)
						}
						localSearchDurationOurs := time.Since(startTime).Nanoseconds()

						// 计算 ClientTimeCost（OurScheme）
						clientTimeCostOurs := genTokenDurationOurs + localSearchDurationOurs

						// 写入每次实验的耗时记录（OurScheme）
						writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDurationOurs), fmt.Sprintf("%d", localSearchDurationOurs), fmt.Sprintf("%d", clientTimeCostOurs), fmt.Sprintf("%d", len(tokensOurs))})

						// 如果有效查询次数达到 300 次，停止循环
						if validCount >= resultCounts {
							fmt.Printf("达到 %d 次有效查询，停止测试", resultCounts)
							break
						}
					}

				}

				// 打印完成信息
				fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathOurs)
			}

		}
	}
}
