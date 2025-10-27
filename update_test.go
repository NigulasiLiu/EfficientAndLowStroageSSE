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

// TestUpdatePerformance 测试索引构建及更新性能对比（保留代码段2，修改代码段3）
func TestUpdatePerformance(t *testing.T) {
	// 设置参数（补充25000关键字的数据集）
	files := []string{
		"dataset/Gowalla_invertedIndex_new_5000.txt",
		"dataset/Gowalla_invertedIndex_new_10000.txt",
		"dataset/Gowalla_invertedIndex_new_15000.txt",
		"dataset/Gowalla_invertedIndex_new_20000.txt",
		"dataset/Gowalla_invertedIndex_new_25000.txt", // 新增25000关键字数据集
	}

	indexNum := []int{5000, 10000, 15000, 20000, 25000} // 对应关键字数量
	FB_BsLen := 1 << 15                                 // FB_DSSE 参数
	LValues := []int{6424}                              // L值范围
	resultsDir := "results"                             // 结果存储目录

	// 确保结果目录存在
	if err := ensureDirExists(resultsDir); err != nil {
		t.Fatalf("创建结果目录失败: %v", err)
	}

	// 初始化随机种子（确保每次测试随机值可复现）
	rand.Seed(time.Now().UnixNano())

	// 遍历每个文件进行测试
	for fileIndex, file := range files {
		// 加载倒排索引
		invertedIndex, err := loadInvertedIndex(file)
		if err != nil {
			t.Fatalf("无法加载文件 %s: %v", file, err)
		}

		// 提取关键词
		keywords := extractKeywordsFromIndex(invertedIndex)
		if len(keywords) == 0 {
			t.Fatalf("文件 %s 中未找到关键词", file)
		}
		t.Logf("当前测试数据集: %d 个关键字", indexNum[fileIndex])

		// 测量关键词排序时间
		startTime := time.Now()
		sortedKeywords := sortKeywords(invertedIndex)
		sortDuration := time.Since(startTime).Nanoseconds()
		t.Logf("关键词排序耗时: %d 纳秒", sortDuration)

		// 遍历每个L值
		for _, L := range LValues {
			// 初始化两种方案
			ours := OurScheme.Setup(L)
			fbDsseParams := FB_RSSE.Setup(FB_BsLen) // FB_DSSE的系统参数对象

			// --------------------------
			// 1. 构建索引（OurScheme）
			// --------------------------
			startTime = time.Now()
			err = ours.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("OurScheme 构建索引失败: %v", err)
			}
			buildOursDur := time.Since(startTime).Nanoseconds()
			t.Logf("OurScheme 构建索引耗时: %d 纳秒", buildOursDur)

			// --------------------------
			// 2. 构建索引（FB_DSSE）
			// --------------------------
			startTime = time.Now()
			err = fbDsseParams.BuildIndex(invertedIndex, sortedKeywords)
			if err != nil {
				t.Fatalf("FB_DSSE 构建索引失败: %v", err)
			}
			buildFBDur := time.Since(startTime).Nanoseconds()
			t.Logf("FB_DSSE 构建索引耗时: %d 纳秒", buildFBDur)

			// --------------------------
			// 3. 执行更新逻辑（OurScheme）- 代码段3修改后
			// 新签名：func (sp *OurScheme) Update(w string, docID []*big.Int) error
			// --------------------------
			const updateTotalTimes = 10 // 固定更新10次，取平均值
			t.Logf("开始执行 OurScheme 更新测试，共更新 %d 次", updateTotalTimes)

			var totalOursDur int64 = 0
			// 执行10次更新，累计总耗时
			for updateRound := 0; updateRound < updateTotalTimes; updateRound++ {
				// 准备单次更新参数
				kw := keywords[rand.Intn(len(keywords))] // 随机选关键字
				docIDs := generateRandomDocIDs(10)       // 生成10个随机文档ID（可调整数量）

				// 单次更新计时
				start := time.Now()
				if err := ours.Update(strconv.Itoa(kw), docIDs); err != nil {
					t.Errorf("OurScheme 第 %d 次更新关键字 %s 失败: %v", updateRound+1, kw, err)
					continue // 失败仍继续后续测试，避免单次失败中断整体流程
				}
				totalOursDur += time.Since(start).Nanoseconds()
			}

			// 计算平均耗时
			avgOursDur := totalOursDur / updateTotalTimes
			t.Logf("OurScheme 完成 %d 次更新，总耗时: %d 纳秒，单次平均耗时: %d 纳秒",
				updateTotalTimes, totalOursDur, avgOursDur)

			// --------------------------
			// 4. 执行更新逻辑（FB_DSSE）- 代码段3修改后
			// 新签名：func (sp *SystemParameters) Update(w string, docID []*big.Int) error
			// --------------------------
			t.Logf("开始执行 FB_DSSE 更新测试，共更新 %d 次", updateTotalTimes)

			var totalFBDur int64 = 0
			// 执行10次更新，累计总耗时
			for updateRound := 0; updateRound < updateTotalTimes; updateRound++ {
				// 准备单次更新参数
				kw := keywords[rand.Intn(len(keywords))]
				docIDs := generateRandomDocIDs(10)

				// 单次更新计时
				start := time.Now()
				if err := fbDsseParams.Update(strconv.Itoa(kw), docIDs[0]); err != nil {
					t.Errorf("FB_DSSE 第 %d 次更新关键字 %s 失败: %v", updateRound+1, kw, err)
					continue
				}
				totalFBDur += time.Since(start).Nanoseconds()
			}

			// 计算平均耗时
			avgFBDur := totalFBDur / updateTotalTimes
			t.Logf("FB_DSSE 完成 %d 次更新，总耗时: %d 纳秒，单次平均耗时: %d 纳秒",
				updateTotalTimes, totalFBDur, avgFBDur)

			// --------------------------
			// 5. 保存结果（新增平均耗时字段）
			// --------------------------
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
				"doc_ids_per_update": 10, // 每次更新的文档ID数量
			}
			if err := saveResult(resultsDir, indexNum[fileIndex], L, result); err != nil {
				t.Errorf("保存结果失败: %v", err)
			}
		}
	}
	t.Log("所有测试完成")
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

// generateRandomDocIDs 生成指定数量的随机文档ID（big.Int类型），仅依赖math/rand
func generateRandomDocIDs(count int) []*big.Int {
	docIDs := make([]*big.Int, count)
	// 初始化math/rand随机种子（确保每次运行生成不同随机序列）
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		// 步骤1：用math/rand生成int64范围的随机数（0 ~ 1e9-1）
		randInt64 := int64(rand.Intn(1e9))
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
