package tchannel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckedFramePoolForTest(t *testing.T) {
	tests := []struct {
		msg                string
		operations         func(pool *CheckedFramePoolForTest)
		wantBadAllocations int
		wantBadReleases    int
	}{
		{
			msg: "no bad releases or leaks",
			operations: func(pool *CheckedFramePoolForTest) {
				for i := 0; i < 10; i++ {
					pool.Release(pool.Get())
				}
			},
		},
		{
			msg: "frames are leaked",
			operations: func(pool *CheckedFramePoolForTest) {
				for i := 0; i < 10; i++ {
					pool.Release(pool.Get())
				}
				for i := 0; i < 10; i++ {
					_ = pool.Get()
				}
			},
			wantBadAllocations: 10,
			wantBadReleases:    0,
		},
		{
			msg: "frames are double released",
			operations: func(pool *CheckedFramePoolForTest) {
				for i := 0; i < 10; i++ {
					pool.Release(pool.Get())
				}
				f := pool.Get()
				pool.Release(f)
				pool.Release(f)
			},
			wantBadAllocations: 0,
			wantBadReleases:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			pool := NewCheckedFramePoolForTest()

			tt.operations(pool)
			results := pool.CheckEmpty()

			assert.Equal(t, tt.wantBadAllocations, len(results.UnexpectedAllocations), "Unexpected allocs")
			assert.Equal(t, tt.wantBadReleases, len(results.BadReleases), "Unexpected bad releases")
		})
	}
}
