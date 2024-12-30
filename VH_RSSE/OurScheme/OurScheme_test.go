package OurScheme

import (
	"EfficientAndLowStroageSSE/config"
	"EfficientAndLowStroageSSE/tool"
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestOurScheme_random(t *testing.T) {
	// 初始化查询范围的上下界
	minKey := 1
	maxKey := 5
	k := 10 // 随机生成 10 个查询范围

	// 模拟倒排索引
	invertedIndex := map[string][]int{
		"1": {1, 3},
		"2": {4, 2, 5},
		"3": {6, 7},
		"4": {8, 9, 10, 11},
		"5": {12, 13, 14, 15, 16},
	}

	// 设置 L 值
	L := 10
	sp := Setup(L)
	// 测量关键词排序时间
	startTime := time.Now()
	sortedKeywords := sortKeywords(invertedIndex) // 排序函数
	sortDuration := time.Since(startTime).Nanoseconds()
	t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
	// 构建索引
	err := sp.BuildIndex(invertedIndex, sortedKeywords)
	if err != nil {
		t.Fatalf("BuildIndex returned an error: %v", err)
	}

	// 打印 LocalTree 和 ClusterKlist 以供调试
	t.Logf("LocalTree: %v", sp.LocalTree)
	t.Logf("ClusterKlist: %v", sp.ClusterKlist)

	// ------------------------------
	// Search Test Section
	// ------------------------------

	// 随机生成 k 个查询范围
	for i := 0; i < k; i++ {
		// 随机生成查询范围
		startKey := rand.Intn(maxKey-minKey+1) + minKey
		endKey := rand.Intn(maxKey-startKey+1) + startKey

		queryRange := [2]string{strconv.Itoa(startKey), strconv.Itoa(endKey)}
		t.Logf("Query range %d: %v", i+1, queryRange)

		// 第 1 步：生成查询 token
		tokens, err := sp.GenToken(queryRange)
		if err != nil {
			t.Fatalf("GenToken returned an error: %v", err)
		}
		//t.Logf("Generated tokens for query range %d: %v", i+1, tokens)

		// 第 2 步：从加密数据库中查询加密结果
		searchResult := sp.SearchTokens(tokens)
		//t.Logf("Search result (encrypted) for query range %d: %v", i+1, searchResult)

		// 第 3 步：执行本地搜索
		actualSearchResult, err := sp.LocalSearch(searchResult, tokens)
		if err != nil {
			t.Fatalf("LocalSearch returned an error: %v", err)
		}
		//t.Logf("Search result (decrypted) for query range %d: %v", i+1, actualSearchResult)

		// 打印搜索结果，供手动验证
		fmt.Printf("Query range %d: %v, Decrypted result: %v\n", i+1, queryRange, actualSearchResult)
	}
}
func TestOurScheme_fix(t *testing.T) {
	queryRange := [2]string{"6", "9"} // Example query range: ["2", "4"]
	// Mock inverted index
	invertedIndex := map[string][]int{
		"1":  {1, 3},
		"2":  {4, 2, 5},
		"3":  {6, 7},
		"4":  {8, 9, 10, 11},
		"50": {12, 13, 14, 15, 16},
	}

	// Parameters for the test
	L := 10

	// Expected results
	expectedEDB := map[string]string{
		"1":  "1100000000",
		"2":  "1111100000",
		"3":  "1111111000",
		"4":  "1111000000",
		"50": "1111111110",
	}
	expectedClusterFlist := [][]int{
		{1, 2, 3, 4, 5, 6, 7},
		{8, 9, 10, 11, 12, 13, 14, 15, 16},
	}
	expectedClusterKlist := [][]string{
		{"1", "2", "3"},
		{"4", "50"},
	}
	expectedLocalTree := map[string]string{
		"00": "3",
		"01": "50",
		"0":  "3",
	}

	// Initialize SystemParameters
	sp := Setup(L)
	// 测量关键词排序时间
	startTime := time.Now()
	sortedKeywords := sortKeywords(invertedIndex) // 排序函数
	sortDuration := time.Since(startTime).Nanoseconds()
	t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
	// Build the index
	err := sp.BuildIndex(invertedIndex, sortedKeywords)
	if err != nil {
		t.Fatalf("BuildIndex returned an error: %v", err)
	}

	// Validate EDB
	t.Logf("Validating EDB...")
	for key, expectedValue := range expectedEDB {
		hashedKey := hex.EncodeToString(sp.H1([]byte(key)))
		actualValue, ok := sp.EDB[hashedKey]
		if !ok {
			t.Errorf("EDB is missing key: %s", key)
		} else {
			t.Logf("EDB[%s:%v] = %s", key, hashedKey, actualValue)
			if string(actualValue) != expectedValue {
				t.Errorf("EDB value mismatch for key %s: expected %s, got %s", key, expectedValue, actualValue)
			}
		}
	}

	// Validate LocalTree
	t.Logf("Validating LocalTree...")
	t.Logf("Got LocalTree: %v", sp.LocalTree)
	if !reflect.DeepEqual(sp.LocalTree, expectedLocalTree) {
		t.Errorf("LocalTree mismatch: expected %v, got %v", expectedLocalTree, sp.LocalTree)
	}

	// Validate ClusterFlist
	t.Logf("Validating ClusterFlist...")
	t.Logf("Got ClusterFlist: %v", sp.ClusterFlist)
	if !reflect.DeepEqual(sp.ClusterFlist, expectedClusterFlist) {
		t.Errorf("ClusterFlist mismatch: expected %v, got %v", expectedClusterFlist, sp.ClusterFlist)
	}

	// Validate ClusterKlist
	t.Logf("Validating ClusterKlist...")
	t.Logf("Got ClusterKlist: %v", sp.ClusterKlist)
	if !reflect.DeepEqual(sp.ClusterKlist, expectedClusterKlist) {
		t.Errorf("ClusterKlist mismatch: expected %v, got %v", expectedClusterKlist, sp.ClusterKlist)
	}

	// ------------------------------
	// Search Test Section
	// ------------------------------

	// Randomly generate a query range
	//queryRange := [2]string{"2", "3"} // Example query range: ["2", "4"]
	t.Logf("Query range: %v", queryRange)

	// Expected search result
	expectedSearchResult := []int{1}

	// Step 1: Generate tokens
	tokens, err := sp.GenToken(queryRange)
	if err != nil {
		t.Fatalf("GenToken returned an error: %v", err)
	}
	t.Logf("Generated tokens: %v", tokens)

	// 如果 tokens 为空，直接返回结果为空
	if len(tokens) == 0 {
		t.Logf("Tokens are empty, returning empty result")
		return
	}

	// Step 2: Query server for search results
	searchResult := sp.SearchTokens(tokens)
	t.Logf("Search result (encrypted): %v", searchResult)

	// Step 3: Perform local search
	actualSearchResult, err := sp.LocalSearch(searchResult, tokens)
	if err != nil {
		t.Fatalf("LocalSearch returned an error: %v", err)
	}
	t.Logf("Search result (decrypted): %v", actualSearchResult)

	// Validate search result
	if !reflect.DeepEqual(actualSearchResult, expectedSearchResult) {
		t.Errorf("Search result mismatch: expected %v, got %v", expectedSearchResult, actualSearchResult)
	}
}

// TestIntegratedScheme 使用文件生成的 invertedIndex 测试整个方案
func TestIntegratedScheme(t *testing.T) {
	// 使用实际文件路径生成 invertedIndex
	filePath := config.FilePath // 文件路径，根据实际情况设置
	t.Logf("Testing with file: %s", filePath)

	// 调用 BuildInvertedIndex 函数
	invertedIndex, err := tool.BuildInvertedIndex(filePath)
	if err != nil {
		t.Fatalf("BuildInvertedIndex failed: %v", err)
	}

	// 打印倒排索引的总数量
	totalKeywords := len(invertedIndex)
	t.Logf("倒排索引关键词总数: %d", totalKeywords)

	// 随机挑选 20 个关键词打印倒排索引
	if totalKeywords <= 20 {
		for keyword, rows := range invertedIndex {
			t.Logf("Keyword: %s, Row IDs: %v", keyword, rows)
		}
	} else {
		rand.Seed(time.Now().UnixNano())
		randomKeys := getRandomKeys(invertedIndex, 20)
		for _, keyword := range randomKeys {
			t.Logf("Keyword: %s, Row IDs: %v", keyword, invertedIndex[keyword])
		}
	}

	// 设置测试参数
	L := 10 // 每个分区允许的最大大小
	t.Logf("Using L=%d for partitioning", L)

	// 初始化系统参数
	sp := Setup(L)

	// 测量关键词排序时间
	startTime := time.Now()
	sortedKeywords := sortKeywords(invertedIndex) // 排序函数
	sortDuration := time.Since(startTime).Nanoseconds()
	t.Logf("关键词排序耗时: %d 纳秒", sortDuration)
	// 构建索引
	err = sp.BuildIndex(invertedIndex, sortedKeywords)
	if err != nil {
		t.Fatalf("BuildIndex returned an error: %v", err)
	}

	// 验证 EDB（打印部分加密数据库内容）
	t.Logf("Validating EDB...")
	edbSampleCount := 5 // 仅打印前 5 项
	sampleCount := 0
	for key, value := range sp.EDB {
		t.Logf("EDB[%s]: %v", key, value)
		sampleCount++
		if sampleCount >= edbSampleCount {
			break
		}
	}
	if len(sp.EDB) == 0 {
		t.Errorf("EDB is empty after BuildIndex.")
	}

	// 验证 LocalTree
	t.Logf("Validating LocalTree...")
	t.Logf("Got LocalTree: %v", sp.LocalTree)
	if len(sp.LocalTree) == 0 {
		t.Errorf("LocalTree is empty after BuildIndex.")
	}

	// 验证 ClusterFlist 和 ClusterKlist
	t.Logf("Validating ClusterFlist and ClusterKlist...")
	t.Logf("ClusterFlist: %v", sp.ClusterFlist)
	t.Logf("ClusterKlist: %v", sp.ClusterKlist)
	if len(sp.ClusterFlist) == 0 || len(sp.ClusterKlist) == 0 {
		t.Errorf("ClusterFlist or ClusterKlist is empty after BuildIndex.")
	}

	// ------------------------------
	// Search Test Section
	// ------------------------------

	// 随机生成一个查询范围
	rand.Seed(time.Now().UnixNano())
	queryRange := generateRandomQueryRange(invertedIndex)
	t.Logf("Query range: %v", queryRange)

	// 生成期望的查询结果
	expectedSearchResult := simulateSearchResult(queryRange, invertedIndex)

	// Step 1: 生成查询 token
	tokens, err := sp.GenToken(queryRange)
	if err != nil {
		t.Fatalf("GenToken returned an error: %v", err)
	}
	t.Logf("Generated tokens: %v", tokens)

	// Step 2: 从服务器查询加密结果
	searchResult := sp.SearchTokens(tokens)
	t.Logf("Search result (encrypted): %v", searchResult)

	// Step 3: 在本地解密和处理搜索结果
	actualSearchResult, err := sp.LocalSearch(searchResult, tokens)
	if err != nil {
		t.Fatalf("LocalSearch returned an error: %v", err)
	}
	t.Logf("Search result (decrypted): %v", actualSearchResult)

	// 验证搜索结果
	if !reflect.DeepEqual(actualSearchResult, expectedSearchResult) {
		t.Errorf("Search result mismatch: expected %v, got %v", expectedSearchResult, actualSearchResult)
	}
}

// generateRandomQueryRange 生成随机的查询范围
func generateRandomQueryRange(invertedIndex map[string][]int) [2]string {
	rand.Seed(time.Now().UnixNano())
	keywords := getKeys(invertedIndex)
	if len(keywords) < 2 {
		return [2]string{keywords[0], keywords[0]}
	}
	startIndex := rand.Intn(len(keywords) - 1)
	endIndex := rand.Intn(len(keywords)-startIndex-1) + startIndex + 1
	return [2]string{keywords[startIndex], keywords[endIndex]}
}

// simulateSearchResult 模拟搜索范围对应的结果
func simulateSearchResult(queryRange [2]string, invertedIndex map[string][]int) []int {
	result := []int{}
	for keyword, ids := range invertedIndex {
		if keyword >= queryRange[0] && keyword <= queryRange[1] {
			result = append(result, ids...)
		}
	}
	return result
}

// getKeys 从 map 中获取所有 key 的切片
func getKeys(m map[string][]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// getRandomKeys 随机获取 map 中的指定数量 key
func getRandomKeys(m map[string][]int, count int) []string {
	keys := getKeys(m)
	if len(keys) <= count {
		return keys
	}

	rand.Seed(time.Now().UnixNano())
	selected := map[string]struct{}{}
	for len(selected) < count {
		randomKey := keys[rand.Intn(len(keys))]
		selected[randomKey] = struct{}{}
	}

	result := make([]string, 0, count)
	for key := range selected {
		result = append(result, key)
	}
	return result
}
