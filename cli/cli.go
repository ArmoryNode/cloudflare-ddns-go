package cli

import (
	"fmt"
	"math"
	"time"
)

const UP_LINE = "\033[A"
const HIDE_CURSOR = "\033[?25l"
const CLEAR_LINE = "\x1b[2K"
const CLEAR_SCREEN = "\033[2J"
const MOVE_CURSOR_HOME = "\033[H"

const COLOR_RESET = "\033[0m"

const COLOR_RED = "\033[31m"
const COLOR_GREEN = "\033[32m"
const COLOR_YELLOW = "\033[33m"
const COLOR_BLUE = "\033[34m"
const COLOR_PURPLE = "\033[35m"
const COLOR_CYAN = "\033[36m"
const COLOR_WHITE = "\033[37m"

var spinner = [4]rune{'/', '-', '\\', '|'}

func DisplaySpinner(text string, done chan bool) {
	count := 0
	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("[%c]\u0020%s\r", spinner[count%4], text)
		}

		// Guard against integer overflow
		if count == math.MaxInt {
			count = 0
		} else {
			count++
		}

		time.Sleep(time.Millisecond * 150)
	}
}
