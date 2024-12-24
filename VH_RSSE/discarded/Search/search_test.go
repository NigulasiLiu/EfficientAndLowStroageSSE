package Search

import (
	"VH_RSSE/discarded/binarytree"
	"testing"
)

func TestExecuteSearch(t *testing.T) {
	// 示例集合
	C := []int{1, 3, 5, 7, 9, 11, 13, 15}
	BT := binarytree.BuildTree(C)

	// 测试查询范围
	Q := [2]int{5, 13}
	results := ExecuteSearch(Q, BT)

	// 检查返回结果
	if len(results) == 0 {
		t.Errorf("测试失败：未返回结果")
	}
}
