package utils

import "github.com/fatih/color"

func Red() func(a ...interface{}) string {
	return color.New(color.FgRed).SprintFunc()
}

func Green() func(a ...interface{}) string {
	return color.New(color.FgGreen).SprintFunc()
}

func Blue() func(a ...interface{}) string {
	return color.New(color.FgBlue).SprintFunc()
}

func Yellow() func(a ...interface{}) string {
	return color.New(color.FgHiYellow).SprintFunc()
}
