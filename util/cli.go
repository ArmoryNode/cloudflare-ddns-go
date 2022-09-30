package util

import (
	"fmt"
	"time"
)

var spinner = [4]rune{'/', '-', '\\', '|'}

func DisplaySpinner(text string, quit chan bool) {
	count := 0
	for {
		select {
		case <-quit:
			return
		default:
			fmt.Printf("[%c]\u0020%s\r", spinner[count%4], text)
		}
		count++
		time.Sleep(time.Millisecond * 150)
	}
}
