package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtoHelper(t *testing.T) {
	t.Run("ToEnumProto", func(t *testing.T) {
		t.Run("should convert from model to proto type", func(t *testing.T) {
			modelType := "task"
			expectedProtoType := "TYPE_TASK"
			enumName := "TYPE"
			actualType := ToEnumProto(modelType, enumName)
			assert.Equal(t, expectedProtoType, actualType)
		})
	})
	t.Run("FromEnumProto", func(t *testing.T) {
		t.Run("should convert from proto to model type", func(t *testing.T) {
			expectedModelType := "task"
			protoType := "TYPE_TASK"
			enumName := "type"
			actualType := FromEnumProto(protoType, enumName)
			assert.Equal(t, expectedModelType, actualType)
		})
	})
}
