package main

import (
	"io"
	"math/rand"
	"os"
	"time"
)

const (
	height = byte(0x20)
	width  = byte(0x40)
)

type CPU struct {
	pc            uint16
	sp            uint16
	memory        [4096]byte
	stack         [16]uint16
	V             [16]byte
	I             uint16
	delayTimer    byte
	soundTimer    byte
	display       [height][width]byte
	keys          [16]byte
	draw          bool
	inputflag     bool
	inputRegister byte
}

var fontset = [...]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func NewCPU() CPU {
	c := CPU{pc: 0x200}
	c.LoadFontSet()
	return c
}

func (c *CPU) LoadFontSet() {
	for i := 0x00; i < 0x50; i++ {
		c.memory[i] = fontset[i]
	}
}

func (c *CPU) ClearDisplay() {
	for x := 0x00; x < 0x20; x++ {
		for y := 0x00; y < 0x40; y++ {
			c.display[x][y] = 0x00
		}
	}
}

func (c *CPU) LoadProgram(rom string) int {
	file, err := os.Open(rom)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	memory := make([]byte, 0xE00)
	n, err := file.Read(memory)

	if err != nil && err != io.EOF {
		panic(err)
	}

	for index, byte := range memory {
		// 0x200 is the start of the program
		c.memory[0x200+index] = byte
	}

	return n
}

func (c *CPU) Run() {
	c.RunCpuCycle()

	if c.delayTimer > 0 {
		c.delayTimer--
	}

	if c.soundTimer > 0 {
		c.soundTimer--
	}
}

func (c *CPU) RunCpuCycle() {
	// instruction identifier
	opcode := uint16(c.memory[c.pc]<<8 | c.memory[c.pc+1])

	c.pc = c.pc + 2

	// identify opcode
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode & 0x000F {
		case 0x0000:
			// 00E0: CLS
			c.ClearDisplay()
		case 0x000E:
			// 0x00EE: RET
			c.pc = c.stack[c.sp-1]
			c.sp--
		}
	case 0x1000:
		// 1NNN: JP addr
		c.pc = opcode & 0x0FFF
	case 0x2000:
		// 2NNN: CALL addr
		c.stack[c.sp] = c.pc
		c.sp++
		c.pc = opcode & 0x0FFF
	case 0x3000:
		// 3XKK: SE Vx, byte
		compareTo := byte(opcode & 0x00FF)
		register := byte(opcode & 0x0F00 >> 8)
		if c.V[register] == compareTo {
			c.pc += 2
		}
	case 0x4000:
		// 4xKK SNE Vx, byte
		compareTo := byte(opcode & 0x00FF)
		register := byte(opcode & 0x0F00 >> 8)
		if c.V[register] != compareTo {
			c.pc += 2
		}
	case 0x5000:
		// 5xy0 SE Vx, Vy
		rX := opcode & 0x0F00 >> 8
		rY := opcode & 0x00F0 >> 4
		if c.V[rX] == c.V[rY] {
			c.pc += 2
		}
	case 0x6000:
		// 6xkk LD Vx, byte
		register := byte(opcode & 0xF00 >> 8)
		value := byte(opcode & 0x00FF)
		c.V[register] = value
	case 0x7000:
		// 7xkk - ADD Vx, byte
		register := byte(opcode & 0x0F00 >> 8)
		c.V[register] += byte(opcode & 0x00FF)
	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000:
			// 8xy0 - LD Vx, Vy
			rX := opcode & 0x0F00 >> 8
			rY := opcode & 0x00F0 >> 4
			c.V[rX] = c.V[rY]
		case 0x0001:
			// 8xy1 - OR Vx, Vy
			rX := opcode & 0x0F00 >> 8
			rY := opcode & 0x00F0 >> 4
			c.V[rX] = c.V[rX] | c.V[rY]
		case 0x0002:
			// 8xy2 - AND Vx, Vy
			rX := opcode & 0x0F00 >> 8
			rY := opcode & 0x00F0 >> 4
			c.V[rX] = c.V[rX] & c.V[rY]
		case 0x0003:
			// 8xy3 - XOR Vx, Vy
			rX := opcode & 0x0F00 >> 8
			rY := opcode & 0x00F0 >> 4
			c.V[rX] = c.V[rX] ^ c.V[rY]
		case 0x0004:
			// 8xy4 - ADD Vx, Vy
			rX := opcode & 0x0F00 >> 8
			rY := opcode & 0x00F0 >> 4
			c.V[rX] = c.V[rX] + c.V[rY]
			if uint16(c.V[rX])+uint16(c.V[rY]) > 0xFF {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
		case 0x0005:
			// 8xy5 - SUB Vx, Vy
			rX := opcode & 0x0F00 >> 8
			rY := opcode & 0x00F0 >> 4
			if c.V[rX] > c.V[rY] {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[rX] -= c.V[rY]
		case 0x0006:
			// 0x8xy6 - SHR Vx {, Vy}
			rX := (opcode & 0x0F00) >> 8
			if c.V[rX]&0x1 == 1 {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[rX] >>= 1
		case 0x0007:
			// 0x8xy7 - SUBN Vx, Vy
			rX := (opcode & 0x0F00) >> 8
			rY := (opcode & 0x00F0) >> 4
			if c.V[rY] > c.V[rX] {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[rX] = c.V[rY] - c.V[rX]
		case 0x000E:
			// 0x8xyE - SHL Vx {, Vy}
			rX := (opcode & 0x0F00) >> 8
			if c.V[rX]&0x1 == 1 {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[rX] <<= 1
		}
	case 0x9000:
		// 0x9xy0 - SNE Vx, Vy
		rX := (opcode & 0x0F00) >> 8
		rY := (opcode & 0x00F0) >> 4
		if c.V[rX] != c.V[rY] {
			c.pc += 2
		}
	case 0xA000:
		// 0xAnnn - LD I, addr
		c.I = opcode & 0x0FFF
	case 0xB000:
		// Bnnn - JP V0, addr
		c.pc = (opcode & 0x0FFF) + uint16(c.V[0x0])
	case 0xC000:
		// Cxkk - RND Vx, byte
		rX := (opcode & 0x0F00 >> 8)
		v := byte(opcode & 0x00FF)
		rand.Seed(time.Now().Unix())
		c.V[rX] = byte(rand.Intn(256)) & v
	case 0xD000:
		// Dxyn - DRw Vx, Vy, nibble - THIS IS THE MOST HARD INSTRUCTION TO IMPLEMENT
		rX := opcode & 0x0F00 >> 8
		rY := opcode & 0x00F0 >> 4
		nibble := byte(opcode & 0x000F)

		x := c.V[rX]
		y := c.V[rY]

		c.V[0xF] = 0

		for i := y; i < y+nibble; i++ {
			for j := x; j < x+8; j++ {
				bit := (c.memory[c.I+uint16(i-y)] >> (7 - j + x)) & 0x01
				xIndex, yIndex := j, i
				if j >= width {
					xIndex = j - width
				}
				if i >= height {
					yIndex = i - height
				}
				if bit == 0x01 && c.display[yIndex][xIndex] == 0x01 {
					c.V[0xF] = 0x01
				}
				c.display[yIndex][xIndex] = c.display[yIndex][xIndex] ^ bit
			}
		}

		c.draw = true
	case 0xE000:
		switch opcode & 0x00FF {
		case 0x009E:
			// Ex9E - SKP Vx
			rX := opcode & 0x0F00 >> 8
			x := c.V[rX]
			if c.keys[x] == 0x1 {
				c.pc += 2
			}
		case 0x00A1:
			// ExA1 - SKNP Vx
			rX := opcode & 0x0F00 >> 8
			x := c.V[rX]
			if c.keys[x] == 0x0 {
				c.pc += 2
			}
		}
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007:
			// Fx07 - LD Vx, DT
			rX := opcode & 0x0F00 >> 8
			c.V[rX] = c.delayTimer
		case 0x000A:
			// Fx0A - LD Vx, K
			rX := opcode & 0x0F00 >> 8
			c.inputflag = true
			c.inputRegister = byte(rX)
		case 0x0015:
			// Fx15 - LD DT, Vx
			rX := opcode & 0x0F00 >> 8
			c.delayTimer = c.V[rX]
		case 0x0018:
			// Fx18 - LD ST, Vx
			rX := opcode & 0x0F00 >> 8
			c.soundTimer = c.V[rX]
		case 0x001E:
			// Fx1E - ADD I, Vx
			rX := opcode & 0x0F00 >> 8
			c.I += uint16(c.V[rX])
		case 0x0029:
			// Fx29 - LD F, Vx
			rX := opcode & 0x0F00 >> 8
			c.I = uint16(c.V[rX] * 5)
		case 0x0033:
			// Fx33 - LD B, Vx
			rX := opcode & 0x0F00 >> 8
			number := c.V[rX]
			c.memory[c.I] = (number / 100) % 10
			c.memory[c.I+1] = (number / 10) % 10
			c.memory[c.I+2] = number % 10
		case 0x0055:
			// Fx55 - LD [I], Vx
			rX := opcode & 0x0F00 >> 8
			for i := uint16(0x0); i <= rX; i++ {
				c.memory[c.I+i] = c.V[i]
			}
		case 0x0065:
			// Fx65 - LD Vx, [I]
			rX := opcode & 0x0F00 >> 8
			for i := uint16(0x0); i <= rX; i++ {
				c.V[i] = c.memory[c.I+i]
			}
		}
	}
}
