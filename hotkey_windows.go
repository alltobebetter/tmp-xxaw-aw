//go:build windows

package main

import "golang.design/x/hotkey"

func hotkeyModifiers() []hotkey.Modifier {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}
}
