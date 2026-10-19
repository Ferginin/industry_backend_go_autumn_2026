package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	checkedName := strings.TrimSpace(name)
	if checkedName == "" {
		return "Hello, World!"
	} else {
		return fmt.Sprintf("Hello, %s!", checkedName)
	}
}
