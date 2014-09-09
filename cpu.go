/*
  Emulator for subset of RCA 1802 Cosmac CPU, for genetic programming.
*/
package genetic

// TODO:  Fix subtration and DF borrowing.  I'm sure it's wrong.

const SZ = 256
const STEPS_PER_TICK = 16
const TICKS_PER_TOCK = 16
const MAX_TIME = 50000

// Sim is the simulator for a Machine in a World.
type Sim struct {
	// For Simulation:
	Time  uint64
	Ticks uint64
	Tocks uint64
	M     *Mach
	W     World // The world of the simulator
}

func RunSimulation(code []byte, w World) (*Sim, bool) {
	m := NewMachine(code)
	sim := &Sim{
		M: m,
		W: w,
	}

	for sim.Time = 1; sim.Time <= MAX_TIME; sim.Time++ {
		ok := m.Step()
		if !ok {
			return sim, false
		}
		if sim.Time%STEPS_PER_TICK == 0 {
			ok = w.Tick(sim)
			if !ok {
				return sim, false
			}
			if sim.Time%(STEPS_PER_TICK*TICKS_PER_TOCK) == 0 {
				ok = w.Tock(sim)
				if !ok {
					return sim, false
				}
			}
		}
	}
	return sim, true
}

type World interface {
	Tick(sim *Sim) bool
	Tock(sim *Sim) bool
	Fitness(sim *Sim) float64
}

func NewMachine(code []byte) *Mach {
	p := new(Mach)
	copy(p.Mem[:], code)
	return p
}

// Mach is the Machine State.
type Mach struct {
	// Registers
	Mem  [SZ]byte   // Memory
	Reg  [16]uint16 // Wide Registers
	D    byte       // Accumulator
	X, P byte       // Nybble pointer to Reg
	DF   bool       // Status bits

	// Outputs.
	Q   bool    // Q output bit.
	Out [8]byte // Output ports (0 is unused)

	// Inputs.
	EF byte    // Four External Flag inputs, in low nybble, at bit (16 >> i), for i in 1..4
	In [8]byte // Input ports (0 is unused)
}

func (p *Mach) dfByte() byte {
	if p.DF {
		return 1
	}
	return 0
}
func (p *Mach) dfInt() int {
	if p.DF {
		return 1
	}
	return 0
}
func (p *Mach) imm() byte {
	p.Reg[p.P]++
	return p.Mem[p.Reg[p.P]%SZ]
}

func (p *Mach) shortBranch(cond bool) {
	p.Reg[p.P]++
	if cond {
		p.Reg[p.P] = (p.Reg[p.P] & 0xFF00) | uint16(p.Mem[p.Reg[p.P]%SZ])
		p.Reg[p.P]-- // Because it will be ++ at bottom of Step().
	}
}

// Step returns false if IDLE instruction, else true.
func (p *Mach) Step() bool {
	instr := p.Mem[p.Reg[p.P]%SZ]
	i, n := (instr>>4)&15, instr&15
	switch i {
	case 0: // LDN, but IDL if N==0
		if n == 0 {
			// NOP -- we have moved WAIT to 0x68.
		} else {
			p.D = p.Mem[p.Reg[n]%SZ]
		}
	case 1: // INC rn
		p.Reg[n]++
	case 2: // DEC rn
		p.Reg[n]--
	case 3: // BR...
		switch n {
		case 0:
			p.shortBranch(true)
		case 1:
			p.shortBranch(!p.Q)
		case 2:
			p.shortBranch(p.D == 0)
		case 3:
			p.shortBranch(p.DF)
		case 4:
			p.shortBranch(16 == 16&(p.EF<<1))
		case 5:
			p.shortBranch(16 == 16&(p.EF<<2))
		case 6:
			p.shortBranch(16 == 16&(p.EF<<3))
		case 7:
			p.shortBranch(16 == 16&(p.EF<<4))

		case 8:
			p.shortBranch(false)
		case 9:
			p.shortBranch(p.Q)
		case 10:
			p.shortBranch(p.D != 0)
		case 11:
			p.shortBranch(!p.DF)
		case 12:
			p.shortBranch(0 == 16&(p.EF<<1))
		case 13:
			p.shortBranch(0 == 16&(p.EF<<2))
		case 14:
			p.shortBranch(0 == 16&(p.EF<<3))
		case 15:
			p.shortBranch(0 == 16&(p.EF<<4))
		}
	case 4: // LDA
		p.D = p.Mem[p.Reg[n]%SZ]
		p.Reg[n]++
	case 5: // STR
		p.Mem[p.Reg[n]%SZ] = p.D
	case 6: // I/O
    // Special case 0x68:
		if n == 8 {
			return false // The spare opcode 0x68 becomes IDL/STOP (rather than opcode 0)
		}
    // Special case 0x60:
		if n == 0 {
			p.Reg[p.X]++ // IRX
		}
    // Inputs and Outputs
		if n < 8 {
			p.Out[n] = p.D // OUT n&7
			// println("OUT", n, p.D)
		} else {
			p.D = p.In[(n & 7)] // IN n&7
		}
	case 7: // misc...
		switch n {
		case 0: // RET

		case 1: // DIS

		case 2: // LDXA
			p.D = p.Mem[p.Reg[p.X]%SZ]
			p.Reg[p.X]++
		case 3: // STXD
			p.Mem[p.Reg[p.X]%SZ] = p.D
			p.Reg[p.X]--
		case 4: //
			b := p.dfByte()
			p.DF = (int(b)+int(p.D)+int(p.Mem[p.Reg[p.X]%SZ]) > 255)
			p.D += b + p.Mem[p.Reg[p.X]%SZ]
		case 5: //
			b := p.dfByte()
			p.DF = (int(p.D) > int(p.Mem[p.Reg[p.X]%SZ])+int(b))
			p.D = b + p.Mem[p.Reg[p.X]%SZ] - p.D
		case 6: // SHRC
			c := (p.D&1 == 128)
			p.D = (p.D >> 1) | (p.dfByte() << 8)
			p.DF = c
		case 7: //
			b := p.dfByte()
			p.DF = (int(p.D)+int(b) < int(p.Mem[p.Reg[p.X]%SZ]))
			p.D = b + p.D - p.Mem[p.Reg[p.X]%SZ]
		case 8: // LDI
			p.D = p.imm()
		case 9: //
			p.D |= p.imm()
		case 10: // REQ
			p.Q = false
		case 11: // SEQ
			p.Q = true
		case 12: // ADCI
			b := p.dfByte()
			imm := p.imm()
			p.DF = (int(b)+int(p.D)+int(imm) > 255)
			p.D += b + imm
		case 13: // SDBI
			b := p.dfByte()
			imm := p.imm()
			p.DF = (p.D > b+imm)
			p.D = b + imm - p.D
		case 14: // SHLC
			c := (p.D&128 == 128)
			p.D = (p.D << 1) | p.dfByte()
			p.DF = c
		case 15: // SMBI
			b := p.dfByte()
			imm := p.imm()
			p.DF = (b+p.D < imm)
			p.D = b + p.D - imm
		}
	case 8: // Get Lo
		p.D = byte(p.Reg[n])
	case 9: // Get Hi
		p.D = byte(p.Reg[n] >> 8)
	case 10: // Set Lo
		p.Reg[n] = (p.Reg[n] & 0xFF00) | uint16(p.D)
	case 11: // Set Hi
		p.Reg[n] = (p.Reg[n] & 0x00FF) | (uint16(p.D) << 8)
	case 12: // LBR...
	case 13: // Set P
		p.P = n
	case 14: // Set X
		p.X = n
	case 15: // misc...

		switch n {
		case 0: // LDX
			p.D = p.Mem[p.Reg[p.X]%SZ]
		case 1: // OR
			p.D |= p.Mem[p.Reg[p.X]%SZ]
		case 2: // AND
			p.D &= p.Mem[p.Reg[p.X]%SZ]
		case 3: // XOR
			p.D ^= p.Mem[p.Reg[p.X]%SZ]
		case 4: //
			p.DF = (int(p.D)+int(p.Mem[p.Reg[p.X]%SZ]) > 255)
			p.D += p.Mem[p.Reg[p.X]%SZ]
		case 5: // SD
			p.DF = (int(p.D) > int(p.Mem[p.Reg[p.X]%SZ]))
			p.D = p.Mem[p.Reg[p.X]%SZ] - p.D
		case 6: // SHR
			p.D = p.D >> 1
		case 7: // SM
			p.DF = (int(p.D) < int(p.Mem[p.Reg[p.X]%SZ]))
			p.D = p.D - p.Mem[p.Reg[p.X]%SZ]
		case 8: // LDI
			p.D = p.imm()
		case 9: //
			p.D |= p.imm()
		case 10: //
			p.D &= p.imm()
		case 11: //
			p.D ^= p.imm()
		case 12: //
			imm := p.imm()
			p.DF = (p.D+imm > 255)
			p.D += imm
		case 13: //
			imm := p.imm()
			p.DF = (p.D > imm)
			p.D = imm - p.D
		case 14: // SHL
			p.D = p.D << 1
		case 15: //
			imm := p.imm()
			p.DF = (p.D < imm)
			p.D = p.D - imm
		}
	}
	p.Reg[p.P]++
	return true
}
