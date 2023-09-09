# Chip-8 Emulator

This is a reproduction of [go-8](https://github.com/h4ck3rk3y/go-8/), an emulator for the [Chip-8](https://en.wikipedia.org/wiki/CHIP-8), made by [h4ck3rk3y](https://github.com/h4ck3rk3y/). I made this to learn more about emulation and to get more familiar with Golang.

After understanding the core parts of the reimplementation, I was able to implement all instructing by myself, only using [Technical Reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM) and checking for type diferences (and [cpu.go#251](/cpu.go#L251), which is a very hard instruction).