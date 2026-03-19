package main

import (
	"fmt"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) ts() string {
	return fmt.Sprintf("%s[%s]%s", colorGray, time.Now().Format("15:04:05"), colorReset)
}

func (l *Logger) Info(msg string) {
	fmt.Printf("%s %s[gwatch]%s %s\n", l.ts(), colorCyan, colorReset, msg)
}

func (l *Logger) Warn(msg string) {
	fmt.Printf("%s %s[warn]%s   %s\n", l.ts(), colorYellow, colorReset, msg)
}

func (l *Logger) Error(msg string) {
	fmt.Printf("%s %s[error]%s  %s\n", l.ts(), colorRed, colorReset, msg)
}

func (l *Logger) Change(msg string) {
	fmt.Printf("%s %s[change]%s %s\n", l.ts(), colorGreen, colorReset, msg)
}

func (l *Logger) Exec(msg string) {
	fmt.Printf("%s %s[exec]%s   %s%s%s\n", l.ts(), colorBlue, colorReset, colorYellow, msg, colorReset)
}
