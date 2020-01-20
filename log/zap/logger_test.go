package zaplog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
	"time"
)

func init() {
	os.Setenv("KUBE_PODNAME", "kubepod")
	InitLogger(WithServiceName("data"), WithLogPath("../"))
}

func TestLogger(t *testing.T) {
	tick := time.Tick(time.Second * 2)
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
		Infof("%s", "kkkkkkk")
		Errorf("%s", "lllll")
		time.Sleep(time.Second * 1)
		DPanicf("%s", "pppp")
		Warnw("ww", zap.String("cc", "vv"), zap.Int64("int", 123))
		Infow("xx", zap.String("dd", "ff"), zap.Int64("int", 1111))
		break
	}
	err := Sync()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func TestLL(t *testing.T) {
	l, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
		return
	}
	l.Sugar().Infow("aa", "l", "n")
}
