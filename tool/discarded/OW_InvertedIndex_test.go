package discarded

//
//import (
//	"EfficientAndLowStroageSSE/config"
//	"fmt"
//	"sort"
//	"testing"
//)
//
//func TestBuildOWInvertedIndexWithFile(t *testing.T) {
//	// 使用实际文件路径
//	filePath := config.FilePath
//
//	// 调用 BuildInvertedIndex 函数生成基础倒排索引
//	invertedIndex, err := BuildInvertedIndex(filePath)
//	if err != nil {
//		t.Fatalf("BuildInvertedIndex failed: %v", err)
//	}
//
//	// 调用 BuildOWInvertedIndex 函数生成 Order-Weighted Inverted Index
//	owIndex, err := BuildOWInvertedIndex(invertedIndex)
//	if err != nil {
//		t.Fatalf("BuildOWInvertedIndex failed: %v", err)
//	}
//
//	// 按关键字从小到大排序打印
//	keys := make([]string, 0, len(owIndex.Index))
//	for key := range owIndex.Index {
//		keys = append(keys, key)
//	}
//	sort.Strings(keys) // 按字符串升序排序
//
//	// 打印生成的 Order-Weighted Inverted Index 的内容
//	t.Logf("Order-Weighted Inverted Index (sorted by keys):\n")
//	for _, key := range keys {
//		t.Logf("Keyword: %s, Cumulative Identifiers: %v", key, owIndex.Index[key])
//	}
//
//	// 验证 owIndex 的内容是否合理（可根据实际数据添加断言）
//	if len(owIndex.Index) == 0 {
//		t.Errorf("BuildOWInvertedIndex failed: expected non-zero entries, got %d", len(owIndex.Index))
//	}
//}
//
//func TestBuildOWInvertedIndex(t *testing.T) {
//	// 示例倒排索引
//	invertedIndex := map[string][]int{
//		"6.0000":  {2, 8},
//		"18.0000": {7},
//		"21.0000": {9, 13},
//		"28.0000": {10},
//		"33.0000": {6, 15},
//	}
//
//	// 构建 Order-Weighted Inverted Index
//	owIndex, err := BuildOWInvertedIndex(invertedIndex)
//	if err != nil {
//		fmt.Printf("Error: %v\n", err)
//		return
//	}
//
//	// 按关键字从小到大排序
//	keys := make([]string, 0, len(owIndex.Index))
//	for key := range owIndex.Index {
//		keys = append(keys, key)
//	}
//	sort.Strings(keys) // 按字符串升序排序
//
//	// 打印 Order-Weighted Inverted Index
//	fmt.Println("Order-Weighted Inverted Index (sorted by keys):")
//	for _, key := range keys {
//		fmt.Printf("Keyword: %s, Cumulative Identifiers: %v\n", key, owIndex.Index[key])
//	}
//
//	// 执行范围查询
//	result, err := owIndex.RangeQuery("6.0000", "28.0000")
//	if err != nil {
//		fmt.Printf("Error: %v\n", err)
//		return
//	}
//
//	// 打印范围查询结果
//	fmt.Printf("Range Query Result [6.0000, 28.0000]: %v\n", result)
//}
