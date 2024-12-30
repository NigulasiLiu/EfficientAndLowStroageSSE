package binarytree

import (
	"EfficientAndLowStroageSSE/Experiment"
	tool2 "EfficientAndLowStroageSSE/Experiment/tool"
	"testing"
)

// TestBuildTreeWithDataset 测试从实际数据生成二叉树并展示前四层节点
func TestBuildTreeWithDataset(t *testing.T) {
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
	//C := [][2]int{
	//	{1, 3},
	//	{4, 6},
	//	{7, 10},
	//	{11, 15},
	//	{16, 20},
	//}
	// 调用 BuildTree 函数生成二叉树
	BT := BuildTree(collection)

	// 打印二叉树的前四层节点
	t.Logf("Top 4 levels of the Binary Tree:\n")
	printTreeLevel(BT.Root, 0, 4, t)
}

// printTreeLevel 打印二叉树指定层的节点，将同一层的节点放在同一行
func printTreeLevel(node *TreeNode, currentLevel, maxLevel int, t *testing.T) {
	if node == nil || currentLevel >= maxLevel {
		return
	}

	// 使用队列实现按层遍历
	type LevelNode struct {
		Node  *TreeNode
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

func TestBuildTree(t *testing.T) {
	// 测试用例 1：空集合
	C1 := [][2]int{}
	BT1 := BuildTree(C1)
	if BT1.Root != nil {
		t.Errorf("Test failed: expected Root to be nil, got %+v", BT1.Root)
	}

	// 测试用例 2：单个元素
	C2 := [][2]int{{5, 10}}
	BT2 := BuildTree(C2)
	if BT2.Root == nil || BT2.Root.Data != [2]int{5, 10} {
		t.Errorf("Test failed: expected Root Data to be {5, 10}, got %+v", BT2.Root)
	}

	// 测试用例 3：多个元素
	C3 := [][2]int{
		{1, 3},
		{4, 6},
		{7, 10},
		{11, 15},
		{16, 20},
	}
	BT3 := BuildTree(C3)
	if BT3.Root == nil {
		t.Errorf("Test failed: expected Root to be non-nil, got nil")
	} else {
		t.Logf("Binary Tree Root Data: %+v", BT3.Root.Data)
	}

	// 测试树的结构是否正确
	// 检查根节点
	expectedRootData := [2]int{1, 20}
	if BT3.Root.Data != expectedRootData {
		t.Errorf("Test failed: expected Root Data to be %+v, got %+v", expectedRootData, BT3.Root.Data)
	}

	// 检查左子节点
	expectedLeftChildData := [2]int{1, 15}
	if BT3.Root.Left == nil || BT3.Root.Left.Data != expectedLeftChildData {
		t.Errorf("Test failed: expected Left Child Data to be %+v, got %+v", expectedLeftChildData, BT3.Root.Left)
	}

	// 检查右子节点
	expectedRightChildData := [2]int{16, 20}
	if BT3.Root.Right == nil || BT3.Root.Right.Data != expectedRightChildData {
		t.Errorf("Test failed: expected Right Child Data to be %+v, got %+v", expectedRightChildData, BT3.Root.Right)
	}

	t.Logf("TestBuildTree passed")
}

func TestLocalSearchWithOutOfRange(t *testing.T) {
	// 构建测试用例的二叉树
	C := [][2]int{
		{1, 3},
		{4, 6},
		{7, 10},
		{11, 15},
		{16, 20},
	}
	BT := BuildTree(C)

	// 测试用例 1：左端超出范围
	Q1 := [2]int{-10, 5}
	expected1 := [][2]int{{1, 3}, {4, 6}}
	result1 := LocalSearch(Q1, BT)
	if !compareSlices(result1, expected1) {
		t.Errorf("Test failed for Q1: expected %+v, got %+v", expected1, result1)
	}

	// 测试用例 2：右端超出范围
	Q2 := [2]int{18, 30}
	expected2 := [][2]int{{16, 20}}
	result2 := LocalSearch(Q2, BT)
	if !compareSlices(result2, expected2) {
		t.Errorf("Test failed for Q2: expected %+v, got %+v", expected2, result2)
	}

	// 测试用例 3：两端均超出范围
	Q3 := [2]int{-5, 25}
	expected3 := [][2]int{{1, 3}, {4, 6}, {7, 10}, {11, 15}, {16, 20}}
	result3 := LocalSearch(Q3, BT)
	if !compareSlices(result3, expected3) {
		t.Errorf("Test failed for Q3: expected %+v, got %+v", expected3, result3)
	}

	t.Logf("TestLocalSearchWithOutOfRange passed")
}

// compareSlices 比较两个 [][2]int 切片是否相等
func compareSlices(a, b [][2]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
