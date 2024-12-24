package FB_RSSE_test

import (
	"VolumeHidingSSE/FB_RSSE"
	"reflect"
	"testing"
)

// 测试FB_RSSE方法
func TestFB_RSSE(t *testing.T) {
	// 模拟invertedIndex数据
	invertedIndex := map[string][]int{
		"1": {1, 3},
		"2": {4, 2, 5},
		"3": {6, 7},
		"4": {8, 9, 10, 11},
		"5": {12, 13, 14, 15, 16},
	}

	// 位图长度（可以调整）
	L := 1 << 20

	// 创建FB_RSSE对象
	rsse, err := FB_RSSE.NewFB_RSSE(invertedIndex, L)
	if err != nil {
		t.Fatalf("Error initializing RSSE: %v", err)
	}

	// 生成EDB
	err = rsse.GenEDB()
	if err != nil {
		t.Fatalf("Error generating EDB: %v", err)
	}

	// 生成查询令牌
	queryRange := [2]string{"2", "4"} // 示例查询范围
	tokensList := rsse.GenToken(queryRange)
	t.Logf("Generated tokens: %v", tokensList)

	// 执行搜索
	searchResult := rsse.Search(tokensList)
	t.Logf("Search result: %v", searchResult)

	// 生成ID
	finalResult := rsse.GenIds(searchResult, tokensList)
	t.Logf("Final result (IDs): %v", finalResult)

	// 验证结果
	expectedSearchResult := []int{3, 4, 5, 6, 7, 8, 9, 10, 11}
	if !reflect.DeepEqual(finalResult, expectedSearchResult) {
		t.Errorf("Search result mismatch: expected %v, got %v", expectedSearchResult, finalResult)
	}
}
