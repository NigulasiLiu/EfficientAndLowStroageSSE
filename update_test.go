package main

import (
	"EfficientAndLowStroageSSE/FB_RSSE"
	"EfficientAndLowStroageSSE/VH_RSSE/OurScheme"
	"fmt"
	"math/big"
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
	largeM       = 335349                             // 大关键字数量
	datasetDir   = "dataset/"                         // 数据集目录（.gitignore中已忽略，实际运行需确保存在）
	fileTemplate = "Gowalla_invertedIndex_new_%d.txt" // 数据集文件名模板
)

// 待测试的关键字数量
var testMs = []int{5000, 10000, 15000, 20000}

// 提取并排序关键字
func getSortedKeywords(invertedIndex map[string][]int) []string {
	keywords := make([]string, 0, len(invertedIndex))
	for k := range invertedIndex {
		keywords = append(keywords, k)
	}
	// 按数值排序（与项目中sortKeywords逻辑一致）
	sort.Slice(keywords, func(i, j int) bool {
		ki, _ := strconv.ParseInt(keywords[i], 10, 64)
		kj, _ := strconv.ParseInt(keywords[j], 10, 64)
		return ki < kj
	})
	return keywords
}

// 生成测试用的更新位图（模拟文档ID对应的位图）
func generateUpdateBitmap(docIDs []int) *big.Int {
	bs := big.NewInt(0)
	for _, docID := range docIDs {
		bs.SetBit(bs, docID, 1) // 将文档ID对应位置设为1
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

// TestUpdatePerformance 测试Update函数耗时
func TestUpdatePerformance(t *testing.T) {
	// 初始化结果文件
	_, err := os.Create(resultFile)
	if err != nil {
		t.Fatalf("创建结果文件失败: %v", err)
	}
	writeResult("Update函数耗时测试结果 (单位: ms)")
	writeResult("----------------------------------------")

	// 测试m={5000,10000,15000,20000}的情况（两种方案）
	for _, m := range testMs {
		filePath := fmt.Sprintf("%s%s", datasetDir, fmt.Sprintf(fileTemplate, m))
		t.Logf("开始测试 m=%d, 文件: %s", m, filePath)

		// 加载数据
		invertedIndex, err := loadInvertedIndex(filePath)
		if err != nil {
			t.Fatalf("加载索引失败: %v", err)
		}
		sortedKeywords := getSortedKeywords(invertedIndex)
		if len(sortedKeywords) < m {
			t.Fatalf("实际关键字数量不足 %d, 实际: %d", m, len(sortedKeywords))
		}
		sortedKeywords = sortedKeywords[:m] // 截取前m个关键字

		// 1. 测试FB_RSSE的Update
		fb := FB_RSSE.Setup(1 << 15) // 使用项目中默认的BsLen（1<<15）
		// 先构建索引
		err = fb.BuildIndex(invertedIndex, sortedKeywords)
		if err != nil {
			t.Fatalf("FB_RSSE构建索引失败: %v", err)
		}
		// 选择一个关键字及其文档ID生成更新位图
		targetKeyword := sortedKeywords[len(sortedKeywords)/2] // 取中间关键字
		docIDs := invertedIndex[targetKeyword]                 // 获取该关键字对应的文档ID列表
		updateBitmap := generateUpdateBitmap(docIDs)           // 生成位图
		// 测试Update耗时（修正参数：传入keyword和bitmap）
		start := time.Now()
		err = fb.Update(targetKeyword, updateBitmap)
		if err != nil {
			t.Fatalf("FB_RSSE更新失败: %v", err)
		}
		fbDuration := time.Since(start).Milliseconds()
		result := fmt.Sprintf("FB_RSSE, m=%d, 耗时: %d ms", m, fbDuration)
		t.Log(result)
		writeResult(result)

		// 2. 测试OurScheme的Update
		ours := OurScheme.Setup(L)
		// 先构建索引
		err = ours.BuildIndex(invertedIndex, sortedKeywords)
		if err != nil {
			t.Fatalf("OurScheme构建索引失败: %v", err)
		}
		// 选择同一个关键字进行更新（OurScheme的Update仅需keyword）
		start = time.Now()
		err = ours.Update(targetKeyword)
		if err != nil {
			t.Fatalf("OurScheme更新失败: %v", err)
		}
		ourDuration := time.Since(start).Milliseconds()
		result = fmt.Sprintf("OurScheme, m=%d, 耗时: %d ms", m, ourDuration)
		t.Log(result)
		writeResult(result)
	}

	// 测试m=335349的情况（仅OurScheme）
	t.Logf("开始测试大数量 m=%d", largeM)
	filePath := fmt.Sprintf("%s%s", datasetDir, fmt.Sprintf(fileTemplate, largeM))
	invertedIndex, err := loadInvertedIndex(filePath)
	if err != nil {
		t.Fatalf("加载大索引失败: %v", err)
	}
	sortedKeywords := getSortedKeywords(invertedIndex)
	if len(sortedKeywords) < largeM {
		t.Fatalf("实际关键字数量不足 %d, 实际: %d", largeM, len(sortedKeywords))
	}
	sortedKeywords = sortedKeywords[:largeM]

	// 构建索引并测试Update
	ours := OurScheme.Setup(L)
	err = ours.BuildIndex(invertedIndex, sortedKeywords)
	if err != nil {
		t.Fatalf("OurScheme构建大索引失败: %v", err)
	}
	targetKeyword := sortedKeywords[len(sortedKeywords)/2] // 取中间关键字
	start := time.Now()
	err = ours.Update(targetKeyword)
	if err != nil {
		t.Fatalf("OurScheme大数量更新失败: %v", err)
	}
	largeDuration := time.Since(start).Milliseconds()
	result := fmt.Sprintf("OurScheme, m=%d, 耗时: %d ms", largeM, largeDuration)
	t.Log(result)
	writeResult(result)

	writeResult("----------------------------------------")
	writeResult("测试完成")
}
