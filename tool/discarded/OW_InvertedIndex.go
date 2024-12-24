package discarded

import (
	"fmt"
	"sort"
	"strconv"
)

// OWInvertedIndex 表示 Order-Weighted Inverted Index
type OWInvertedIndex struct {
	Index map[string][]int // 关键字 -> 文档标识符的累积集合
}

// BuildOWInvertedIndex 构建 Order-Weighted Inverted Index
func BuildOWInvertedIndex(invertedIndex map[string][]int) (*OWInvertedIndex, error) {
	owIndex := make(map[string][]int)

	// 将关键字转换为 float64 并排序
	keys := make([]float64, 0, len(invertedIndex))
	keyMap := make(map[float64]string) // 数值键 -> 原始字符串键
	for k := range invertedIndex {
		// 转换字符串关键字为浮点数
		keyFloat, err := strconv.ParseFloat(k, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid key format: %s", k)
		}
		keys = append(keys, keyFloat)
		keyMap[keyFloat] = k
	}

	sort.Float64s(keys) // 数值升序排序

	// 流式生成累积倒排索引
	var cumulativeIdentifiers []int
	for _, keyFloat := range keys {
		key := keyMap[keyFloat]

		// 按需更新累积集合（避免一次性加载所有数据）
		cumulativeIdentifiers = append(cumulativeIdentifiers, invertedIndex[key]...)
		owIndex[key] = cumulativeIdentifiers
	}

	return &OWInvertedIndex{Index: owIndex}, nil
}

// RangeQuery 执行范围查询 [vLow, vHigh]
func (ow *OWInvertedIndex) RangeQuery(vLow, vHigh string) ([]int, error) {
	if len(ow.Index) == 0 {
		return nil, fmt.Errorf("the index is empty")
	}

	lowIdentifiers, lowExists := ow.Index[vLow] // vLow 索引
	highIdentifiers, highExists := ow.Index[vHigh]

	if !highExists {
		return nil, fmt.Errorf("vHigh %s not found in index", vHigh)
	}

	if !lowExists {
		// 如果 low 不存在，则直接返回 highIdentifiers
		return highIdentifiers, nil
	}

	// 计算 I(vHigh) \ I(vLow)
	result := differenceInt(highIdentifiers, lowIdentifiers)
	return result, nil
}

// differenceInt 计算两个集合的差集 A \ B
func differenceInt(a, b []int) []int {
	setB := make(map[int]struct{})
	for _, val := range b {
		setB[val] = struct{}{}
	}

	var result []int
	for _, val := range a {
		if _, exists := setB[val]; !exists {
			result = append(result, val)
		}
	}
	return result
}
