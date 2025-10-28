package main

import (
	"EfficientAndLowStroageSSE/FB_RSSE"
	"EfficientAndLowStroageSSE/VH_RSSE/OurScheme"
	"EfficientAndLowStroageSSE/config"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestComparison 组合 FB_RSSE 和 VH_RSSE 两个测试的实验
func TestComparisonTotal(t *testing.T) {
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
	resultCounts := 300    // 结果存储的有效查询次数

	// 结果存储目录
	resultsDir := "results"
	// 遍历每个文件进行测试
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
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 FB_RSSE 和 OurScheme 的对象
			ours := OurScheme.Setup(L)
			fb_rsse := FB_RSSE.Setup(FB_BsLen)

			// 测量 BuildIndex 时间（OurScheme）
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

			// 测量 BuildIndex 时间（FB_RSSE）
			startTime = time.Now()
			err = fb_rsse.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("FB_RSSE BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationFB := time.Since(startTime).Nanoseconds()

			// 结果文件路径（CSV格式）
			resultFilePathOurs := fmt.Sprintf("%s/comparison_result_m_%d_L_%d_OurScheme.csv", resultsDir, indexNum[fileIndex], L)
			resultFilePathFB := fmt.Sprintf("%s/comparison_result_m_%d_L_%d_FB_RSSE.csv", resultsDir, indexNum[fileIndex], L)

			// 创建并写入结果文件（OurScheme）
			resultFile, err := os.Create(resultFilePathOurs)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePathOurs, err)
			}
			defer resultFile.Close()

			writer := csv.NewWriter(resultFile)
			defer writer.Flush()
			writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "TotalSearchTime(ns)"})

			// 创建并写入结果文件（FB_RSSE）
			resultFile, err = os.Create(resultFilePathFB)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePathFB, err)
			}
			defer resultFile.Close()

			writerFB := csv.NewWriter(resultFile)
			defer writerFB.Flush()
			writerFB.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "CientTimeCost(ns)"})

			// 初始化有效查询计数器
			validCount := 0

			// 开始测试
			for r := range ranges {
				for i := 0; i < k; i++ { // 生成查询区间并计算区间宽度
					queryRange, rangeWidth := generateQueryRangeWithWidth(sortedKeywords, r)

					// 测量 GenToken 时间（OurScheme）
					startTime = time.Now()
					tokensOurs, err := ours.GenToken(queryRange)
					if err != nil {
						t.Fatalf("OurScheme GenToken 返回错误: %v", err)
					}
					genTokenDurationOurs := time.Since(startTime).Nanoseconds()

					// 测量 GenToken 时间（FB_RSSE）
					startTime = time.Now()
					K_set, ST_set, c_set, err := fb_rsse.GenToken(queryRange, sortedKeywords)
					if err != nil {
						t.Fatalf("FB_RSSE GenToken 返回错误: %v", err)
					}
					genTokenDurationFB := time.Since(startTime).Nanoseconds()

					// 如果 tokens 为空，跳过本次循环
					if len(tokensOurs) == 0 || len(c_set) == 0 {
						searchTokensDuration := 0
						localSearchDuration := 0
						clientTimeCost := searchTokensDuration + localSearchDuration
						writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
						writerFB.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationFB), fmt.Sprintf("%d", genTokenDurationFB), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
						t.Logf("Tokens are empty, skipping iteration")
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
						t.Fatalf("OurScheme LocalSearch 返回错误: %v", err)
					}
					localSearchDurationOurs := time.Since(startTime).Nanoseconds()

					// 测量 LocalParse 时间（FB_RSSE）
					startTime = time.Now()
					_, err = fb_rsse.LocalParse(K_set, c_set, searchResultFB)
					if err != nil {
						t.Fatalf("FB_RSSE LocalSearch 返回错误: %v", err)
					}
					localSearchDurationFB := time.Since(startTime).Nanoseconds()

					// 计算 ClientTimeCost（OurScheme）
					clientTimeCostOurs := genTokenDurationOurs + localSearchDurationOurs

					// 计算 ClientTimeCost（FB_RSSE）
					clientTimeCostFB := genTokenDurationFB + localSearchDurationFB

					// 写入每次实验的耗时记录（OurScheme）
					writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDurationOurs), fmt.Sprintf("%d", localSearchDurationOurs), fmt.Sprintf("%d", clientTimeCostOurs)})

					// 写入每次实验的耗时记录（FB_RSSE）
					writerFB.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationFB), fmt.Sprintf("%d", genTokenDurationFB), fmt.Sprintf("%d", searchTokensDurationFB), fmt.Sprintf("%d", localSearchDurationFB), fmt.Sprintf("%d", clientTimeCostFB)})

					// 如果有效查询次数达到 300 次，停止循环
					if validCount >= resultCounts {
						t.Logf("达到 %d 次有效查询，停止测试", resultCounts)
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

// TestComparisonBuildIndex 测试 BuildIndex 方法的耗时
func TestComparisonBuildIndex(t *testing.T) {
	// 设置参数
	files := []string{
		"dataset/Gowalla_invertedIndex_new_5000.txt",
		"dataset/Gowalla_invertedIndex_new_10000.txt",
		"dataset/Gowalla_invertedIndex_new_15000.txt",
		"dataset/Gowalla_invertedIndex_new_20000.txt",
	}

	indexNum := []int{5000, 10000, 15000, 20000}
	FB_BsLen := 1 << 15    // 设置 FB_RSSE 的参数
	LValues := []int{6424} // 设置 L 值范围
	//k := 999999            // 设置最大查询次数
	//resultCounts := 500    // 结果存储的有效查询次数

	// 结果存储目录
	resultsDir := "results/results_20251028"

	// 遍历每个文件进行测试
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
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 FB_RSSE 和 OurScheme 的对象
			ours := OurScheme.Setup(L)
			fb_rsse := FB_RSSE.Setup(FB_BsLen)

			// 测量 BuildIndex 时间（OurScheme）
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

			// 测量 BuildIndex 时间（FB_RSSE）
			startTime = time.Now()
			err = fb_rsse.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("FB_RSSE BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationFB := time.Since(startTime).Nanoseconds()

			// 输出文件路径（按行数命名）
			resultFilePathOurs := fmt.Sprintf("%s/result_m_%d_L_%d_OurScheme_BuildIndex.csv", resultsDir, indexNum[fileIndex], L)
			resultFilePathFB := fmt.Sprintf("%s/result_m_%d_L_%d_FB_RSSE_BuildIndex.csv", resultsDir, indexNum[fileIndex], L)

			// 创建并写入结果文件（OurScheme）
			resultFile, err := os.Create(resultFilePathOurs)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePathOurs, err)
			}
			defer resultFile.Close()

			writer := csv.NewWriter(resultFile)
			defer writer.Flush()
			writer.Write([]string{"File", "Lines", "BuildIndex(ns)"}) // 文件名，行数，BuildIndex耗时

			// 写入结果文件
			writer.Write([]string{file, fmt.Sprintf("%d", len(invertedIndex)), fmt.Sprintf("%d", buildIndexDurationOurs)})

			// 创建并写入结果文件（FB_RSSE）
			resultFile, err = os.Create(resultFilePathFB)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePathFB, err)
			}
			defer resultFile.Close()

			writerFB := csv.NewWriter(resultFile)
			defer writerFB.Flush()
			writerFB.Write([]string{"File", "Lines", "BuildIndex(ns)"}) // 文件名，行数，BuildIndex耗时

			// 写入结果文件
			writerFB.Write([]string{file, fmt.Sprintf("%d", len(invertedIndex)), fmt.Sprintf("%d", buildIndexDurationFB)})

			// 打印完成信息
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathOurs)
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathFB)
		}
	}
}
func TestGenerateQueryRangeWithWidth(t *testing.T) {
	// 测试数据：一个已排序的 keywords 列表
	sortedKeywords := []string{"2", "100", "150", "200", "300", "400"}

	// 测试宽度 r
	r := 100

	// 调用待测试的函数
	result, width := generateQueryRangeWithWidth(sortedKeywords, r)

	// 检查生成的区间宽度是否与 r 相等
	left, _ := strconv.Atoi(result[0])
	right, _ := strconv.Atoi(result[1])
	if right-left != r {
		t.Errorf("Expected width %d, but got %d", r, right-left)
	}

	// 确保返回的宽度和传入的宽度一致
	if width != r {
		t.Errorf("Expected width %d, but got %d", r, width)
	}

	// 测试宽度为 0 的情况
	emptyResult, emptyWidth := generateQueryRangeWithWidth(sortedKeywords, 0)
	if emptyWidth != 0 {
		t.Errorf("Expected width 0, but got %d", emptyWidth)
	}
	if emptyResult != [2]string{"0", "0"} {
		t.Errorf("Expected result [0, 0], but got %v", emptyResult)
	}

	// 测试当关键词列表长度小于2时的情况
	emptyKeywords := []string{}
	emptyResult, emptyWidth = generateQueryRangeWithWidth(emptyKeywords, r)
	if emptyWidth != 0 {
		t.Errorf("Expected width 0, but got %d", emptyWidth)
	}
	if emptyResult != [2]string{"0", "0"} {
		t.Errorf("Expected result [0, 0], but got %v", emptyResult)
	}
}

func TestOurSchemeBuildIndex_storage_metadata(t *testing.T) {
	// 设置参数
	files := []string{
		"dataset/Gowalla_invertedIndex_new_5000.txt",
		"dataset/Gowalla_invertedIndex_new_10000.txt",
		"dataset/Gowalla_invertedIndex_new_15000.txt",
		"dataset/Gowalla_invertedIndex_new_20000.txt",
		config.FilePath_txt,
	}

	LValues := []int{6424} // 设置 L 值范围

	// 遍历每个文件进行测试
	for _, file := range files {
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
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 OurScheme 的对象
			ours := OurScheme.Setup(L)

			// 测量 BuildIndex 时间（OurScheme）
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

			// 打印构建时间
			fmt.Printf("文件: %s, L: %d, BuildIndex耗时: %d 纳秒\n", file, L, buildIndexDurationOurs)

			// 打印 EDB、LocalTree、ClusterFlist 和 ClusterKlist 的长度
			fmt.Println("\nEDB 长度:", len(ours.EDB))
			fmt.Println("LocalTree 长度:", len(ours.LocalTree))
			// 计算 ClusterFlist 和 ClusterKlist 的总元素数量
			clusterFlistTotal := countNestedElements(ours.ClusterFlist)
			clusterKlistTotal := countNestedElements(ours.ClusterKlist)

			// 打印 ClusterFlist 和 ClusterKlist 的长度和所有元素的数量
			fmt.Printf("ClusterFlist 长度: %d, 所有元素数量: %d\n", len(ours.ClusterFlist), clusterFlistTotal)
			fmt.Printf("ClusterKlist 长度: %d, 所有元素数量: %d\n", len(ours.ClusterKlist), clusterKlistTotal)

			//// 计算这些结构的内存占用
			//fmt.Printf("\n内存占用 (KB/MB):\n")
			//printMemoryUsage(ours.EDB, "EDB")
			//printMemoryUsage(ours.LocalTree, "LocalTree")
			//printMemoryUsage(ours.ClusterFlist, "ClusterFlist")
			//printMemoryUsage(ours.ClusterKlist, "ClusterKlist")
		}
	}
}

// countNestedElements 计算嵌套切片中的所有元素数量
func countNestedElements(nestedSlice interface{}) int {
	totalCount := 0
	switch v := nestedSlice.(type) {
	case [][]int:
		for _, subSlice := range v {
			totalCount += len(subSlice)
		}
	case [][]string:
		for _, subSlice := range v {
			totalCount += len(subSlice)
		}
	default:
		fmt.Println("无法计算嵌套切片的元素数量")
	}
	return totalCount
}
func TestOurSchemeBuildIndex_storage(t *testing.T) {
	// 设置参数
	files := []string{
		"dataset/Gowalla_invertedIndex_new_5000.txt",
		"dataset/Gowalla_invertedIndex_new_10000.txt",
		"dataset/Gowalla_invertedIndex_new_15000.txt",
		"dataset/Gowalla_invertedIndex_new_20000.txt",
	}

	LValues := []int{6424} // 设置 L 值范围

	// 遍历每个文件进行测试
	for _, file := range files {
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
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 OurScheme 的对象
			ours := OurScheme.Setup(L)

			// 测量 BuildIndex 时间（OurScheme）
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

			// 打印构建时间
			fmt.Printf("文件: %s, L: %d, BuildIndex耗时: %d 纳秒\n", file, L, buildIndexDurationOurs)

			////打印 EDB、LocalTree、ClusterFlist 和 ClusterKlist 的值
			//fmt.Println("\nEDB (加密倒排索引):")
			//printMapValues(ours.EDB)
			//
			//fmt.Println("\nLocalTree (本地树):")
			//printMapValues(ours.LocalTree)
			//
			//fmt.Println("\nClusterFlist (分区文件列表):")
			//printNestedSliceValues(ours.ClusterFlist)
			//
			//fmt.Println("\nClusterKlist (分区关键词列表):")
			//printNestedSliceValues(ours.ClusterKlist)

			// 计算这些结构的内存占用
			fmt.Printf("\n内存占用 (KB/MB):\n")
			printMemoryUsage(ours.EDB, "EDB")
			printMemoryUsage(ours.LocalTree, "LocalTree")
			printMemoryUsage(ours.ClusterFlist, "ClusterFlist")
			printMemoryUsage(ours.ClusterKlist, "ClusterKlist")
		}
	}
}

// printMapValues 打印 map 数据结构的值
func printMapValues(m map[string][]byte) {
	fmt.Printf("Map Size: %d entries\n", len(m))
	for key, value := range m {
		fmt.Printf("Key: %s, Value length: %d bytes\n", key, len(value))
	}
}

// printMapValues 打印 map 数据结构的值（针对 int64 的 map）
func printMapValuesInt64(m map[string][]int64) {
	fmt.Printf("Map Size: %d entries\n", len(m))
	for key, value := range m {
		fmt.Printf("Key: %s, Value length: %d entries\n", key, len(value))
	}
}

// printNestedSliceValues 打印嵌套切片的数据结构值
func printNestedSliceValues(s interface{}) {
	switch v := s.(type) {
	case [][]int:
		fmt.Printf("Nested Slice Size: %d entries\n", len(v))
		for i, subSlice := range v {
			fmt.Printf("Sub-slice %d length: %d\n", i, len(subSlice))
		}
	case [][]string:
		fmt.Printf("Nested Slice Size: %d entries\n", len(v))
		for i, subSlice := range v {
			fmt.Printf("Sub-slice %d length: %d\n", i, len(subSlice))
		}
	default:
		fmt.Println("无法打印嵌套切片的值")
	}
}

// printMemoryUsage 计算和打印数据结构的内存占用
func printMemoryUsage(data interface{}, label string) {
	var size int64
	switch v := data.(type) {
	case map[string][]byte:
		for key, value := range v {
			size += int64(len(key))   // key的大小
			size += int64(len(value)) // value的大小
		}
	case map[string][]int64:
		for key, value := range v {
			size += int64(len(key))       // key的大小
			size += int64(len(value) * 8) // 每个int64占用8字节
		}
	case [][]int:
		for _, subSlice := range v {
			size += int64(len(subSlice) * 8) // 每个int占用4字节
		}
	case [][]string:
		for _, subSlice := range v {
			size += int64(len(subSlice) * 16) // 假设每个string占16字节（包括指针和长度）
		}
	default:
		fmt.Println("无法计算内存占用")
	}

	sizeInMB := float64(size) / (1024 * 1024) // 转换为MB
	sizeInKB := float64(size) / 1024          // 转换为KB
	fmt.Printf("%s 内存占用: %.2f KB (%.2f MB)\n", label, sizeInKB, sizeInMB)
}

func TestOurSchemeBuildIndex2(t *testing.T) {
	// 设置参数
	files := []string{
		config.FilePath_txt,
	}

	Lines := []int{67070, 134140, 201210, 268280, 335349}
	LValues := []int{3212, 6424, 9636, 12848, 16060} // 设置 L 值范围
	//k := 999999            // 设置最大查询次数
	//resultCounts := 300    // 结果存储的有效查询次数

	// 结果存储目录
	resultsDir := "results"

	// 遍历每个文件进行测试
	for _, file := range files {
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
		SortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
		for _, lines := range Lines {
			sortedKeywords := SortedKeywords[:lines]
			// 遍历每个 L 值
			for _, L := range LValues {
				// 初始化 FB_RSSE 和 OurScheme 的对象
				ours := OurScheme.Setup(L)

				// 测量 BuildIndex 时间（OurScheme）
				startTime = time.Now()
				err = ours.BuildIndex(invertedIndex, sortedKeywords)
				if err != nil {
					t.Fatalf("OurScheme BuildIndex 返回错误: %v", err)
				}
				buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

				// 输出文件路径（按行数命名）
				resultFilePathOurs := fmt.Sprintf("%s/self_result_m_%d_L_%d_OurScheme_BuildIndex.csv", resultsDir, lines, L)
				// 创建并写入结果文件（OurScheme）
				resultFile, err := os.Create(resultFilePathOurs)
				if err != nil {
					t.Fatalf("无法创建结果文件 %s: %v", resultFilePathOurs, err)
				}
				defer resultFile.Close()

				writer := csv.NewWriter(resultFile)
				defer writer.Flush()
				writer.Write([]string{"File", "Lines", "BuildIndex(ns)"}) // 文件名，行数，BuildIndex耗时

				// 写入结果文件
				writer.Write([]string{file, fmt.Sprintf("%d", len(invertedIndex)), fmt.Sprintf("%d", buildIndexDurationOurs)})

				// 打印完成信息
				fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathOurs)
			}
		}
	}
}
func TestSelfTotal(t *testing.T) {
	// 设置参数
	files := []string{
		config.FilePath_txt,
	}

	Lines := []int{67070, 134140, 201210, 268280, 335349}
	LValues := []int{3212, 6424, 9636, 12848, 16060} // 设置 L 值范围
	ranges := []int{600, 600 * 2, 600 * 3, 600 * 4, 600 * 5, 600 * 6, 600 * 7, 600 * 8}
	k := 999999         // 设置最大查询次数
	resultCounts := 300 // 结果存储的有效查询次数

	// 结果存储目录
	resultsDir := "results"
	// 遍历每个文件进行测试
	for _, file := range files {
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
		SortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 FB_RSSE 和 OurScheme 的对象
			ours := OurScheme.Setup(L)
			for lines := range Lines {
				sortedKeywords := SortedKeywords[:lines]
				// 测量 BuildIndex 时间（OurScheme）
				startTime = time.Now()
				err = ours.BuildIndex(invertedIndex, sortedKeywords)
				if err != nil {
					t.Fatalf("OurScheme BuildIndex 返回错误: %v", err)
				}
				buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

				// 结果文件路径（CSV格式）
				resultFilePathOurs := fmt.Sprintf("%s/self_result_m_%d_L_%d_OurScheme.csv", resultsDir, lines, L)

				// 创建并写入结果文件（OurScheme）
				resultFile, err := os.Create(resultFilePathOurs)
				if err != nil {
					t.Fatalf("无法创建结果文件 %s: %v", resultFilePathOurs, err)
				}
				defer resultFile.Close()

				writer := csv.NewWriter(resultFile)
				defer writer.Flush()
				writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "TotalSearchTime(ns)"})

				// 初始化有效查询计数器
				validCount := 0

				// 开始测试
				for r := range ranges {
					for i := 0; i < k; i++ { // 生成查询区间并计算区间宽度
						queryRange, rangeWidth := generateQueryRangeWithWidth(sortedKeywords, r)

						// 测量 GenToken 时间（OurScheme）
						startTime = time.Now()
						tokensOurs, err := ours.GenToken(queryRange)
						if err != nil {
							t.Fatalf("OurScheme GenToken 返回错误: %v", err)
						}
						genTokenDurationOurs := time.Since(startTime).Nanoseconds()

						// 如果 tokens 为空，跳过本次循环
						if len(tokensOurs) == 0 {
							searchTokensDuration := 0
							localSearchDuration := 0
							clientTimeCost := searchTokensDuration + localSearchDuration
							writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
							t.Logf("Tokens are empty, skipping iteration")
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
							t.Fatalf("OurScheme LocalSearch 返回错误: %v", err)
						}
						localSearchDurationOurs := time.Since(startTime).Nanoseconds()

						// 计算 ClientTimeCost（OurScheme）
						clientTimeCostOurs := genTokenDurationOurs + localSearchDurationOurs

						// 写入每次实验的耗时记录（OurScheme）
						writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDurationOurs), fmt.Sprintf("%d", localSearchDurationOurs), fmt.Sprintf("%d", clientTimeCostOurs)})

						// 如果有效查询次数达到 300 次，停止循环
						if validCount >= resultCounts {
							t.Logf("达到 %d 次有效查询，停止测试", resultCounts)
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

// TestComparison 组合 FB_RSSE 和 VH_RSSE 两个测试的实验
func TestComparison(t *testing.T) {
	// 设置参数
	files := []string{
		//"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\Gowalla_invertedIndex_new.txt",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split_files\\Gowalla_invertedIndex_new_1000.txt",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split_files\\Gowalla_invertedIndex_new_3000.txt",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split_files\\Gowalla_invertedIndex_new_5000.txt",
		"C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\split_files\\Gowalla_invertedIndex_new_7000.txt",
		// 其他文件路径...
	}

	FB_BsLen := 1 << 20                                            // 设置 FB_RSSE 的参数
	LValues := []int{6424, 6424 * 2, 6424 * 3, 6424 * 4, 6424 * 5} // 设置 L 值范围
	k := 999999                                                    // 设置最大查询次数
	resultCounts := 500                                            // 结果存储的有效查询次数

	// 结果存储目录
	resultsDir := "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\results"

	// 遍历每个文件进行测试
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
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个 L 值
		for _, L := range LValues {
			// 初始化 FB_RSSE 和 OurScheme 的对象
			ours := OurScheme.Setup(L)
			fb_rsse := FB_RSSE.Setup(FB_BsLen)

			// 测量 BuildIndex 时间（OurScheme）
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationOurs := time.Since(startTime).Nanoseconds()

			// 测量 BuildIndex 时间（FB_RSSE）
			startTime = time.Now()
			err = fb_rsse.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("FB_RSSE BuildIndex 返回错误: %v", err)
			}
			buildIndexDurationFB := time.Since(startTime).Nanoseconds()

			// 结果文件路径（CSV格式）
			resultFilePathOurs := fmt.Sprintf("%s/result_m_%d_L_%d_OurScheme.csv", resultsDir, fileIndex+1, L)
			resultFilePathFB := fmt.Sprintf("%s/result_m_%d_L_%d_FB_RSSE.csv", resultsDir, fileIndex+1, L)

			// 创建并写入结果文件（OurScheme）
			resultFile, err := os.Create(resultFilePathOurs)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePathOurs, err)
			}
			defer resultFile.Close()

			writer := csv.NewWriter(resultFile)
			defer writer.Flush()
			writer.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)"})

			// 创建并写入结果文件（FB_RSSE）
			resultFile, err = os.Create(resultFilePathFB)
			if err != nil {
				t.Fatalf("无法创建结果文件 %s: %v", resultFilePathFB, err)
			}
			defer resultFile.Close()

			writerFB := csv.NewWriter(resultFile)
			defer writerFB.Flush()
			writerFB.Write([]string{"Iteration", "Left", "Right", "RangeWidth", "BuildIndex(ns)", "GenToken(ns)", "SearchTokens(ns)", "LocalSearch(ns)", "ClientTimeCost(ns)"})

			// 初始化有效查询计数器
			validCount := 0

			// 开始测试
			for i := 0; i < k; i++ {
				// 生成查询区间并计算区间宽度
				queryRange, rangeWidth := generateQueryRange(sortedKeywords)

				// 测量 GenToken 时间（OurScheme）
				startTime = time.Now()
				tokensOurs, err := ours.GenToken(queryRange)
				if err != nil {
					t.Fatalf("OurScheme GenToken 返回错误: %v", err)
				}
				genTokenDurationOurs := time.Since(startTime).Nanoseconds()

				// 测量 GenToken 时间（FB_RSSE）
				startTime = time.Now()
				K_set, ST_set, c_set, err := fb_rsse.GenToken(queryRange, sortedKeywords)
				if err != nil {
					t.Fatalf("FB_RSSE GenToken 返回错误: %v", err)
				}
				genTokenDurationFB := time.Since(startTime).Nanoseconds()

				// 如果 tokens 为空，跳过本次循环
				if len(tokensOurs) == 0 || len(c_set) == 0 {
					searchTokensDuration := 0
					localSearchDuration := 0
					clientTimeCost := searchTokensDuration + localSearchDuration
					writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
					writerFB.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1], fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationFB), fmt.Sprintf("%d", genTokenDurationFB), fmt.Sprintf("%d", searchTokensDuration), fmt.Sprintf("%d", localSearchDuration), fmt.Sprintf("%d", clientTimeCost)})
					t.Logf("Tokens are empty, skipping iteration")
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
					t.Fatalf("OurScheme LocalSearch 返回错误: %v", err)
				}
				localSearchDurationOurs := time.Since(startTime).Nanoseconds()

				// 测量 LocalParse 时间（FB_RSSE）
				startTime = time.Now()
				_, err = fb_rsse.LocalParse(K_set, c_set, searchResultFB)
				if err != nil {
					t.Fatalf("FB_RSSE LocalSearch 返回错误: %v", err)
				}
				localSearchDurationFB := time.Since(startTime).Nanoseconds()

				// 计算 ClientTimeCost（OurScheme）
				clientTimeCostOurs := genTokenDurationOurs + localSearchDurationOurs

				// 计算 ClientTimeCost（FB_RSSE）
				clientTimeCostFB := genTokenDurationFB + localSearchDurationFB

				// 写入每次实验的耗时记录（OurScheme）
				writer.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1],
					fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationOurs), fmt.Sprintf("%d", genTokenDurationOurs), fmt.Sprintf("%d", searchTokensDurationOurs), fmt.Sprintf("%d", localSearchDurationOurs), fmt.Sprintf("%d", clientTimeCostOurs)})

				// 写入每次实验的耗时记录（FB_RSSE）
				writerFB.Write([]string{fmt.Sprintf("%d", i+1), queryRange[0], queryRange[1],
					fmt.Sprintf("%d", rangeWidth), fmt.Sprintf("%d", buildIndexDurationFB), fmt.Sprintf("%d", genTokenDurationFB), fmt.Sprintf("%d", searchTokensDurationFB), fmt.Sprintf("%d", localSearchDurationFB), fmt.Sprintf("%d", clientTimeCostFB)})

				// 如果有效查询次数达到 100 次，停止循环
				if validCount >= resultCounts {
					t.Logf("达到 100 次有效查询，停止测试")
					break
				}
			}

			// 打印完成信息
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathOurs)
			fmt.Printf("完成文件: %s, L: %d, 结果存储于: %s\n", file, L, resultFilePathFB)
		}
	}
}

func generateQueryRange(keywords []string) ([2]string, int) {
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
