package binarytree

// TreeNode 表示二叉树中的一个节点
type TreeNode struct {
	Left  *TreeNode // 左子节点
	Right *TreeNode // 右子节点
	Data  [2]int    // 节点存储的数据范围
}

// BinaryTree 表示整个二叉树结构
type BinaryTree struct {
	Root *TreeNode // 二叉树的根节点
}

// BuildTree 构建二叉树，从给定的集合 C 开始
func BuildTree(C [][2]int) *BinaryTree {
	// 基础情况：如果集合为空，返回 nil（没有树）
	if len(C) == 0 {
		return &BinaryTree{Root: nil}
	}

	// 基础情况：如果集合只有一个元素
	if len(C) == 1 {
		node := &TreeNode{
			Data: C[0],
		}
		return &BinaryTree{Root: node}
	}

	// 计算树的高度 kappa
	kappa := calculateKappa(len(C))
	numLeafNodes := 1 << kappa // 2^kappa 个叶子节点
	leafNodes := make([]*TreeNode, numLeafNodes)

	// Step 1: 将集合 C 的元素分配到叶子节点
	for i := 0; i < numLeafNodes; i++ {
		if i < len(C) {
			leafNodes[i] = &TreeNode{
				Data: C[i],
			}
		} else {
			// 多余的叶子节点设置为空
			leafNodes[i] = nil
		}
	}

	// Step 2: 自底向上构建二叉树
	for level := kappa - 1; level >= 0; level-- {
		numNodes := 1 << level // 2^level
		newNodes := make([]*TreeNode, numNodes)

		for i := 0; i < numNodes; i++ {
			leftChild := leafNodes[2*i]
			rightChild := leafNodes[2*i+1]

			if leftChild == nil && rightChild == nil {
				// 如果两个子节点都为空，不再构建父节点
				newNodes[i] = nil
			} else if leftChild != nil && rightChild != nil {
				// 如果左右子节点都存在，父节点的 Data 为 {左子节点的左值, 右子节点的右值}
				newNodes[i] = &TreeNode{
					Left:  leftChild,
					Right: rightChild,
					Data:  [2]int{leftChild.Data[0], rightChild.Data[1]},
				}
			} else if leftChild != nil {
				// 只有左子节点时，父节点直接复制左子节点
				newNodes[i] = &TreeNode{
					Left: leftChild,
					Data: leftChild.Data,
				}
			}
		}
		leafNodes = newNodes
	}

	// 返回根节点作为二叉树
	return &BinaryTree{Root: leafNodes[0]}
}

// 辅助函数：计算二叉树高度
func calculateKappa(n int) int {
	kappa := 0
	for (1 << kappa) < n {
		kappa++
	}
	return kappa
}

// LocalSearch 在二叉树中执行范围查询，返回符合条件的叶子节点
// 输入范围 Q = [vLeft, vRight]，返回叶子节点数组 Cs
func LocalSearch(Q [2]int, BT *BinaryTree) [][2]int {
	if BT == nil || BT.Root == nil {
		return nil
	}

	// 查询范围
	vLeft, vRight := Q[0], Q[1]

	// 存储找到的两端叶子节点
	var Temp []*TreeNode

	// 查找包含 vLeft 的叶子节点
	leftNode := findLeafNode(vLeft, BT.Root)
	if leftNode != nil {
		Temp = append(Temp, leftNode)
	}

	// 查找包含 vRight 的叶子节点
	rightNode := findLeafNode(vRight, BT.Root)
	if rightNode != nil {
		Temp = append(Temp, rightNode)
	}

	// 返回完整的叶子节点数组 Cs
	return constructLeafArray(Temp, BT)
}

// findLeafNode 根据值查找包含该值的叶子节点，如果超出范围，将其定位到最左或最右端
func findLeafNode(value int, node *TreeNode) *TreeNode {
	if node == nil {
		return nil // 如果输入节点为空，直接返回 nil
	}

	// 找到最左叶子节点
	leftMost := findLeftMostLeaf(node)

	// 找到最右叶子节点
	rightMost := findRightMostLeaf(node)

	// 检查是否超出范围
	if value < leftMost.Data[0] {
		return leftMost // 小于最左范围，返回最左叶子节点
	}
	if value > rightMost.Data[1] {
		return rightMost // 大于最右范围，返回最右叶子节点
	}

	// 在树范围内的正常查询
	current := node
	for current != nil {
		nodeRange := current.Data

		// 如果当前节点是叶子节点，检查范围是否包含该值
		if current.Left == nil && current.Right == nil {
			if value >= nodeRange[0] && value <= nodeRange[1] {
				return current
			}
			return nil
		}

		// 根据值与范围判断向左还是向右移动
		if value <= nodeRange[1] && value >= nodeRange[0] {
			if current.Left != nil && value <= current.Left.Data[1] {
				current = current.Left
			} else if current.Right != nil {
				current = current.Right
			} else {
				break
			}
		} else {
			// 值超出范围，直接返回 nil
			return nil
		}
	}

	return nil
}

// findLeftMostLeaf 查找树中最左端的叶子节点
func findLeftMostLeaf(node *TreeNode) *TreeNode {
	current := node
	for current != nil {
		// 如果是叶子节点，直接返回
		if current.Left == nil && current.Right == nil {
			return current
		}

		// 优先向左移动，如果没有左子树，则转向右子树
		if current.Left != nil {
			current = current.Left
		} else {
			current = current.Right
		}
	}
	return nil
}

// findRightMostLeaf 查找树中最右端的叶子节点
func findRightMostLeaf(node *TreeNode) *TreeNode {
	current := node
	for current != nil {
		// 如果是叶子节点，直接返回
		if current.Left == nil && current.Right == nil {
			return current
		}

		// 优先向右移动，如果没有右子树，则转向左子树
		if current.Right != nil {
			current = current.Right
		} else {
			current = current.Left
		}
	}
	return nil
}

// constructLeafArray 构建完整的叶子节点数组 Cs
func constructLeafArray(Temp []*TreeNode, BT *BinaryTree) [][2]int {
	if len(Temp) == 0 {
		return nil
	}

	// 如果 Temp 中只有一个叶子节点，直接返回该节点
	if len(Temp) == 1 {
		return [][2]int{Temp[0].Data}
	}

	// 获取左端和右端叶子节点
	leftNode := Temp[0]
	rightNode := Temp[1]

	// 遍历所有叶子节点，收集范围内的节点
	var result [][2]int
	collectLeaves(leftNode, rightNode, BT.Root, &result)

	return result
}

// collectLeaves 遍历二叉树收集范围内的叶子节点
func collectLeaves(leftNode, rightNode, root *TreeNode, result *[][2]int) {
	// 如果根节点为 nil，直接返回
	if root == nil {
		return
	}

	// 如果是叶子节点，检查是否在范围内
	if root.Left == nil && root.Right == nil {
		if root.Data[0] >= leftNode.Data[0] && root.Data[1] <= rightNode.Data[1] {
			*result = append(*result, root.Data)
		}
		return
	}

	// 递归收集左子树和右子树的叶子节点
	collectLeaves(leftNode, rightNode, root.Left, result)
	collectLeaves(leftNode, rightNode, root.Right, result)
}
