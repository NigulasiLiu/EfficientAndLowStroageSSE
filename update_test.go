package main

import (
	"EfficientAndLowStroageSSE/FB_RSSE"
	"EfficientAndLowStroageSSE/VH_RSSE/OurScheme"
	"math/big"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"
)

// 测试配置
const (
	L            = 6424                               // 固定L值
	resultFile   = "update_time_results.txt"          // 结果输出文件
	datasetDir   = "dataset/"                         // 数据集目录
	fileTemplate = "Gowalla_invertedIndex_new_%d.txt" // 数据集文件名模板
)

// 待测试的关键字数量（仅小数据集）
var testMs = []int{5000, 10000, 15000, 20000}

// 提取并排序关键字
func getSortedKeywords(invertedIndex map[string][]int) []string {
	keywords := make([]string, 0, len(invertedIndex))
	for k := range invertedIndex {
		keywords = append(keywords, k)
	}
	// 按数值排序（与项目中逻辑一致）
	sort.Slice(keywords, func(i, j int) bool {
		ki, _ := strconv.ParseInt(keywords[i], 10, 64)
		kj, _ := strconv.ParseInt(keywords[j], 10, 64)
		return ki < kj
	})
	return keywords
}

// 生成测试用的更新位图
func generateUpdateBitmap(docIDs []int) *big.Int {
	bs := big.NewInt(0)
	for _, docID := range docIDs {
		bs.SetBit(bs, docID, 1)
	}
	return bs
}

// 写入测试结果到文件
func writeResult(content string) error {
	f, err := os.OpenFile(resultFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content + "\n")
	return err
}

// TestUpdatePerformance 测试索引构建及更新性能对比（仅保留更新平均耗时日志）
func TestUpdatePerformance1(t *testing.T) {
	// 设置参数
	files := []string{
		//"dataset/Gowalla_invertedIndex_new_5000.txt",
		//"dataset/Gowalla_invertedIndex_new_10000.txt",
		//"dataset/Gowalla_invertedIndex_new_15000.txt",
		"dataset/Gowalla_invertedIndex_new_20000.txt",
		//"dataset/Gowalla_invertedIndex_new_25000.txt",
	}

	//indexNum := []int{5000}
	//indexNum := []int{10000}
	//indexNum := []int{15000}
	indexNum := []int{20000}
	//indexNum := []int{25000}
	FB_BsLen := 1 << 15
	LValues := []int{6424}
	resultsDir := "results"

	// 确保目录存在
	if err := ensureDirExists(resultsDir); err != nil {
		t.Fatalf("创建结果目录失败: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	for fileIndex, file := range files {
		// 加载索引
		invertedIndex, err := loadInvertedIndex(file)
		if err != nil {
			t.Fatalf("无法加载文件 %s: %v", file, err)
		}

		// 提取关键词
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			t.Fatalf("文件 %s 中未找到关键词", file)
		}

		// 排序关键词（无日志）
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()

		for _, L := range LValues {
			// 初始化方案
			ours := OurScheme.Setup(L)
			fbDsseParams := FB_RSSE.Setup(FB_BsLen)

			// 构建索引（无日志）
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme 构建索引失败: %v", err)
			}
			buildOursDur := time.Since(startTime).Nanoseconds()

			startTime = time.Now()
			err = fbDsseParams.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("FB_DSSE 构建索引失败: %v", err)
			}
			buildFBDur := time.Since(startTime).Nanoseconds()

			// OurScheme 更新测试
			const updateTotalTimes = 500
			var totalOursDur int64 = 0
			for updateRound := 0; updateRound < updateTotalTimes; updateRound++ {
				kw := strconv.Itoa(keywords[rand.Intn(len(keywords))])
				docIDs := generateRandomDocIDsBigInt(1)

				start := time.Now()
				if err := ours.Update(kw, docIDs); err != nil {
					t.Errorf("OurScheme 第 %d 次更新失败: %v", updateRound+1, err)
					continue
				}
				totalOursDur += time.Since(start).Nanoseconds()
			}
			avgOursDur := totalOursDur / updateTotalTimes
			// 仅保留平均耗时日志
			t.Logf("OurScheme 关键字数量=%d, 平均更新耗时=%d 纳秒", indexNum[fileIndex], avgOursDur)

			// FB_DSSE 更新测试
			var totalFBDur int64 = 0
			for updateRound := 0; updateRound < updateTotalTimes; updateRound++ {
				kw := strconv.Itoa(keywords[rand.Intn(len(keywords))])
				docIDs := generateRandomDocIDsBigInt(1)
				if len(docIDs) == 0 {
					t.Errorf("FB_DSSE 第 %d 次更新: 文档ID为空", updateRound+1)
					continue
				}

				start := time.Now()
				if err := fbDsseParams.UpdateBigInt(kw, docIDs[0]); err != nil {
					t.Errorf("FB_DSSE 第 %d 次更新失败: %v", updateRound+1, err)
					continue
				}
				totalFBDur += time.Since(start).Nanoseconds()
			}
			avgFBDur := totalFBDur / updateTotalTimes
			// 仅保留平均耗时日志
			t.Logf("FB_DSSE 关键字数量=%d, 平均更新耗时=%d 纳秒", indexNum[fileIndex], avgFBDur)

			// 保存结果（无日志）
			result := map[string]interface{}{
				"keyword_count":      indexNum[fileIndex],
				"L":                  L,
				"sort_duration":      sortDuration,
				"build_ours":         buildOursDur,
				"build_fb":           buildFBDur,
				"update_ours_total":  totalOursDur,
				"update_ours_avg":    avgOursDur,
				"update_fb_total":    totalFBDur,
				"update_fb_avg":      avgFBDur,
				"update_rounds":      updateTotalTimes,
				"doc_ids_per_update": 10,
			}
			if err := saveResult(resultsDir, indexNum[fileIndex], L, result); err != nil {
				t.Errorf("保存结果失败: %v", err)
			}
		}
	}
}
func TestUpdatePerformance(t *testing.T) {
	// 设置参数
	files := []string{
		"dataset/Gowalla_invertedIndex_new_5000.txt",
		"dataset/Gowalla_invertedIndex_new_10000.txt",
		"dataset/Gowalla_invertedIndex_new_15000.txt",
		"dataset/Gowalla_invertedIndex_new_20000.txt",
		"dataset/Gowalla_invertedIndex_new_25000.txt",
	}

	indexNum := []int{5000, 10000, 15000, 20000, 25000}
	FB_BsLen := 15
	LValues := []int{6424}
	resultsDir := "results"

	// 确保目录存在
	if err := ensureDirExists(resultsDir); err != nil {
		t.Fatalf("创建结果目录失败: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	for fileIndex, file := range files {
		t.Logf("开始处理数据集：%s（关键字数量：%d）", file, indexNum[fileIndex])

		// 加载索引
		invertedIndex, err := loadInvertedIndex(file)
		if err != nil {
			t.Fatalf("无法加载文件 %s: %v", file, err)
		}

		// 提取关键词
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			t.Fatalf("文件 %s 中未找到关键词", file)
		}

		// 排序关键词
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()

		for _, L := range LValues {
			t.Logf("开始处理 L值：%d 的测试流程", L)

			// 初始化方案
			ours := OurScheme.Setup(L)
			fbDsseParams := FB_RSSE.Setup(FB_BsLen)

			// --------------------------
			// 索引构建阶段 - 插入流程日志
			// --------------------------
			t.Logf("开始索引构建流程（关键字数量：%d，L值：%d）", indexNum[fileIndex], L)

			// OurScheme 构建索引
			t.Logf("开始 OurScheme 索引构建...")
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme 构建索引失败: %v", err)
			}
			buildOursDur := time.Since(startTime).Nanoseconds()
			t.Logf("OurScheme 索引构建完成，耗时：%d 纳秒", buildOursDur)

			// FB_DSSE 构建索引
			t.Logf("开始 FB_DSSE 索引构建...")
			startTime = time.Now()
			err = fbDsseParams.BuildIndexMock(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("FB_DSSE 构建索引失败: %v", err)
			}
			buildFBDur := time.Since(startTime).Nanoseconds()
			t.Logf("FB_DSSE 索引构建完成，耗时：%d 纳秒", buildFBDur)
			t.Logf("索引构建流程全部完成（关键字数量：%d，L值：%d）", indexNum[fileIndex], L)

			// --------------------------
			// 更新测试阶段 - 插入流程日志
			// --------------------------
			const updateTotalTimes = 500
			t.Logf("开始更新测试流程（关键字数量：%d，L值：%d，总更新次数：%d）", indexNum[fileIndex], L, updateTotalTimes)

			// OurScheme 更新测试
			t.Logf("开始 OurScheme 更新测试...")
			var totalOursDur int64 = 0
			for updateRound := 0; updateRound < updateTotalTimes; updateRound++ {
				kw := strconv.Itoa(keywords[rand.Intn(len(keywords))])
				docIDs := generateRandomDocIDsBigInt(100)

				start := time.Now()
				if err := ours.Update(kw, docIDs); err != nil {
					t.Errorf("OurScheme 第 %d 次更新失败: %v", updateRound+1, err)
					continue
				}
				totalOursDur += time.Since(start).Nanoseconds()
			}
			avgOursDur := totalOursDur / updateTotalTimes
			t.Logf("OurScheme 更新测试完成，关键字数量=%d, 平均更新耗时=%d 纳秒", indexNum[fileIndex], avgOursDur)

			// FB_DSSE 更新测试
			t.Logf("开始 FB_DSSE 更新测试...")
			var totalFBDur int64 = 0
			for updateRound := 0; updateRound < updateTotalTimes; updateRound++ {
				kw := strconv.Itoa(keywords[rand.Intn(len(keywords))])
				docIDs := generateRandomDocIDsBigInt(1)
				if len(docIDs) == 0 {
					t.Errorf("FB_DSSE 第 %d 次更新: 文档ID为空", updateRound+1)
					continue
				}

				start := time.Now()
				if err := fbDsseParams.UpdateBigInt(kw, docIDs[0]); err != nil {
					t.Errorf("FB_DSSE 第 %d 次更新失败: %v", updateRound+1, err)
					continue
				}
				totalFBDur += time.Since(start).Nanoseconds()
			}
			avgFBDur := totalFBDur / updateTotalTimes
			t.Logf("FB_DSSE 更新测试完成，关键字数量=%d, 平均更新耗时=%d 纳秒", indexNum[fileIndex], avgFBDur)
			t.Logf("更新测试流程全部完成（关键字数量：%d，L值：%d）", indexNum[fileIndex], L)

			// 保存结果
			result := map[string]interface{}{
				"keyword_count":      indexNum[fileIndex],
				"L":                  L,
				"sort_duration":      sortDuration,
				"build_ours":         buildOursDur,
				"build_fb":           buildFBDur,
				"update_ours_total":  totalOursDur,
				"update_ours_avg":    avgOursDur,
				"update_fb_total":    totalFBDur,
				"update_fb_avg":      avgFBDur,
				"update_rounds":      updateTotalTimes,
				"doc_ids_per_update": 10,
			}
			if err := saveResult(resultsDir, indexNum[fileIndex], L, result); err != nil {
				t.Errorf("保存结果失败: %v", err)
			}

			t.Logf("当前L值（%d）测试流程全部完成（关键字数量：%d）", L, indexNum[fileIndex])
		}

		t.Logf("当前数据集（%s）测试流程全部完成", file)
	}
	t.Log("所有测试流程全部完成")
}

// 辅助函数：确保目录存在（如果不存在则创建）
func ensureDirExists(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return err
}

// 辅助函数：保存测试结果（根据实际需求实现）
func saveResult(dir string, count, L int, data map[string]interface{}) error {
	// 示例：可将结果保存为JSON文件
	// 实际实现需引入encoding/json和os包
	return nil
}

// generateRandomDocIDs 生成指定数量的随机文档ID（int类型），仅依赖math/rand
func generateRandomDocIDs(count int) []int {
	docIDs := make([]int, count)
	// 初始化随机种子（全局只需初始化一次，若外部已初始化可移除）
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		// 生成1~1e3范围的随机int（确保非零）
		randInt := rand.Intn(1e3) // 0~999
		if randInt == 0 {
			randInt = 1 // 避免0值，确保ID有效
		}
		docIDs[i] = randInt
	}
	return docIDs
}

// generateRandomDocIDs 生成指定数量的随机文档ID（big.Int类型），仅依赖math/rand
func generateRandomDocIDsBigInt(count int) []*big.Int {
	docIDs := make([]*big.Int, count)
	// 初始化math/rand随机种子（确保每次运行生成不同随机序列）
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		// 步骤1：用math/rand生成int64范围的随机数（0 ~ 1e9-1）
		//randInt64 := int64(rand.Intn(1e9))
		randInt64 := int64(rand.Intn(1e6))
		// 步骤2：将int64转为big.Int类型
		randBigInt := big.NewInt(randInt64)
		// 步骤3：确保文档ID非零（若随机数为0则加1，否则保持原数）
		if randBigInt.Sign() == 0 {
			randBigInt.Add(randBigInt, big.NewInt(1))
		}
		docIDs[i] = randBigInt
	}
	return docIDs
}
