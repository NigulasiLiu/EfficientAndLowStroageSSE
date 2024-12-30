package BuildIndex

import (
	"EfficientAndLowStroageSSE/Experiment/tool"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

// BuildIndex 构建加密数据库和二叉树
func BuildIndex(filePath string, L int) (map[string]string, map[string]string, []map[string]interface{}, error) {
	// Step 1: 构建倒排索引
	invertedIndex, err := tool.BuildInvertedIndex(filePath)
	if err != nil {
		return nil, nil, nil, errors.New("failed to build inverted index")
	}

	// Step 2: 初始化参数
	EDB := make(map[string]string)          // 加密数据库
	clusterFlist := make([][]int, 0)        // 每个分区的文件ID集合
	clusterKlist := make([][]string, 0)     // 每个分区的关键词集合
	clusterVolume := make([]int, 0)         // 每个分区的偏移量和文件数
	currentGroup := make([]int, 0)          // 当前分区的文件ID
	currentKlist := make([]string, 0)       // 当前分区的关键词
	keywordToKey := make(map[string][]byte) // 关键词对应的 OTP 密钥

	// Step 3: 获取所有关键词并排序
	keywords := make([]string, 0, len(invertedIndex))
	for keyword := range invertedIndex {
		keywords = append(keywords, keyword)
	}

	// 按关键词的数值顺序排序
	sort.Slice(keywords, func(i, j int) bool {
		ki, _ := strconv.ParseFloat(keywords[i], 64)
		kj, _ := strconv.ParseFloat(keywords[j], 64)
		return ki < kj
	})

	// Step 4: 遍历关键词
	for i, keyword := range keywords {
		postings := invertedIndex[keyword] // 当前关键词对应的文档ID列表

		// 检查是否可以继续添加到当前分区
		if len(currentGroup)+len(postings) <= L {
			currentGroup = append(currentGroup, postings...)
			currentKlist = append(currentKlist, keyword)

			// 生成加密的 Bitmap 并存储到 EDB
			bitmap := generateBitmap(currentGroup, L)
			otpKey := make([]byte, len(bitmap))
			encBitmap := xorBitmap(bitmap, otpKey)
			EDB[keyword] = encBitmap
			keywordToKey[keyword] = otpKey

			// 如果是最后一个关键词
			if i == len(keywords)-1 {
				clusterFlist = append(clusterFlist, append([]int{}, currentGroup...))
				clusterKlist = append(clusterKlist, append([]string{}, currentKlist...))
				clusterVolume = append(clusterVolume, len(currentGroup))
			}
		} else {
			// 超出分区大小时，将当前分区写入
			clusterFlist = append(clusterFlist, append([]int{}, currentGroup...))
			clusterKlist = append(clusterKlist, append([]string{}, currentKlist...))
			clusterVolume = append(clusterVolume, len(currentGroup))

			// 重置分区
			currentGroup = append([]int{}, postings...)
			currentKlist = append([]string{}, keyword)

			// 生成加密的 Bitmap 并存储到 EDB
			bitmap := generateBitmap(currentGroup, L)
			otpKey := make([]byte, len(bitmap))
			encBitmap := xorBitmap(bitmap, otpKey)
			EDB[keyword] = encBitmap
			keywordToKey[keyword] = otpKey
		}
	}

	// Step 5: 构建基于 map 的 localTree
	clusterHeight := int(math.Ceil(math.Log2(float64(len(clusterFlist)))))
	localTree := make(map[string]string) // 使用 map 代替二叉树存储 localTree
	genList := make([]string, len(clusterKlist))
	for i, klist := range clusterKlist {
		genList[i] = klist[len(klist)-1]
	}

	// 填充 Padding
	padding := genList[len(genList)-1]
	for len(genList) < int(math.Pow(2, float64(clusterHeight))) {
		genList = append(genList, padding)
	}

	// 构建 map 形式的 localTree
	for i := clusterHeight; i >= 0; i-- {
		for j := 0; j < int(math.Pow(2, float64(i))); j++ {
			tempKeyword := strconv.FormatInt(int64(j), 2)
			tempKeyword = fmt.Sprintf("%0*s", i+1, tempKeyword) // 补齐位数
			if i == clusterHeight {
				localTree[tempKeyword] = genList[j]
			} else {
				tempVal := localTree[tempKeyword+"0"]
				localTree[tempKeyword] = tempVal
			}
		}
	}

	// Step 6: 返回结果
	return EDB, localTree, []map[string]interface{}{
		{
			"flist":  clusterFlist,
			"volume": clusterVolume,
			"klist":  clusterKlist,
		},
	}, nil
}
func BuildIndexFromInvertedIndex(invertedIndex map[string][]int, L int) (map[string]string, map[string]string, []map[string]interface{}, error) {
	// Step 1: 初始化参数
	EDB := make(map[string]string)          // 加密数据库
	clusterFlist := make([][]int, 0)        // 每个分区的文件ID集合
	clusterKlist := make([][]string, 0)     // 每个分区的关键词集合
	clusterVolume := make([]int, 0)         // 每个分区的偏移量和文件数
	currentGroup := make([]int, 0)          // 当前分区的文件ID
	currentKlist := make([]string, 0)       // 当前分区的关键词
	keywordToKey := make(map[string][]byte) // 关键词对应的 OTP 密钥

	// Step 2: 获取所有关键词并排序
	keywords := make([]string, 0, len(invertedIndex))
	for keyword := range invertedIndex {
		keywords = append(keywords, keyword)
	}

	// 按关键词的数值顺序排序
	sort.Slice(keywords, func(i, j int) bool {
		ki, _ := strconv.ParseFloat(keywords[i], 64)
		kj, _ := strconv.ParseFloat(keywords[j], 64)
		return ki < kj
	})

	// Step 3: 遍历关键词
	for i, keyword := range keywords {
		postings := invertedIndex[keyword] // 当前关键词对应的文档ID列表

		// 检查是否可以继续添加到当前分区
		if len(currentGroup)+len(postings) <= L {
			currentGroup = append(currentGroup, postings...)
			currentKlist = append(currentKlist, keyword)

			// 生成加密的 Bitmap 并存储到 EDB
			bitmap := generateBitmap(currentGroup, L)

			otpKey := make([]byte, len(bitmap))
			encBitmap := xorBitmap(bitmap, otpKey)
			EDB[keyword] = encBitmap
			keywordToKey[keyword] = otpKey

			// 如果是最后一个关键词
			if i == len(keywords)-1 {
				clusterFlist = append(clusterFlist, append([]int{}, currentGroup...))
				clusterKlist = append(clusterKlist, append([]string{}, currentKlist...))
				clusterVolume = append(clusterVolume, len(currentGroup))
			}
		} else {
			// 超出分区大小时，将当前分区写入
			clusterFlist = append(clusterFlist, append([]int{}, currentGroup...))
			clusterKlist = append(clusterKlist, append([]string{}, currentKlist...))
			clusterVolume = append(clusterVolume, len(currentGroup))

			// 重置分区
			currentGroup = append([]int{}, postings...)
			currentKlist = append([]string{}, keyword)

			// 生成加密的 Bitmap 并存储到 EDB
			bitmap := generateBitmap(currentGroup, L)
			otpKey := make([]byte, len(bitmap))
			encBitmap := xorBitmap(bitmap, otpKey)
			EDB[keyword] = encBitmap
			keywordToKey[keyword] = otpKey
		}
	}

	// Step 4: 构建基于 map 的 localTree
	clusterHeight := int(math.Ceil(math.Log2(float64(len(clusterFlist)))))
	localTree := make(map[string]string) // 使用 map 代替二叉树存储 localTree
	genList := make([]string, len(clusterKlist))
	for i, klist := range clusterKlist {
		genList[i] = klist[len(klist)-1]
	}

	// 填充 Padding
	padding := genList[len(genList)-1]
	for len(genList) < int(math.Pow(2, float64(clusterHeight))) {
		genList = append(genList, padding)
	}

	// 构建 map 形式的 localTree
	for i := clusterHeight; i >= 0; i-- {
		for j := 0; j < int(math.Pow(2, float64(i))); j++ {
			tempKeyword := strconv.FormatInt(int64(j), 2)
			tempKeyword = fmt.Sprintf("%0*s", i+1, tempKeyword) // 补齐位数
			if i == clusterHeight {
				localTree[tempKeyword] = genList[j]
			} else {
				tempVal := localTree[tempKeyword+"0"]
				localTree[tempKeyword] = tempVal
			}
		}
	}

	// Step 5: 返回结果
	return EDB, localTree, []map[string]interface{}{
		{
			"flist":  clusterFlist,
			"volume": clusterVolume,
			"klist":  clusterKlist,
		},
	}, nil
}

// 辅助函数：生成位图
func generateBitmap(group []int, L int) []byte {
	bitString := strings.Repeat("1", len(group)) + strings.Repeat("0", L-len(group))
	bitmap, _ := strconv.ParseUint(bitString, 2, 64)
	return []byte(fmt.Sprintf("%b", bitmap))
}

func xorBitmap(bitmap []byte, otpKey []byte) string {
	result := make([]byte, len(bitmap))
	for i := range bitmap {
		result[i] = bitmap[i] ^ otpKey[i]
	}
	return string(result)
}

// 辅助函数：计算范围
func rangeBetween(min string, max string) []string {
	// 假设关键词为字符串，可以实现具体范围计算逻辑
	return []string{min, max} // 示例返回
}

// 辅助函数：最大值
func max(vlist []string) string {
	if len(vlist) == 0 {
		return ""
	}

	maxValue := vlist[0]
	maxFloat, err := strconv.ParseFloat(maxValue, 64)
	if err != nil {
		log.Fatalf("Failed to parse %s to float64: %v", maxValue, err)
	}

	for _, v := range vlist {
		floatVal, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Fatalf("Failed to parse %s to float64: %v", v, err)
		}

		if floatVal > maxFloat {
			maxValue = v
			maxFloat = floatVal
		}
	}

	return maxValue
}

// 辅助函数：最小值
func min(vlist []string) string {
	if len(vlist) == 0 {
		return ""
	}

	minValue := vlist[0]
	minFloat, err := strconv.ParseFloat(minValue, 64)
	if err != nil {
		log.Fatalf("Failed to parse %s to float64: %v", minValue, err)
	}

	for _, v := range vlist {
		floatVal, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Fatalf("Failed to parse %s to float64: %v", v, err)
		}

		if floatVal < minFloat {
			minValue = v
			minFloat = floatVal
		}
	}

	return minValue
}

//// 辅助函数：最大值
//func max(vlist []string) string {
//	if len(vlist) == 0 {
//		return ""
//	}
//	maxValue := vlist[0]
//	for _, v := range vlist {
//		if v > maxValue {
//			maxValue = v
//		}
//	}
//	return maxValue
//}
//
//// 辅助函数：最小值
//func min(vlist []string) string {
//	if len(vlist) == 0 {
//		return ""
//	}
//	minValue := vlist[0]
//	for _, v := range vlist {
//		if v < minValue {
//			minValue = v
//		}
//	}
//	return minValue
//}
