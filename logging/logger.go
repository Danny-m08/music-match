package logging

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
)

func Info(msg string) { log.Printf("%s\t%s\t%s", color.GreenString("[INFO]"), getSource(), msg) }

func Error(msg string) { log.Printf("%s\t%s\t%s", color.RedString("[ERROR]"), getSource(), msg) }

func Warn(msg string) { log.Printf("%s\t%s\t%s", color.YellowString("[WARN]"), getSource(), msg) }

func Debug(msg string) { log.Printf("%s\t%s\t%s", color.BlueString("[DEBUG]"), getSource(), msg) }

func getSource() string {
	_, file, line, _ := runtime.Caller(2)
	_, file = filepath.Split(file)
	return fmt.Sprintf("%s:%d", file, line)
}
