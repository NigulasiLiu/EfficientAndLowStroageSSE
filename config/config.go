package config

// 全局配置参数
var (
	FilePath       string  // 数据文件路径
	FilePath_raw   string  // 数据文件路径
	FilePath_index string  // 数据文件路径
	FilePath_txt   string  // 数据文件路径
	L              int     // 分区大小限制
	Lambda         int     // 安全参数 lambda
	Divide         float64 // 安全参数 lambda
	Range          []int
)

// init 函数初始化全局变量
func init() {
	// 默认文件路径
	FilePath = "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\origin.csv"
	FilePath_raw = "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\Gowalla_totalCheckins.txt"
	FilePath_index = "C:\\Users\\Admin\\Desktop\\GoPros\\EfficientAndLowStroageSSE\\dataset\\InvertedIndex.csv"
	FilePath_txt = "dataset/Gowalla_invertedIndex_new.txt"
	// 默认分区大小限制
	L = 6264

	// 默认安全参数 lambda
	Lambda = 128

	Divide = 10000
	Range = []int{0, 335348}
}
