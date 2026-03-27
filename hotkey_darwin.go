//go:build darwin

package main

import "golang.design/x/hotkey"

func hotkeyModifiers() []hotkey.Modifier {
	// macOS: ⌘ (Command) + Option
	return []hotkey.Modifier{hotkey.ModOption, hotkey.ModCmd}
}
