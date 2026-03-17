package gui

import (
	"testing"
)

func TestRegisterFeatures(t *testing.T) {
	features := registerFeatures()
	if features == nil {
		t.Error("registerFeatures() 返回 nil")
	}
	if len(features) == 0 {
		t.Error("registerFeatures() 返回空列表")
	}
}
