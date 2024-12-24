package BuildIndex

import (
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func TestBuildIndexFromInvertedIndex(t *testing.T) {
	// Mock inverted index
	invertedIndex := map[string][]int{
		"key1": {1, 2},
		"key2": {3, 4, 5},
		"key3": {6, 7},
		"key4": {8, 9, 10, 11},
		"key5": {12, 13, 14, 15, 16},
	}

	// Parameters for the test
	L := 10

	// Expected results
	expectedEDB := map[string]string{
		"key1": "1100000000",
		"key2": "1111100000",
		"key3": "1111111000",
		"key4": "1111000000",
		"key5": "1111111110",
	}
	expectedClusterFlist := [][]int{
		{1, 2, 3, 4, 5, 6, 7},
		{8, 9, 10, 11, 12, 13, 14, 15, 16},
	}
	expectedClusterVolume := []int{7, 9}
	expectedClusterKlist := [][]string{
		{"key1", "key2", "key3"},
		{"key4", "key5"},
	}

	expectedLocalTree := map[string]string{
		"00": "key3",
		"01": "key5",
		"0":  "key3",
	}
	// Call BuildIndexFromInvertedIndex
	EDB, localTree, clusters, err := BuildIndexFromInvertedIndex(invertedIndex, L)

	// Check for errors
	if err != nil {
		t.Fatalf("BuildIndexFromInvertedIndex returned an error: %v", err)
	}

	// Validate EDB
	for key, expectedValue := range expectedEDB {
		actualValue, ok := EDB[key]
		if !ok {
			t.Errorf("EDB is missing key: %s", key)
		} else if actualValue != expectedValue {
			t.Errorf("EDB value mismatch for key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Validate localTree
	if !reflect.DeepEqual(localTree, expectedLocalTree) {
		t.Errorf("localTree mismatch: expected %v, got %v", expectedLocalTree, localTree)
	}

	// Validate clusters
	actualClusterFlist := clusters[0]["flist"].([][]int)
	actualClusterVolume := clusters[0]["volume"].([]int)
	actualClusterKlist := clusters[0]["klist"].([][]string)

	if !reflect.DeepEqual(actualClusterFlist, expectedClusterFlist) {
		t.Errorf("ClusterFlist mismatch: expected %v, got %v", expectedClusterFlist, actualClusterFlist)
	}
	if !reflect.DeepEqual(actualClusterVolume, expectedClusterVolume) {
		t.Errorf("ClusterVolume mismatch: expected %v, got %v", expectedClusterVolume, actualClusterVolume)
	}
	if !reflect.DeepEqual(actualClusterKlist, expectedClusterKlist) {
		t.Errorf("ClusterKlist mismatch: expected %v, got %v", expectedClusterKlist, actualClusterKlist)
	}
}

func TestBuildIndexLogic(t *testing.T) {
	// 模拟倒排索引
	invertedIndex := map[string][]int{
		"6.0000":  {2, 8},
		"18.0000": {7},
		"21.0000": {9, 13},
		"28.0000": {10},
		"33.0000": {6, 15},
	}

	// 定义分区大小 L
	L := 3

	// Step 2: 初始化参数
	IStar := make(map[string][]int)        // I^*，空倒排索引
	C := make([]map[string]interface{}, 0) // C 集合
	ct := 1                                // 分区编号
	vlist := make([]string, 0)             // 存储当前分区的关键词
	var offset int                         // 偏移量

	// Step 3: 获取所有关键词并排序
	keywords := make([]string, 0, len(invertedIndex))
	for keyword := range invertedIndex {
		keywords = append(keywords, keyword)
	}

	// 将关键词按照 float64 转换后进行排序
	sort.Slice(keywords, func(i, j int) bool {
		ki, _ := strconv.ParseFloat(keywords[i], 64)
		kj, _ := strconv.ParseFloat(keywords[j], 64)
		return ki < kj
	})

	// Step 4: 遍历关键词
	lastKeyword := ""
	for i := 0; i < len(keywords); i++ {
		keyword := keywords[i]
		postings := invertedIndex[keyword]

		// 检查是否可以继续添加到当前分区
		if len(IStar[lastKeyword])+len(postings) <= L {
			// 合并到 IStar[keyword]
			if lastKeyword != "" {
				IStar[keyword] = append(IStar[lastKeyword], postings...)
			} else {
				IStar[keyword] = append([]int{}, postings...)
			}
			vlist = append(vlist, keyword) // 将关键词加入分区
		} else {
			// 超过分区大小时，处理当前分区
			cMax := max(vlist) // 当前分区关键词的最大值
			cMin := min(vlist) // 当前分区关键词的最小值
			cMap := map[string]interface{}{
				"info": []int{len(IStar[lastKeyword]), offset},
				"v":    vlist,                   // 当前分区的关键词集合
				"max":  cMax,                    // 最大关键词
				"min":  cMin,                    // 最小关键词
				"ind":  len(IStar[lastKeyword]), // 当前分区的倒排索引大小
			}
			C = append(C, cMap)           // 添加到 C 集合
			IStar[keyword] = postings     // 新分区的倒排索引
			offset += len(IStar[keyword]) // 更新偏移量
			vlist = []string{keyword}     // 新分区的关键词
			ct++                          // 更新分区编号
		}

		// 更新 lastKeyword 为当前关键词
		lastKeyword = keyword
	}
	// 打印生成的 C 集合
	t.Logf("Generated Collection C:\n")
	for i, cMap := range C {
		t.Logf("Partition %d:", i+1)
		t.Logf("  Keywords: %+v", cMap["v"])
		t.Logf("  Max Keyword: %+v", cMap["max"])
		t.Logf("  Min Keyword: %+v", cMap["min"])
		t.Logf("  Partition Size (ind): %+v", cMap["ind"])
		t.Logf("  Offset Info: %+v", cMap["info"])
		t.Logf("----------------------")
	}
}
