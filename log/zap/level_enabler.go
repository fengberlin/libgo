package zaplog

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	_minLevel = zapcore.DebugLevel
	_maxLevel = zapcore.FatalLevel
)

type levelAndAbove zapcore.Level

func (l levelAndAbove) EnableLevels() map[zapcore.Level]zap.LevelEnablerFunc {
	levels := GetLevels(zapcore.Level(l))
	m := make(map[zapcore.Level]zap.LevelEnablerFunc, len(levels))
	for i := 0; i < len(levels); i++ {
		// 注意这里的索引下标
		idx := i
		m[levels[idx]] = func(lvl zapcore.Level) bool {
			return lvl == levels[idx]
		}
	}
	return m
}

// GetLevelNames get the enabled level names
func GetLevels(lvl zapcore.Level) (enabledLevels []zapcore.Level) {
	if lvl < _minLevel || lvl > _maxLevel {
		panic(errors.New("invalid log level"))
	}
	enabledLevels = make([]zapcore.Level, _maxLevel-lvl+1)
	for i := 0; i < int(_maxLevel-lvl+1); i++ {
		enabledLevels[i] = lvl + zapcore.Level(i)
	}
	return
}
