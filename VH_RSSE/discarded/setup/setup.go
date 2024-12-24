package setup

import (
	"VH_RSSE/discarded/binarytree"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

// SystemParameters 表示系统参数
type SystemParameters struct {
	L   int                    // L 的值
	K   string                 // 随机生成的密钥 K
	H1  func([]byte) string    // 哈希函数 H1
	H2  func([]byte) string    // 哈希函数 H2
	F   func([]byte) string    // 伪随机函数 F
	EDB map[string]string      // 加密的数据库
	BT  *binarytree.BinaryTree // 二叉树
}

// HMACPRF 是伪随机置换函数的实现（使用 HMAC-SHA256）
func HMACPRF(key []byte) func([]byte) string {
	return func(data []byte) string {
		h := hmac.New(sha256.New, key)
		h.Write(data)
		return hex.EncodeToString(h.Sum(nil))
	}
}

// Blake2bHash 是一个通用的哈希函数生成器，返回一个 128 位哈希
func Blake2bHash() func([]byte) string {
	return func(data []byte) string {
		hash, err := blake2b.New(16, nil) // 创建一个 128 位的 Blake2b 哈希实例
		if err != nil {
			panic("failed to initialize Blake2b hash function")
		}
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	}
}

// Setup 函数初始化系统参数
func Setup(C [][2]int) (*SystemParameters, error) {
	// 初始化 L 和 K
	L := 6264 // L 取固定值 6264

	// 生成长度为 lambda 的随机密钥 (128 bits)
	key := make([]byte, 16) // 16 字节 = 128 bits
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %v", err)
	}
	K := hex.EncodeToString(key)

	// 初始化哈希函数 H1 和 H2
	H1 := Blake2bHash() // 使用通用的 Blake2b 哈希函数
	H2 := Blake2bHash()

	// 初始化伪随机函数 F
	F := HMACPRF(key)

	// 初始化空的 EDB 和二叉树 BT
	EDB := make(map[string]string)
	BT := binarytree.BuildTree(C) // 使用输入的集合 C 构建二叉树

	// 返回系统参数
	return &SystemParameters{
		L:   L,
		K:   K,
		H1:  H1,
		H2:  H2,
		F:   F,
		EDB: EDB,
		BT:  BT,
	}, nil
}

// PrintSystemParameters 打印系统参数
func (sp *SystemParameters) PrintSystemParameters() {
	fmt.Printf("System Parameters:\n")
	fmt.Printf("L: %d\n", sp.L)
	fmt.Printf("K: %s\n", sp.K)
	fmt.Printf("Binary Tree: %+v\n", sp.BT)
	fmt.Printf("EDB: %+v\n", sp.EDB)
}
