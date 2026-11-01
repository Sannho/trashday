package cleaner

import (
	"syscall"
	"unsafe"
)

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	shQueryRecycleBin = shell32.NewProc("SHQueryRecycleBinW")
	shEmptyRecycleBin = shell32.NewProc("SHEmptyRecycleBinW")
)

// shQueryRBInfo is SHQUERYRBINFO; Go's natural alignment matches the header on 386 and amd64.
type shQueryRBInfo struct {
	cbSize   uint32
	size     int64
	numItems int64
}

const sherbQuiet = 0x1 | 0x2 | 0x4 // SHERB_NOCONFIRMATION | SHERB_NOPROGRESSUI | SHERB_NOSOUND

// recycleBin measures, then empties, the Recycle Bin of every drive through the shell,
// which knows where each drive keeps its bin and which entries belong to the user.
func recycleBin(dry bool) Result {
	info := shQueryRBInfo{cbSize: uint32(unsafe.Sizeof(shQueryRBInfo{}))}
	if hr, _, _ := shQueryRecycleBin.Call(0, uintptr(unsafe.Pointer(&info))); hr != 0 {
		return Result{Failed: 1}
	}
	if dry || info.numItems == 0 { // emptying an empty bin returns an error
		return Result{Freed: info.size}
	}
	if hr, _, _ := shEmptyRecycleBin.Call(0, 0, sherbQuiet); hr != 0 {
		return Result{Failed: 1}
	}
	return Result{Freed: info.size}
}
