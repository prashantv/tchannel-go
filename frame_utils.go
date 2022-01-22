package tchannel

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

type RecordingFramePool struct {
	sync.Mutex

	allocations map[*Frame]string
	badRelease  []string
}

func NewRecordingFramePool() *RecordingFramePool {
	return &RecordingFramePool{
		allocations: make(map[*Frame]string),
	}
}

func (p *RecordingFramePool) Get() *Frame {
	p.Lock()
	defer p.Unlock()

	frame := NewFrame(MaxFramePayloadSize)
	p.allocations[frame] = recordStack()
	return frame
}

func (p *RecordingFramePool) Release(f *Frame) {
	// Make sure the payload is not used after this point by clearing the frame.
	zeroOut(f.Payload)
	f.Payload = nil
	zeroOut(f.buffer)
	f.buffer = nil
	zeroOut(f.headerBuffer)
	f.headerBuffer = nil
	f.Header = FrameHeader{}

	p.Lock()
	defer p.Unlock()

	if _, ok := p.allocations[f]; !ok {
		p.badRelease = append(p.badRelease, "bad Release at "+recordStack())
		return
	}

	delete(p.allocations, f)
}

func (p *RecordingFramePool) CheckEmpty() (int, string) {
	p.Lock()
	defer p.Unlock()

	var badCalls []string
	badCalls = append(badCalls, p.badRelease...)
	for f, s := range p.allocations {
		badCalls = append(badCalls, fmt.Sprintf("frame %p: %v not released, get from: %v", f, f.Header, s))
	}
	return len(p.allocations), strings.Join(badCalls, "\n")
}

func recordStack() string {
	buf := make([]byte, 4096)
	runtime.Stack(buf, false)
	return string(buf)
}

func zeroOut(bs []byte) {
	for i := range bs {
		bs[i] = 0
	}
}
