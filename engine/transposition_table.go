package engine

const (
	PerftEntrySize            = 16
	MBtoBytesConversionFactor = 1024 * 1024

	PerftEntryNodesMask = 0xffffffffffffff
	PeftEntryDepthMask  = 0xf00000000000000

	NumBuckets = 4
)

type TTEntry interface {
	Hash() uint64
	Depth() uint8
}

type PerftEntry struct {
	hash          uint64
	nodesAndDepth uint64
}

func (entry PerftEntry) Hash() uint64 {
	return entry.hash
}

func (entry PerftEntry) Depth() uint8 {
	return uint8(entry.nodesAndDepth & PeftEntryDepthMask >> 56)
}

func (entry PerftEntry) Nodes() uint64 {
	return entry.nodesAndDepth & PerftEntryNodesMask
}

func (entry *PerftEntry) SetData(hash, nodes uint64, depth uint8) {
	entry.hash = hash
	entry.nodesAndDepth = 0
	entry.nodesAndDepth |= nodes
	entry.nodesAndDepth |= (uint64(depth) << 56)
}

type TranspositionTable[T TTEntry] struct {
	entries []T
	size    uint64
}

func (tt *TranspositionTable[T]) SetSize(sizeInMB, entrySize uint64) {
	tt.size = sizeInMB * MBtoBytesConversionFactor / PerftEntrySize
	tt.entries = make([]T, tt.size)
}

func (tt *TranspositionTable[T]) Probe(hash uint64) *T {
	start_index := hash % tt.size
	for i := uint64(0); i < NumBuckets; i++ {
		entry := &tt.entries[(start_index+i)%tt.size]
		if (*entry).Hash() == hash {
			return entry
		}
	}
	return nil
}

func (tt *TranspositionTable[T]) Store(hash uint64, depth uint8) *T {
	start_index := hash % tt.size
	for i := uint64(0); i < NumBuckets; i++ {
		entry := &tt.entries[(start_index+i)%tt.size]
		if (*entry).Depth() <= depth {
			return entry
		}
	}
	return &tt.entries[(start_index+NumBuckets)%tt.size]
}

func (tt *TranspositionTable[T]) Clear() {
	for i := uint64(0); i < tt.size; i++ {
		tt.entries[i] = *new(T)
	}
}

func (tt *TranspositionTable[T]) Unitialize() {
	tt.entries = nil
	tt.size = 0
}
