// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

// The code below is copied from <https://github.com/opencoff/go-fasthash>

// func Hash64(seed uint64, buf []byte) uint64
TEXT ·Hash64(SB), 4, $0-40
	MOVQ seed+0(FP), AX
	MOVQ buf+8(FP), SI
	MOVQ buf_len+16(FP), CX

	MOVQ  $0x880355f21e6d1965, R8
	MOVQ  $0x2127599bf4325c37, R9
	MOVQ  CX, BX
	IMULQ R8, BX
	XORQ  BX, AX
	MOVQ  CX, BX
	SHRQ  $3, BX
	JZ    _rest

_loop8:
	MOVQ  (SI), DX
	MOVQ  DX, R10
	SHRQ  $23, R10
	XORQ  R10, DX
	IMULQ R9, DX
	MOVQ  DX, R10
	SHRQ  $47, R10
	XORQ  R10, DX
	XORQ  DX, AX
	IMULQ R8, AX

	ADDQ $8, SI
	DECQ BX
	JNZ  _loop8

_rest:
	ANDQ $7, CX
	JZ   _finish
	XORQ BX, BX
	CMPQ CX, $1
	JE   _1
	CMPQ CX, $2
	JE   _2
	CMPQ CX, $3
	JE   _3
	CMPQ CX, $4
	JE   _4
	CMPQ CX, $5
	JE   _5
	CMPQ CX, $6
	JE   _6
	XORQ DX, DX
	MOVB 6(SI), DX
	SHLQ $48, DX
	XORQ DX, BX

_6:
	XORQ DX, DX
	MOVB 5(SI), DX
	SHLQ $40, DX
	XORQ DX, BX

_5:
	XORQ DX, DX
	MOVB 4(SI), DX
	SHLQ $32, DX
	XORQ DX, BX

_4:
	XORQ DX, DX
	MOVB 3(SI), DX
	SHLQ $24, DX
	XORQ DX, BX

_3:
	XORQ DX, DX
	MOVB 2(SI), DX
	SHLQ $16, DX
	XORQ DX, BX

_2:
	XORQ DX, DX
	MOVB 1(SI), DX
	SHLQ $8, DX
	XORQ DX, BX

_1:
	XORQ  DX, DX
	MOVB  (SI), DX
	XORQ  DX, BX
	MOVQ  BX, R10
	SHRQ  $23, R10
	XORQ  R10, BX
	IMULQ R9, BX
	MOVQ  BX, R10
	SHRQ  $47, R10
	XORQ  R10, BX
	XORQ  BX, AX
	IMULQ R8, AX

_finish:
	MOVQ  AX, R10
	SHRQ  $23, R10
	XORQ  R10, AX
	IMULQ R9, AX
	MOVQ  AX, R10
	SHRQ  $47, R10
	XORQ  R10, AX
	MOVQ  AX, ret+32(FP)
	RET
