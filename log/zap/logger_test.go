package zaplog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
	"time"
)

func init() {
	InitLogger()
}

func TestLogger(t *testing.T) {
	tick := time.Tick(time.Second*2)
	for {
		<-tick
		Debug("hello debug", zap.String("k", "v"))
		Info("hello info")
		Warn("hello warn")
		Error("hello error")
		DPanic("hello dpanic")
		atomicLevel.SetLevel(zapcore.ErrorLevel)
		Info("hello info2")
		Warn("hello warn2")
		Error("hello error2")
	}
}