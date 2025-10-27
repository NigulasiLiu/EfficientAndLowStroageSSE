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

// TestUpdatePerformance 测试Update函数耗时（单位：微秒）
func TestUpdatePerformance(t *testing.T) {
	// 初始化结果文件
	_, err := os.Create(resultFile)
	if err != nil {
		t.Fatalf("创建结果文件失败: %v", err)
	}
	writeResult("Update函数耗时测试结果 (单位: μs)")
	writeResult("----------------------------------------")

	// 测试小数据集（m=5000,10000,15000,20000）
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
		sortedKeywords = sortedKeywords[:m]

		// 1. 测试FB_RSSE的Update
		fb := FB_RSSE.Setup(1 << 15)
		err = fb.BuildIndex(invertedIndex, sortedKeywords)
		if err != nil {
			t.Fatalf("FB_RSSE构建索引失败: %v", err)
		}
		targetKeyword := sortedKeywords[len(sortedKeywords)/2]
		docIDs := invertedIndex[targetKeyword]
		updateBitmap := generateUpdateBitmap(docIDs)
		// 计时（单位：微秒）
		start := time.Now()
		err = fb.Update(targetKeyword, updateBitmap)
		if err != nil {
			t.Fatalf("FB_RSSE更新失败: %v", err)
		}
		fbDuration := time.Since(start).Microseconds()
		result := fmt.Sprintf("FB_RSSE, m=%d, 耗时: %d μs", m, fbDuration)
		t.Log(result)
		writeResult(result)

		// 2. 测试OurScheme的Update
		ours := OurScheme.Setup(L)
		err = ours.BuildIndex(invertedIndex, sortedKeywords)
		if err != nil {
			t.Fatalf("OurScheme构建索引失败: %v", err)
		}
		// 计时（单位：微秒）
		start = time.Now()
		err = ours.Update(targetKeyword)
		if err != nil {
			t.Fatalf("OurScheme更新失败: %v", err)
		}
		ourDuration := time.Since(start).Microseconds()
		result = fmt.Sprintf("OurScheme, m=%d, 耗时: %d μs", m, ourDuration)
		t.Log(result)
		writeResult(result)
	}

	// 注释掉大数据集测试
	// /*
	// t.Logf("开始测试大数量 m=%d", largeM)
	// ... 原大数据集测试代码 ...
	// */

	writeResult("----------------------------------------")
	writeResult("测试完成")
}
