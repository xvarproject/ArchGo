package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var Logger *log.Logger
var once sync.Once

func init() {
	once.Do(func() {
		Logger = log.New(os.Stdout, "INFO", log.Ldate|log.Ltime|log.Lshortfile)
	})
}

// Info 详情
func Info(format string, args ...interface{}) {
	Logger.SetPrefix("[INFO]")
	message := fmt.Sprintf(format, args...)

	Logger.Println(message)
}

// Danger 错误 为什么不命名为 error？避免和 error 类型重名
func Danger(format string, args ...interface{}) {
	Logger.SetPrefix("[ERROR]")
	message := fmt.Sprintf(format, args...)
	Logger.Fatal(message)
}

// Warning 警告
func Warning(format string, args ...interface{}) {
	Logger.SetPrefix("[WARNING]")
	message := fmt.Sprintf(format, args...)
	Logger.Println(message)
}

// DeBug debug
func DeBug(format string, args ...interface{}) {
	Logger.SetPrefix("[DeBug]")
	message := fmt.Sprintf(format, args...)
	Logger.Println(message)
}
