package main

import "C"
import (
	"fmt"
	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
	"github.com/moutend/go-hook/pkg/win32"
	"log"
	"os"
	"os/signal"
	"time"
	"unsafe"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("error: ")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	// Buffer size is depends on your need. The 100 is placeholder value.
	keyboardChan := make(chan types.KeyboardEvent, 100)

	if err := keyboard.Install(handler, keyboardChan); err != nil {
		return err
	}

	defer keyboard.Uninstall()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	fmt.Println("start capturing keyboard input")
	for {
		time.Sleep(time.Millisecond * 2)
	}
}

func handler(chan<- types.KeyboardEvent) types.HOOKPROC {
	counter := 0
	keyToUpsMap := make(map[types.VKCode]uint32)

	return func(code int32, wParam, lParam uintptr) uintptr {
		key := (*types.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		message := types.Message(wParam)
		pressed := message == 256 || message == 260  // KEYDOWN || SYSKEYDOWN
		released := message == 257 || message == 261 // KEYUP || SYSKEYUP

		defer func() {
			if pressed {
				keyToUpsMap[key.VKCode] = 0
			} else if released {
				keyToUpsMap[key.VKCode] = keyToUpsMap[key.VKCode] + 1
			}
		}()

		if lParam == 0 {
			goto NEXT
		}

		if counter == 1 {
			counter = 0
			goto NEXT
		}

		if keyToUpsMap[key.VKCode] > 1 && message == 257 {
			counter = 1
			return 1
		}
	NEXT:
		return win32.CallNextHookEx(0, code, wParam, lParam)
	}
}
