package gst

/*
#include "gst.go.h"
*/
import "C"

import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

// Memory is a go representation of GstMemory. This object is implemented in a read-only fashion
// currently primarily for reference, and as such you should not really use it. You can create new
// memory blocks, but there are no methods implemented yet for modifying ones already in existence.
//
// Use the Buffer and its Map methods to interact with memory in both a read and writable way.
type Memory struct {
	ptr     *C.GstMemory
	mapInfo *MapInfo
}

// NewMemoryWrapped allocates a new memory block that wraps the given data.
//
// The prefix/padding must be filled with 0 if flags contains MemoryFlagZeroPrefixed
// and MemoryFlagZeroPadded respectively.
func NewMemoryWrapped(flags MemoryFlags, data []byte, maxSize, offset, size int64) *Memory {
	str := string(data)
	dataPtr := unsafe.Pointer(C.CString(str))
	mem := C.gst_memory_new_wrapped(
		C.GstMemoryFlags(flags),
		(C.gpointer)(dataPtr),
		C.gsize(maxSize),
		C.gsize(offset),
		C.gsize(size),
		nil, // TODO: Allow user to set userdata for destroy notify function
		nil, // TODO: Allow user to set destroy notify function
	)
	return wrapMemory(mem)
}

// Instance returns the underlying GstMemory instance.
func (m *Memory) Instance() *C.GstMemory { return C.toGstMemory(unsafe.Pointer(m.ptr)) }

// Ref increases the ref count on this memory block by one.
func (m *Memory) Ref() *Memory {
	return wrapMemory(C.gst_memory_ref(m.Instance()))
}

// Unref decreases the ref count on this memory block by one. When the refcount reaches
// zero the memory is freed.
func (m *Memory) Unref() { C.gst_memory_unref(m.Instance()) }

// Allocator returns the allocator for this memory.
func (m *Memory) Allocator() *Allocator {
	return wrapAllocator(&glib.Object{GObject: glib.ToGObject(unsafe.Pointer(m.Instance().allocator))})
}

// Parent returns this memory block's parent.
func (m *Memory) Parent() *Memory { return wrapMemory(m.Instance().parent) }

// MaxSize returns the maximum size allocated for this memory block.
func (m *Memory) MaxSize() int64 { return int64(m.Instance().maxsize) }

// Alignment returns the alignment of the memory.
func (m *Memory) Alignment() int64 { return int64(m.Instance().align) }

// Offset returns the offset where valid data starts.
func (m *Memory) Offset() int64 { return int64(m.Instance().offset) }

// Size returns the size of valid data.
func (m *Memory) Size() int64 { return int64(m.Instance().size) }

// Copy returns a copy of size bytes from mem starting from offset. This copy is
// guaranteed to be writable. size can be set to -1 to return a copy from offset
// to the end of the memory region.
func (m *Memory) Copy(offset, size int64) *Memory {
	mem := C.gst_memory_copy(m.Instance(), C.gssize(offset), C.gssize(size))
	return wrapMemory(mem)
}

// Map the data inside the memory. This function can return nil if the memory is not read or writable.
//
// Unmap the Memory after usage.
func (m *Memory) Map(flags MapFlags) *MapInfo {
	mapInfo := C.malloc(C.sizeof_GstMapInfo)
	C.gst_memory_map(
		(*C.GstMemory)(m.Instance()),
		(*C.GstMapInfo)(mapInfo),
		C.GstMapFlags(flags),
	)
	if mapInfo == C.NULL {
		return nil
	}
	m.mapInfo = wrapMapInfo((*C.GstMapInfo)(mapInfo))
	return m.mapInfo
}

// Unmap will unmap the data inside this memory. Use this after calling Map on the Memory.
func (m *Memory) Unmap() {
	if m.mapInfo == nil {
		return
	}
	C.gst_memory_unmap(m.Instance(), (*C.GstMapInfo)(m.mapInfo.Instance()))
}

// Bytes will return a byte slice of the data inside this memory if it is readable.
func (m *Memory) Bytes() []byte {
	mapInfo := m.Map(MapRead)
	if mapInfo == nil {
		return nil
	}
	defer m.Unmap()
	return mapInfo.Bytes()
}
