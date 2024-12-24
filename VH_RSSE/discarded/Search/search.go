package Search

import (
	"crypto/sha256"
	"math"
	"strconv"
	"strings"
)

// SearchStruct 结构体，保存需要的状态
type SearchStruct struct {
	LocalTree     map[string]string // 本地二叉树存储
	ClusterFlist  [][]int           // 分区文件列表
	ClusterKlist  [][]string        // 分区关键词列表
	KeywordToSK   map[string][]byte // 关键词对应的 OTP 密钥
	EDB           map[string][]byte // 加密数据库
	BsLength      int               // bitmap 长度
	LocalPosition [2]int            // 查询范围的分区索引
	Flags         []string          // 标记左或右查询
	ServerTokens  []string          // 需要查询的 token
	K             []byte            // 全局密钥
}

// NewSearchStruct 构造函数
func NewSearchStruct(localTree map[string]string, clusterFlist [][]int, clusterKlist [][]string, keywordToSK map[string][]byte, edb map[string][]byte, bsLength int, k []byte) *SearchStruct {
	return &SearchStruct{
		LocalTree:    localTree,
		ClusterFlist: clusterFlist,
		ClusterKlist: clusterKlist,
		KeywordToSK:  keywordToSK,
		EDB:          edb,
		BsLength:     bsLength,
		K:            k,
	}
}

// GenToken 生成 token，用于查询服务器
func (s *SearchStruct) GenToken(queryRange [2]string) ([]string, error) {
	// 初始化
	s.Flags = []string{}
	s.ServerTokens = []string{}

	// 查询分区范围
	p1, err := s.searchTree(queryRange[0])
	if err != nil {
		return nil, err
	}
	p2, err := s.searchTree(queryRange[1])
	if err != nil {
		return nil, err
	}
	s.LocalPosition = [2]int{p1, p2}

	// 获取分区的关键词范围
	clusterKlist := s.ClusterKlist
	if queryRange[0] == clusterKlist[p1][0] && queryRange[1] == clusterKlist[p2][len(clusterKlist[p2])-1] {
		// 不需要查询服务器
		return s.ServerTokens, nil
	}

	// 生成需要查询的 token
	if queryRange[0] != clusterKlist[p1][0] {
		// 左侧边界的 token
		tempIndex := s.indexOf(clusterKlist[p1], queryRange[0])
		if tempIndex > 0 {
			tempToken := clusterKlist[p1][tempIndex-1]
			s.ServerTokens = append(s.ServerTokens, tempToken)
			s.Flags = append(s.Flags, "l")
		}
	}
	if queryRange[1] != clusterKlist[p2][len(clusterKlist[p2])-1] {
		// 右侧边界的 token
		s.ServerTokens = append(s.ServerTokens, queryRange[1])
		s.Flags = append(s.Flags, "r")
	}

	// 对 token 进行哈希
	tokens := []string{}
	for _, token := range s.ServerTokens {
		hashed := primitiveHashH([]byte(token), s.K)
		tokens = append(tokens, string(hashed))
	}
	return tokens, nil
}

// Search 查询服务器，获取结果
func (s *SearchStruct) Search(tokens []string) [][]byte {
	searchResult := [][]byte{}
	for _, token := range tokens {
		if value, ok := s.EDB[token]; ok {
			searchResult = append(searchResult, value)
		}
	}
	return searchResult
}

// LocalSearch 在本地执行解密和文件解析
func (s *SearchStruct) LocalSearch(searchResult [][]byte, tokens []string) ([]int, error) {
	lastBitmap := s.bsToBitmap("1" + strings.Repeat("0", s.BsLength-1))
	finalResult := []int{}
	p1, p2 := s.LocalPosition[0], s.LocalPosition[1]

	if len(searchResult) == 0 {
		// 无需解密，直接返回分区文件列表
		for _, fileList := range s.ClusterFlist[p1 : p2+1] {
			finalResult = append(finalResult, fileList...)
		}
		return finalResult, nil
	}

	if len(searchResult) == 2 {
		// 双 token 的情况，分别解密左右边界
		decResult := [][]byte{
			xorBytes(s.KeywordToSK[tokens[0]], searchResult[0]),
			xorBytes(s.KeywordToSK[tokens[1]], searchResult[1]),
		}
		if p1 == p2 {
			// 单分区，合并两个 bitmap
			compBitmap := xorBytes(decResult[0], decResult[1])
			finalResult = append(finalResult, s.parseFileID(compBitmap, s.ClusterFlist[p1])...)
		} else {
			// 多分区，分别解密左右 bitmap，并合并中间文件
			leftBitmap := xorBytes(decResult[0], lastBitmap)
			finalResult = append(finalResult, s.parseFileID(leftBitmap, s.ClusterFlist[p1])...)
			rightBitmap := decResult[1]
			finalResult = append(finalResult, s.parseFileID(rightBitmap, s.ClusterFlist[p2])...)
			for _, fileList := range s.ClusterFlist[p1+1 : p2] {
				finalResult = append(finalResult, fileList...)
			}
		}
	} else {
		// 单 token 的情况，处理左或右边界
		decResult := xorBytes(s.KeywordToSK[tokens[0]], searchResult[0])
		if contains(s.Flags, "l") {
			leftBitmap := xorBytes(decResult, lastBitmap)
			finalResult = append(finalResult, s.parseFileID(leftBitmap, s.ClusterFlist[p1])...)
			for _, fileList := range s.ClusterFlist[p1+1 : p2+1] {
				finalResult = append(finalResult, fileList...)
			}
		}
		if contains(s.Flags, "r") {
			finalResult = append(finalResult, s.parseFileID(decResult, s.ClusterFlist[p2])...)
			for _, fileList := range s.ClusterFlist[p1:p2] {
				finalResult = append(finalResult, fileList...)
			}
		}
	}

	return finalResult, nil
}

// 辅助方法

// __search_tree 等效的 searchTree 方法
func (s *SearchStruct) searchTree(queryValue string) (int, error) {
	node := "0"
	for i := 0; i < int(math.Log2(float64(len(s.ClusterFlist)))); i++ {
		if queryValue > s.LocalTree[node] {
			node += "1"
		} else {
			node += "0"
		}
	}
	position, err := strconv.ParseInt(node, 2, 64)
	if err != nil {
		return 0, err
	}
	return int(position), nil
}

// primitiveHashH 哈希函数实现
func primitiveHashH(msg []byte, key []byte) []byte {
	hasher := sha256.New()
	hasher.Write(key)
	hasher.Write(msg)
	return hasher.Sum(nil)
}

// xorBytes 按位异或
func xorBytes(a, b []byte) []byte {
	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// parseFileID 解析 bitmap 中的文件 ID
func (s *SearchStruct) parseFileID(bitmap []byte, dbList []int) []int {
	result := []int{}
	for i := 0; i < len(bitmap)*8; i++ {
		if i >= len(dbList) {
			break
		}
		if bitmap[i/8]&(1<<(7-i%8)) != 0 {
			result = append(result, dbList[i])
		}
	}
	return result
}

// bsToBitmap 将 bit string 转换为 bitmap
func (s *SearchStruct) bsToBitmap(bitString string) []byte {
	bitmap := make([]byte, (len(bitString)+7)/8)
	for i, bit := range bitString {
		if bit == '1' {
			bitmap[i/8] |= 1 << (7 - i%8)
		}
	}
	return bitmap
}

// contains 判断 slice 中是否包含某个值
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// indexOf 查找字符串在 slice 中的索引
func (s *SearchStruct) indexOf(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}
