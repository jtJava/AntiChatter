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
	// Emulates pressing and releasing the 'A' key.
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	// Buffer size is depends on your need. The 100 is placeholder value.
	keyboardChan := make(chan types.KeyboardEvent, 1500)

	if err := keyboard.Install(handler, keyboardChan); err != nil {
		return err
	}

	defer keyboard.Uninstall()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	fmt.Println("start capturing keyboard input")

	for {
		select {
		case <-time.After(5 * time.Minute):
			fmt.Println("Received timeout signal")
			return nil
		case <-signalChan:
			fmt.Println("Received shutdown signal")
			return nil
		case k := <-keyboardChan:
			k.Message.String()
			//fmt.Printf("Received %V %v\n", k.Message, k.VKCode)
			continue
		}
	}
}

var keyToTimeMap = make(map[types.VKCode]uint32)

func handler(c chan<- types.KeyboardEvent) types.HOOKPROC {
	counter := 0

	return func(code int32, wParam, lParam uintptr) uintptr {
		key := (*types.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		message := types.Message(wParam)

		if lParam == 0 {
			goto NEXT
		}

		c <- types.KeyboardEvent{
			Message:         types.Message(wParam),
			KBDLLHOOKSTRUCT: *(*types.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam)),
		}

		if counter == 1 {
			counter = 0
			goto NEXT
		}

		defer func() {
			keyToTimeMap[key.VKCode] = key.Time
		}()

		if key.Time-keyToTimeMap[key.VKCode] <= 25 && message == 256 {
			counter = 1
			println("Cancelled key press with delta:", key.Time-keyToTimeMap[key.VKCode])
			return 1
		}
	NEXT:
		return win32.CallNextHookEx(0, code, wParam, lParam)
	}
}
