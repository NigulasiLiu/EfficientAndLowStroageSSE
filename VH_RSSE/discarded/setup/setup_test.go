package setup

import (
	"VH_RSSE/discarded/binarytree"
	"VolumeHidingSSE/Experiment"
	tool2 "VolumeHidingSSE/Experiment/tool"
	"testing"
)

func TestSetupWithCollection(t *testing.T) {
	// 文件路径
	filePath := Experiment.FilePath

	// 调用 BuildInvertedIndex 函数生成基础倒排索引
	invertedIndex, err := tool2.BuildInvertedIndex(filePath)
	if err != nil {
		t.Fatalf("BuildInvertedIndex failed: %v", err)
	}

	// 调用 BuildOWInvertedIndex 函数生成 Order-Weighted Inverted Index
	owIndex, err := tool2.BuildOWInvertedIndex(invertedIndex)
	if err != nil {
		t.Fatalf("BuildOWInvertedIndex failed: %v", err)
	}

	// 调用 GenerateCollection 函数生成 Collection
	collection, err := tool2.GenerateCollection(owIndex)
	if err != nil {
		t.Fatalf("GenerateCollection failed: %v", err)
	}

	// 调用 Setup 函数
	systemParams, err := Setup(collection)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 打印初始化的参数
	t.Logf("Initialized Parameters:")
	t.Logf("L (max volume of database): %d", systemParams.L)
	t.Logf("K (random key): %s", systemParams.K)
	t.Logf("H1 (hash function example): %s", systemParams.H1([]byte("example")))
	t.Logf("H2 (hash function example): %s", systemParams.H2([]byte("example")))
	t.Logf("F (PRF example): %s", systemParams.F([]byte("example")))
	t.Logf("EDB (encrypted database): %+v", systemParams.EDB)

	// 打印生成的二叉树（前 4 层节点）
	t.Logf("Binary Tree (Top 4 Levels):")
	printTreeLevel(systemParams.BT.Root, 0, 4, t)
}

// printTreeLevel 打印二叉树指定层的节点，将同一层的节点放在同一行
func printTreeLevel(node *binarytree.TreeNode, currentLevel, maxLevel int, t *testing.T) {
	if node == nil || currentLevel >= maxLevel {
		return
	}

	// 使用队列实现按层遍历
	type LevelNode struct {
		Node  *binarytree.TreeNode
		Level int
	}

	queue := []LevelNode{{Node: node, Level: 0}} // 初始化队列，根节点入队
	levelMap := make(map[int][][2]int)           // 存储每一层的节点数据

	for len(queue) > 0 {
		// 弹出队列头
		current := queue[0]
		queue = queue[1:]

		// 将节点数据存储到对应的层
		if current.Node != nil {
			levelMap[current.Level] = append(levelMap[current.Level], current.Node.Data)

			// 将左右子节点加入队列
			if current.Level+1 < maxLevel {
				queue = append(queue, LevelNode{Node: current.Node.Left, Level: current.Level + 1})
				queue = append(queue, LevelNode{Node: current.Node.Right, Level: current.Level + 1})
			}
		}
	}

	// 打印每一层的节点数据
	for level := 0; level < maxLevel; level++ {
		if nodes, exists := levelMap[level]; exists {
			t.Logf("Level %d: %v", level, nodes)
		}
	}
}
