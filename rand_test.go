package deterministic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/deterministic"
)

func TestRandFunc_ReturnsFunction(t *testing.T) {
	rf := deterministic.RandFunc()
	require.NotNil(t, rf)
}

func TestRandFunc_FirstCallSignature(t *testing.T) {
	rf := deterministic.RandFunc()
	buf := make([]byte, 10)

	n, err := rf(buf)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, []uint8{0xde, 0xad, 0xbe, 0xef}, buf[:4])
}

func TestRandFunc_FirstCallIncrementing(t *testing.T) {
	rf := deterministic.RandFunc()
	buf := make([]byte, 10)
	_, err := rf(buf)
	require.NoError(t, err)

	for i := 4; i < 10; i++ {
		expected := byte(i - 4)
		if buf[i] != expected {
			t.Errorf("buf[%d] = %d, want %d", i, buf[i], expected)
		}
	}
}

func TestRandFunc_SecondCallContinuesSequence(t *testing.T) {
	rf := deterministic.RandFunc()
	buf1 := make([]byte, 10)
	_, err := rf(buf1)
	require.NoError(t, err)

	buf2 := make([]byte, 10)
	n, err := rf(buf2)

	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
	if n != 10 {
		t.Errorf("got n=%d, want 10", n)
	}

	// buf2 should continue from buf1, no new signature
	for i := 0; i < 10; i++ {
		expected := byte(i + 6) // Continue from where buf1 left off (6, 7, 8, 9, 0, 1, 2, 3, 4, 5)
		if buf2[i] != expected {
			t.Errorf("buf2[%d] = %d, want %d", i, buf2[i], expected)
		}
	}
}

func TestRandFunc_RollsOver(t *testing.T) {
	rf := deterministic.RandFunc()

	// Read 4 bytes for signature
	buf := make([]byte, 4)
	_, err := rf(buf)
	require.NoError(t, err)

	// Read enough to cross byte boundary at 256
	buf = make([]byte, 256)
	_, err = rf(buf)
	require.NoError(t, err)
	assert.Equal(t, uint8(255), buf[255], "before rollover, last byte should be 0xff")

	buf = make([]byte, 5)
	_, err = rf(buf)
	require.NoError(t, err)
	assert.Zero(t, buf[0], "after rollover, first byte should wrap to 0x00")
}

func TestRandFunc_MultipleInstances(t *testing.T) {
	rf1 := deterministic.RandFunc()
	rf2 := deterministic.RandFunc()

	buf1 := make([]byte, 10)
	buf2 := make([]byte, 10)

	_, err := rf1(buf1)
	require.NoError(t, err)
	_, err = rf2(buf2)
	require.NoError(t, err)

	// Both should have same signature
	for i := 0; i < 4; i++ {
		if buf1[i] != buf2[i] {
			t.Errorf("instance 1 buf[%d]=%d != instance 2 buf[%d]=%d", i, buf1[i], i, buf2[i])
		}
	}

	// Both should have same incrementing sequence
	for i := 4; i < 10; i++ {
		if buf1[i] != buf2[i] {
			t.Errorf("instance 1 buf[%d]=%d != instance 2 buf[%d]=%d", i, buf1[i], i, buf2[i])
		}
	}
}

func TestRandFunc_SmallBuffer(t *testing.T) {
	rf := deterministic.RandFunc()
	buf := make([]byte, 2)

	n, err := rf(buf)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, []uint8{0xde, 0xad}, buf[:2])
}

func TestRandFunc_LargeBuffer(t *testing.T) {
	rf := deterministic.RandFunc()
	buf := make([]byte, 1000)

	n, err := rf(buf)
	require.NoError(t, err)
	assert.Equal(t, 1000, n)
	assert.Equal(t, []uint8{0xde, 0xad, 0xbe, 0xef}, buf[:4], "signature should be at start")
}
