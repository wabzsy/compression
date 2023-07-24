//go:build windows
// +build windows

package rtl

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	COMPRESSION_FORMAT_NONE        = 0x0000
	COMPRESSION_FORMAT_DEFAULT     = 0x0001
	COMPRESSION_FORMAT_LZNT1       = 0x0002
	COMPRESSION_FORMAT_XPRESS      = 0x0003
	COMPRESSION_FORMAT_XPRESS_HUFF = 0x0004
	COMPRESSION_FORMAT_XP10        = 0x0005
	COMPRESSION_ENGINE_STANDARD    = 0x0000
	COMPRESSION_ENGINE_MAXIMUM     = 0x0100
	COMPRESSION_ENGINE_HIBER       = 0x0200
)

var (
	modNtdll                           = syscall.NewLazyDLL("ntdll.dll")
	procRtlGetCompressionWorkSpaceSize = modNtdll.NewProc("RtlGetCompressionWorkSpaceSize")
	procRtlCompressBuffer              = modNtdll.NewProc("RtlCompressBuffer")
	procRtlDecompressBuffer            = modNtdll.NewProc("RtlDecompressBuffer")
)

// RtlDecompressBuffer
// https://learn.microsoft.com/zh-cn/windows-hardware/drivers/ddi/ntifs/nf-ntifs-rtldecompressbuffer
// NT_RTL_COMPRESS_API NTSTATUS RtlDecompressBuffer(
//
//	[in]  USHORT CompressionFormat,
//	[out] PUCHAR UncompressedBuffer,
//	[in]  ULONG  UncompressedBufferSize,
//	[in]  PUCHAR CompressedBuffer,
//	[in]  ULONG  CompressedBufferSize,
//	[out] PULONG FinalUncompressedSize
//
// );
func RtlDecompressBuffer(
	CompressionFormat uint16, // in
	UncompressedBuffer *byte, // out
	UncompressedBufferSize uint32, // in
	CompressedBuffer *byte, // in
	CompressedBufferSize uint32, // in
	FinalCompressedSize *uint32, // out
) error {
	r0, _, _ := syscall.SyscallN(
		procRtlDecompressBuffer.Addr(),
		uintptr(CompressionFormat),
		uintptr(unsafe.Pointer(UncompressedBuffer)),
		uintptr(UncompressedBufferSize),
		uintptr(unsafe.Pointer(CompressedBuffer)),
		uintptr(CompressedBufferSize),
		uintptr(unsafe.Pointer(FinalCompressedSize)),
	)

	if r0 != 0 {
		return fmt.Errorf("error: (NTSTATUS)0x%08x", r0)
	}
	return nil
}

// RtlCompressBuffer
// https://learn.microsoft.com/zh-cn/windows-hardware/drivers/ddi/ntifs/nf-ntifs-rtlcompressbuffer
// NT_RTL_COMPRESS_API NTSTATUS RtlCompressBuffer(
//
//	[in]  USHORT CompressionFormatAndEngine,
//	[in]  PUCHAR UncompressedBuffer,
//	[in]  ULONG  UncompressedBufferSize,
//	[out] PUCHAR CompressedBuffer,
//	[in]  ULONG  CompressedBufferSize,
//	[in]  ULONG  UncompressedChunkSize,
//	[out] PULONG FinalCompressedSize,
//	[in]  PVOID  WorkSpace
//
// );
func RtlCompressBuffer(
	CompressionFormatAndEngine uint16, // in
	UncompressedBuffer *byte, // in
	UncompressedBufferSize uint32, // in
	CompressedBuffer *byte, // out
	CompressedBufferSize uint32, // in
	UncompressedChunkSize uint32, // in
	FinalCompressedSize *uint32, // out
	WorkSpace *byte, // in
) error {
	r0, _, _ := syscall.SyscallN(
		procRtlCompressBuffer.Addr(),
		uintptr(CompressionFormatAndEngine),
		uintptr(unsafe.Pointer(UncompressedBuffer)),
		uintptr(UncompressedBufferSize),
		uintptr(unsafe.Pointer(CompressedBuffer)),
		uintptr(CompressedBufferSize),
		uintptr(UncompressedChunkSize),
		uintptr(unsafe.Pointer(FinalCompressedSize)),
		uintptr(unsafe.Pointer(WorkSpace)),
	)

	if r0 != 0 {
		return fmt.Errorf("error: (NTSTATUS)0x%08x", r0)
	}
	return nil
}

// RtlGetCompressionWorkSpaceSize
//
// https://learn.microsoft.com/zh-cn/windows-hardware/drivers/ddi/ntifs/nf-ntifs-rtlgetcompressionworkspacesize
//
// NT_RTL_COMPRESS_API NTSTATUS RtlGetCompressionWorkSpaceSize(
//
//	[in]  USHORT CompressionFormatAndEngine,
//	[out] PULONG CompressBufferWorkSpaceSize,
//	[out] PULONG CompressFragmentWorkSpaceSize
//
// );
func RtlGetCompressionWorkSpaceSize(
	CompressionFormatAndEngine uint16, // in
	CompressBufferWorkSpaceSize, // out
	CompressFragmentWorkSpaceSize *uint32, // out
) error {
	r0, _, _ := syscall.SyscallN(
		procRtlGetCompressionWorkSpaceSize.Addr(),
		uintptr(CompressionFormatAndEngine),
		uintptr(unsafe.Pointer(CompressBufferWorkSpaceSize)),
		uintptr(unsafe.Pointer(CompressFragmentWorkSpaceSize)),
	)

	if r0 != 0 {
		return fmt.Errorf("error: (NTSTATUS)0x%08x", r0)
	}
	return nil
}

// RtlCompress
// compressionFormat: COMPRESSION_FORMAT_LZNT1 / COMPRESSION_FORMAT_XPRESS
func RtlCompress(compressionFormat uint16, source []byte) ([]byte, error) {
	var wSpace, fSpace uint32

	if err := RtlGetCompressionWorkSpaceSize(
		compressionFormat|COMPRESSION_ENGINE_MAXIMUM,
		&wSpace,
		&fSpace,
	); err != nil {
		return nil, err
	}

	workSpace := make([]byte, wSpace)

	var finalCompressedSize uint32

	compressedBuffer := make([]byte, len(source)*2) // 此处*2是避免传如的内容已被压缩过,无法再次压缩(再次压缩会变大)的情况

	if err := RtlCompressBuffer(
		compressionFormat|COMPRESSION_ENGINE_MAXIMUM,
		&source[0],
		uint32(len(source)),
		&compressedBuffer[0],
		uint32(len(compressedBuffer)),
		0,
		&finalCompressedSize,
		&workSpace[0],
	); err != nil {
		return nil, err
	}

	return compressedBuffer[:finalCompressedSize], nil
}

// RtlDecompress
// compressionFormat: COMPRESSION_FORMAT_LZNT1 / COMPRESSION_FORMAT_XPRESS
func RtlDecompress(compressionFormat uint16, source []byte, uncompressedBufferSize uint32) ([]byte, error) {

	var finalCompressedSize uint32

	uncompressedBuffer := make([]byte, uncompressedBufferSize)

	if err := RtlDecompressBuffer(
		compressionFormat,
		&uncompressedBuffer[0],
		uncompressedBufferSize,
		&source[0],
		uint32(len(source)),
		&finalCompressedSize,
	); err != nil {
		return nil, err
	}

	return uncompressedBuffer[:finalCompressedSize], nil
}

func RtlDecompressWithDefaultBufferSize(compressionFormat uint16, source []byte) ([]byte, error) {
	// TODO:
	//  此处需要优化, 因为使用场景无法提前预知解压后的大小, 所以:
	//    假设压缩率不会低于1/16(6.25%), 分配固定大小(压缩文件x16)的空间
	//  解压1G的文件件需要申请16G的空间, 很离谱
	//  但是shellcode一般不会太大, 按shellcode压缩后 10M 的极端情况计算, 需要 160M 的内存来解压, 勉强能接受吧..
	return RtlDecompress(compressionFormat, source, uint32(len(source)*16))
}

func LZNT1Compress(source []byte) ([]byte, error) {
	return RtlCompress(COMPRESSION_FORMAT_LZNT1, source)
}

func LZNT1Decompress(source []byte) ([]byte, error) {
	return RtlDecompressWithDefaultBufferSize(COMPRESSION_FORMAT_LZNT1, source)
}

func XPressCompress(source []byte) ([]byte, error) {
	return RtlCompress(COMPRESSION_FORMAT_XPRESS, source)
}

func XPressDecompress(source []byte) ([]byte, error) {
	return RtlDecompressWithDefaultBufferSize(COMPRESSION_FORMAT_XPRESS, source)
}
