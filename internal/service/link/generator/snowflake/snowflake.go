package snowflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	timestampBits = 38
	machineIDBits = 8
	sequenceBits  = 12

	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits

	maxMachineID = (1 << machineIDBits) - 1
	maxSequence  = (1 << sequenceBits) - 1
	maxTimestamp = (1 << timestampBits) - 1

	codeLength = 10
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

var snowflakeEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

type snowflakeGenerator struct {
	mu        sync.Mutex
	machineID uint64
	lastMs    int64
	sequence  uint64
	clock     func() int64
}

func New(machineID uint64) (*snowflakeGenerator, error) {
	if machineID > maxMachineID {
		return nil, fmt.Errorf("machineID %d exceeds max %d", machineID, maxMachineID)
	}
	return &snowflakeGenerator{
		machineID: machineID,
		clock:     func() int64 { return time.Now().UnixMilli() },
	}, nil
}

func (g *snowflakeGenerator) Generate() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.clock() - snowflakeEpoch
	if now < 0 || now > maxTimestamp {
		return "", errors.New("snowflake: timestamp out of range — adjust epoch")
	}

	switch {
	case now < g.lastMs:
		return "", fmt.Errorf("snowflake: clock moved backwards by %dms", g.lastMs-now)
	case now == g.lastMs:
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for now <= g.lastMs {
				now = g.clock() - snowflakeEpoch
			}
		}
	default:
		g.sequence = 0
	}
	g.lastMs = now

	id := (uint64(now) << timestampShift) |
		(g.machineID << machineIDShift) |
		g.sequence

	return encodeBase63(id), nil
}

func encodeBase63(n uint64) string {
	var buf [codeLength]byte
	for i := codeLength - 1; i >= 0; i-- {
		buf[i] = alphabet[n%63]
		n /= 63
	}
	return string(buf[:])
}
