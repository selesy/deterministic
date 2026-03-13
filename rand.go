package deterministic

// RandFunc returns a deterministic function with the same signature as crypto/rand.Read.
// Each call to RandFunc returns an isolated instance with its own byte offset.
// The returned function fills the buffer with a signature sequence (0xdeadbeef)
// followed by incrementing bytes that roll over at 256.
func RandFunc() func([]byte) (int, error) {
	offset := uint64(0) // Global byte offset
	signatureBytes := []byte{0xde, 0xad, 0xbe, 0xef}

	return func(b []byte) (int, error) {
		for i := 0; i < len(b); i++ {
			if offset < uint64(len(signatureBytes)) {
				// Signature phase
				b[i] = signatureBytes[offset]
			} else {
				// Incrementing phase
				b[i] = byte((offset - uint64(len(signatureBytes))) % 256)
			}
			offset++
		}

		return len(b), nil
	}
}
