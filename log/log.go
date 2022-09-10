package log

import (
	"fmt"

	"github.com/fatih/color"
)

func Info(text ...interface{}) {
	fmt.Printf("%v%v%v%s\n", color.HiBlackString("("), color.GreenString("+"), color.HiBlackString(") "), fmt.Sprint(text...))
}

func Error(text ...interface{}) {
	fmt.Printf("%v%v%v%s\n", color.HiBlackString("("), color.RedString("!"), color.HiBlackString(") "), fmt.Sprint(text...))
}

func Warn(text ...interface{}) {
	fmt.Printf("%v%v%v%s\n", color.HiBlackString("("), color.YellowString("-"), color.HiBlackString(") "), fmt.Sprint(text...))
}
