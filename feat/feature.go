package feat

import "fyne.io/fyne/v2"

// Feature 功能模块接口
type Feature interface {
	// Name 返回功能模块名称
	Name() string
	// Create 创建功能模块的UI组件
	Create() fyne.CanvasObject
	// Help 返回帮助文档
	Help() string
}

// FeatureFactory 功能模块工厂
type FeatureFactory struct {
	features map[string]Feature
}

// NewFeatureFactory 创建一个新的功能模块工厂
func NewFeatureFactory() *FeatureFactory {
	return &FeatureFactory{
		features: make(map[string]Feature),
	}
}

// Register 注册功能模块
func (f *FeatureFactory) Register(feature Feature) {
	f.features[feature.Name()] = feature
}

// GetAll 获取所有功能模块
func (f *FeatureFactory) GetAll() []Feature {
	var result []Feature
	for _, feature := range f.features {
		result = append(result, feature)
	}
	return result
}

// Get 根据名称获取功能模块
func (f *FeatureFactory) Get(name string) Feature {
	return f.features[name]
}

// 全局功能模块工厂实例
var GlobalFeatureFactory = NewFeatureFactory()

// RegisterFeature 注册功能模块到全局工厂
func RegisterFeature(feature Feature) {
	GlobalFeatureFactory.Register(feature)
}
