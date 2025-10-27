package main

import (
	"EfficientAndLowStroageSSE/FB_RSSE"
	"EfficientAndLowStroageSSE/VH_RSSE/OurScheme"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"
)

// 测试配置
const (
	L            = 6424                               // 固定L值
	resultFile   = "update_time_results.txt"          // 结果输出文件
	largeM       = 335349                             // 大关键字数量
	datasetDir   = "dataset/"                         // 数据集目录
	fileTemplate = "Gowalla_invertedIndex_new_%d.txt" // 数据集文件名模板
)

// 待测试的关键字数量
var testMs = []int{5000, 10000, 15000, 20000}

//// 加载倒排索引（复用项目中已有的逻辑）
//func loadInvertedIndex(filePath string) (map[string][]int, error) {
//	invertedIndex := make(map[string][]int)
//	file, err := os.Open(filePath)
//	if err != nil {
//		return nil, fmt.Errorf("打开文件失败: %v", err)
//	}
//	defer file.Close()
//
//	scanner := bufio.NewScanner(file)
//	for scanner.Scan() {
//		line := strings.TrimSpace(scanner.Text())
//		if line == "" {
//			continue
//		}
//		parts := strings.Fields(line)
//		if len(parts) < 2 {
//			return nil, fmt.Errorf("无效行格式: %s", line)
//		}
//		keyword := parts[0]
//		rowIDs := make([]int, 0, len(parts)-1)
//		for _, part := range parts[1:] {
//			id, err := strconv.Atoi(part)
//			if err != nil {
//				continue // 忽略无效ID
//			}
//			rowIDs = append(rowIDs, id)
//		}
//		invertedIndex[keyword] = rowIDs
//	}
//	return invertedIndex, scanner.Err()
//}

// 提取并排序关键字
func getSortedKeywords(invertedIndex map[string][]int) []string {
	keywords := make([]string, 0, len(invertedIndex))
	for k := range invertedIndex {
		keywords = append(keywords, k)
	}
	sort.Strings(keywords)
	return keywords
}

// 生成测试用的更新数据（随机选择关键字和文档ID）
func generateUpdateData(sortedKeywords []string) (string, int) {
	if len(sortedKeywords) == 0 {
		return "", 0
	}
	// 随机选择一个关键字
	randIdx := time.Now().UnixNano() % int64(len(sortedKeywords))
	keyword := sortedKeywords[randIdx]
	// 生成随机文档ID
	docID := int(time.Now().UnixNano() % 1000000)
	return keyword, docID
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
		fb := FB_RSSE.Setup(1 << 15) // 使用项目中默认的BsLen
		// 先构建索引
		err = fb.BuildIndex(invertedIndex, sortedKeywords)
		if err != nil {
			t.Fatalf("FB_RSSE构建索引失败: %v", err)
		}
		// 生成更新数据
		keyword, docID := generateUpdateData(sortedKeywords)
		if keyword == "" {
			t.Skip("跳过空关键字更新")
		}
		// 测试Update耗时
		start := time.Now()
		err = fb.Update(keyword, docID) // 假设Update函数签名为(keyword string, docID int) error
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
		// 生成更新数据（重新生成避免冲突）
		keyword, docID = generateUpdateData(sortedKeywords)
		if keyword == "" {
			t.Skip("跳过空关键字更新")
		}
		// 测试Update耗时
		start = time.Now()
		err = ours.Update(keyword, docID) // 假设Update函数签名为(keyword string, docID int) error
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
	keyword, docID := generateUpdateData(sortedKeywords)
	if keyword == "" {
		t.Skip("跳过空关键字更新")
	}
	start := time.Now()
	err = ours.Update(keyword, docID)
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
