package FB_RSSE

import (
	"VolumeHidingSSE/config"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math"
)

// SystemParameters 系统参数
type SystemParameters struct {
	K           []byte              // 系统密钥
	d           []int               // 查询范围
	H1          func([]byte) []byte // 哈希函数 H1
	H2          func([]byte) []byte // 哈希函数 H2
	EDB         map[string][]byte   // 加密数据库
	CT          map[string]int      // 本地树
	KeywordToSK map[string][]byte   // 每个关键词对应的密钥
	BsLength    int                 // Bitmap 长度
}

// Setup 初始化系统参数
func Setup(L int) *SystemParameters {
	// 生成随机密钥
	K := make([]byte, 16)
	rand.Read(K)
	// 定义 H1 和 H2
	H1 := func(data []byte) []byte {
		hash := sha256.Sum256(data)
		return hash[:]
	}
	H2 := func(data []byte) []byte {
		hash := sha256.Sum256(data)
		return hash[:]
	}

	// 初始化 EDB 和其他结构
	return &SystemParameters{
		K:           K,
		H1:          H1,
		H2:          H2,
		EDB:         make(map[string][]byte),
		CT:          make(map[string]int),
		KeywordToSK: make(map[string][]byte),
		BsLength:    L,
		d:           config.Range,
	}
}

// 创建FB_RSSE对象
func NewFB_RSSE(invertedIndex map[string][]int, bsLength int) (*FB_RSSE, error) {
	rsse := &FB_RSSE{
		BsLength:   bsLength,
		Db:         make(map[string][]int),
		Keyword2SK: make(map[string][]byte),
		EDB:        make(map[string][]byte),
		LocalTree:  make(map[string]string),
	}

	// 通过invertedIndex生成数据库
	for keyword, docs := range invertedIndex {
		rsse.Db[keyword] = docs
		rsse.KeywordList = append(rsse.KeywordList, keyword)
	}

	// 随机生成系统密钥和IV
	rsse.K = make([]byte, 16)
	rand.Read(rsse.K)
	rsse.IV = make([]byte, 16)
	rand.Read(rsse.IV)

	// 生成EDB
	err := rsse.GenEDB()
	if err != nil {
		return nil, err
	}

	return rsse, nil
}

// 生成EDB
func (rsse *FB_RSSE) GenEDB() error {
	rsse.LocalTree = make(map[string]string)
	rsse.EDB = make(map[string][]byte)

	// 扩展DB
	expandDb := make(map[string]map[int]bool)
	for i := range rsse.KeywordList {
		expandDb[rsse.KeywordList[i]] = make(map[int]bool)
	}

	// 树高
	treeHeight := int(math.Ceil(math.Log2(float64(len(rsse.KeywordList)))))

	// 生成本地树
	for i := treeHeight; i >= 0; i-- {
		for j := 0; j < int(math.Pow(2, float64(i))); j++ {
			tempKeyword := fmt.Sprintf("%0*b", i+1, j)
			if i == treeHeight {
				rsse.LocalTree[tempKeyword] = rsse.KeywordList[j]
			} else {
				leftNode := rsse.LocalTree[tempKeyword+"0"]
				rightNode := rsse.LocalTree[tempKeyword+"1"]
				rsse.LocalTree[tempKeyword] = fmt.Sprintf("%s,%s", leftNode, rightNode)
			}
		}
	}

	// 开始生成位图并加密
	for _, keyword := range rsse.KeywordList {
		hashKeyword := H1([]byte(keyword))
		bitmap := rsse.genBitmap(rsse.Db[keyword])
		otpKey := make([]byte, rsse.BsLength/8)
		rand.Read(otpKey)
		encBitmap := bxor(bitmap, otpKey)
		rsse.EDB[string(hashKeyword)] = encBitmap
		rsse.Keyword2SK[string(hashKeyword)] = otpKey
	}

	return nil
}

// 根据文档生成位图
func (rsse *FB_RSSE) genBitmap(fileList []int) []byte {
	bitmap := make([]byte, rsse.BsLength/8)
	for _, fileID := range fileList {
		// 在指定的位设置为1
		bitmap[fileID/8] |= (1 << uint(fileID%8))
	}
	return bitmap
}

// 生成查询令牌
func (rsse *FB_RSSE) GenToken(queryRange [2]string) []byte {
	// 修改：bin转int，并根据树结构计算令牌
	leftNode := rsse.searchTree(queryRange[0])
	rightNode := rsse.searchTree(queryRange[1])

	var nodes []string
	for int(parseBinToInt(leftNode)) < int(parseBinToInt(rightNode)) {
		if leftNode[len(leftNode)-1] == '1' {
			nodes = append(nodes, leftNode)
		}
		if rightNode[len(rightNode)-1] == '0' {
			nodes = append(nodes, rightNode)
		}
		leftNode = fmt.Sprintf("%b", parseBinToInt(leftNode)+1)
		rightNode = fmt.Sprintf("%b", parseBinToInt(rightNode)-1)
	}
	if leftNode == rightNode {
		nodes = append(nodes, leftNode)
	}
	tokensList := []byte{}
	for _, keyword := range nodes {
		hashKeyword := H1([]byte(keyword))
		tokensList = append(tokensList, hashKeyword...)
	}
	return tokensList
}

// 转换二进制字符串为整数
func parseBinToInt(binStr string) int {
	i, _ := fmt.Sscanf(binStr, "%b")
	return i
}

// 获取树位置
func (rsse *FB_RSSE) searchTree(queryValue string) string {
	node := "0"
	for i := 0; i < len(rsse.KeywordList); i++ {
		if queryValue < rsse.LocalTree[node] {
			node += "0"
		} else {
			node += "1"
		}
	}
	return node
}

// 搜索：根据查询令牌获取加密位图
func (rsse *FB_RSSE) Search(tokensList []byte) [][]byte {
	searchResult := [][]byte{}
	for i := 0; i < len(tokensList); i++ {
		token := tokensList[i]
		encBitmap, exists := rsse.EDB[string(token)]
		if exists {
			searchResult = append(searchResult, encBitmap)
		}
	}
	return searchResult
}

// gen_ids: 解密并生成ID列表
func (rsse *FB_RSSE) GenIds(searchResult [][]byte, tokensList []byte) []int {
	finalResult := []int{}
	for i, encBitmap := range searchResult {
		otpKey := rsse.Keyword2SK[string(tokensList[i])]
		bitmap := bxor(encBitmap, otpKey)
		fileIds := rsse.parseFileIds(bitmap)
		finalResult = append(finalResult, fileIds...)
	}
	return finalResult
}

// 根据位图生成文件ID
func (rsse *FB_RSSE) parseFileIds(bitmap []byte) []int {
	fileIds := []int{}
	for i := 0; i < len(bitmap); i++ {
		if bitmap[i] == 1 {
			fileIds = append(fileIds, i)
		}
	}
	return fileIds
}

// H1 哈希函数：计算哈希值
var H1 = func(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// 位图XOR
func bxor(b1, b2 []byte) []byte {
	if len(b1) != len(b2) {
		log.Fatalf("XOR: byte slices must have the same length")
	}
	result := make([]byte, len(b1))
	for i := 0; i < len(b1); i++ {
		result[i] = b1[i] ^ b2[i]
	}
	return result
}
