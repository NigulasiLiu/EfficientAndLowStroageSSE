package discarded

import (
	"VolumeHidingSSE/config"
	"fmt"
	"sort"
	"strconv"
)

// 全局参数 L
//const L = 6264

// GenerateCollection 从 Order-Weighted Inverted Index 中生成集合
func GenerateCollection(owIndex *OWInvertedIndex) ([][2]int, error) {
	if len(owIndex.Index) == 0 {
		return nil, fmt.Errorf("owIndex is empty")
	}

	// 按关键字排序
	keys := make([]string, 0, len(owIndex.Index))
	for key := range owIndex.Index {
		keys = append(keys, key)
	}
	sortKeysAsFloat(keys)

	// 初始化集合
	var collection [][2]int
	counter := 0
	var currentStart float64

	// 遍历键值对，流式生成集合
	for i, key := range keys {
		currentIDs := owIndex.Index[key]
		currentCount := len(currentIDs)

		// 将 key 转换为浮点数
		keyFloat, err := strconv.ParseFloat(key, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid key format: %v", err)
		}

		if i == 0 {
			currentStart = keyFloat
		}

		if counter+currentCount < config.L {
			counter += currentCount
		} else {
			// 生成集合元素
			currentEnd := keyFloat
			collection = append(collection, [2]int{
				int(currentStart * 10000), // 转换为整数并存储
				int(currentEnd * 10000),
			})

			// 开始新分区
			currentStart = keyFloat
			counter = currentCount
		}
	}

	// 处理最后一组
	if counter > 0 {
		lastKeyFloat, err := strconv.ParseFloat(keys[len(keys)-1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid key format: %v", err)
		}
		collection = append(collection, [2]int{
			int(currentStart * 10000),
			int(lastKeyFloat * 10000),
		})
	}

	return collection, nil
}

// sortKeysAsFloat 将字符串数组按浮点数值从小到大排序
func sortKeysAsFloat(keys []string) {
	sort.Slice(keys, func(i, j int) bool {
		keyI, _ := strconv.ParseFloat(keys[i], 64)
		keyJ, _ := strconv.ParseFloat(keys[j], 64)
		return keyI < keyJ
	})
}
