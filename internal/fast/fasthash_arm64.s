// The code below is generated using ChatGPT-4o based on the hash64() function in
// fasthash.go and the the amd64 assembly code for Hash64() in fasthash_amd64.s.

// func Hash64(seed uint64, buf []byte) uint64
TEXT ·Hash64(SB), 4, $0-40
    // Load parameters
    MOVD seed+0(FP), R0    // Load seed into R0
    MOVD buf+8(FP), R1     // Load buf into R1
    MOVD buf_len+16(FP), R2  // Load len(buf) into R2

    MOVD  $0x880355f21e6d1965, R8   // Load constant m into R8
    MOVD  $0x2127599bf4325c37, R9   // Load constant for mix into R9

    // h := seed ^ (uint64(len(buf)) * m)
    MOVD  R2, R3
    MUL   R3, R8, R3
    EOR   R0, R3, R0

    // len(buf) / 8
    MOVD  R2, R4
    LSR   $3, R4, R4
    CBZ   R4, _rest

_loop8:
    // Load 8 bytes from buf
    MOVD  (R1), R5
    MOVD  R5, R6
    LSR   $23, R6, R6
    EOR   R5, R6, R5
    MUL   R5, R9, R5
    MOVD  R5, R6
    LSR   $47, R6, R6
    EOR   R5, R6, R5
    EOR   R5, R0, R0
    MUL   R0, R8, R0

    ADD   $8, R1, R1
    SUBS  $1, R4, R4
    BNE   _loop8

_rest:
    AND   $7, R2, R2
    CBZ   R2, _finish

    MOVD  $0, R5
    CMP   $1, R2
    BEQ   _1
    CMP   $2, R2
    BEQ   _2
    CMP   $3, R2
    BEQ   _3
    CMP   $4, R2
    BEQ   _4
    CMP   $5, R2
    BEQ   _5
    CMP   $6, R2
    BEQ   _6

_6:
    MOVBU 6(R1), R6
    LSL   $48, R6, R6
    EOR   R6, R5, R5

_5:
    MOVBU 5(R1), R6
    LSL   $40, R6, R6
    EOR   R6, R5, R5

_4:
    MOVBU 4(R1), R6
    LSL   $32, R6, R6
    EOR   R6, R5, R5

_3:
    MOVBU 3(R1), R6
    LSL   $24, R6, R6
    EOR   R6, R5, R5

_2:
    MOVBU 2(R1), R6
    LSL   $16, R6, R6
    EOR   R6, R5, R5

_1:
    MOVBU 1(R1), R6
    LSL   $8, R6, R6
    EOR   R6, R5, R5

    MOVBU 0(R1), R6
    EOR   R6, R5, R5
    MOVD  R5, R6
    LSR   $23, R6, R6
    EOR   R6, R5, R5
    MUL   R5, R9, R5
    MOVD  R5, R6
    LSR   $47, R6, R6
    EOR   R6, R5, R5
    EOR   R5, R0, R0
    MUL   R0, R8, R0

_finish:
    MOVD  R0, R6
    LSR   $23, R6, R6
    EOR   R6, R0, R0
    MUL   R0, R9, R0
    MOVD  R0, R6
    LSR   $47, R6, R6
    EOR   R6, R0, R0
    MOVD  R0, ret+32(FP)
    RET
