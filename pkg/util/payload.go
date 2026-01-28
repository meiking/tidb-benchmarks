package util

// MakePayload returns a deterministic payload of the requested size.
func MakePayload(size int) []byte {
	if size <= 0 {
		return nil
	}
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = byte('a' + (i % 26))
	}
	return b
}
