package settings

import "mvpreal/feat/config"

// IsFavorite 检查模块是否在收藏列表中
func IsFavorite(name string, cfg *config.Config) bool {
	for _, fav := range cfg.Favorites {
		if fav == name {
			return true
		}
	}
	return false
}

// IsPinned 检查模块是否在置顶列表中
func IsPinned(name string, cfg *config.Config) bool {
	for _, pinned := range cfg.PinnedModules {
		if pinned == name {
			return true
		}
	}
	return false
}
