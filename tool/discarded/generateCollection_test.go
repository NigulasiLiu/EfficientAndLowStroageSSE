package discarded

//
//import (
//	"VolumeHidingSSE/config"
//	"testing"
//)
//
//func TestGenerateCollection(t *testing.T) {
//	// 示例 Order-Weighted Inverted Index
//	owIndex := &OWInvertedIndex{
//		Index: map[string][]int{
//			"6.0000":  {1, 2, 3, 4},
//			"18.0000": {5, 6, 7, 8, 9, 10},
//			"21.0000": {11, 12, 13, 14},
//			"28.0000": {15, 16, 17},
//			"33.0000": {18, 19, 20, 21, 22},
//		},
//	}
//
//	// 调用 GenerateCollection 函数
//	collection, err := GenerateCollection(owIndex)
//	if err != nil {
//		t.Fatalf("GenerateCollection failed: %v", err)
//	}
//
//	// 打印生成的集合
//	t.Logf("Generated Collection: %+v\n", collection)
//
//	// 验证集合是否非空
//	if len(collection) == 0 {
//		t.Errorf("GenerateCollection failed: expected non-empty collection, got %d", len(collection))
//	}
//}
//
//func TestGenerateCollectionWithFile(t *testing.T) {
//	// 文件路径
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
//	// 调用 GenerateCollection 函数
//	collection, err := GenerateCollection(owIndex)
//	if err != nil {
//		t.Fatalf("GenerateCollection failed: %v", err)
//	}
//
//	// 打印前 100 个集合
//	t.Logf("Generated Collection (first 100 entries):\n")
//	limit := 100
//	if len(collection) < limit {
//		limit = len(collection)
//	}
//	for i := 0; i < limit; i++ {
//		col := collection[i]
//		t.Logf("C[%d]: [%d, %d]", i, col[0], col[1])
//	}
//
//	// 验证集合是否非空
//	if len(collection) == 0 {
//		t.Errorf("GenerateCollection failed: expected non-empty collection, got %d", len(collection))
//	}
//}
