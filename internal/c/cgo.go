//go:build freebsd

package c

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -ljail
#include "jail_wrap.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	gjsyscall "github.com/zombocoder/go-freebsd-jail/internal/syscall"
)

// mu protects concurrent access to the C layer, since jail_errmsg is a global.
var mu sync.Mutex

// Param represents a jail parameter for the C layer.
type Param struct {
	Name   string
	Value  string
	IsBool bool
}

// Jail operation flags matching sys/jail.h.
const (
	FlagCreate = 0x01
	FlagUpdate = 0x02
)

// JailSet creates or updates a jail. Returns JID on success.
func JailSet(params []Param, flags int) (int, error) {
	if len(params) == 0 {
		return -1, fmt.Errorf("%w: no parameters provided", jailerr.ErrInvalidArgument)
	}

	cparams, cleanup := buildCParams(params)
	defer cleanup()

	mu.Lock()
	res := C.gj_jail_set(cparams, C.int(len(params)), C.int(flags))
	mu.Unlock()

	if res.jid < 0 {
		return -1, gjsyscall.MapErrno(syscall.Errno(res.errnum), C.GoString(&res.errmsg[0]))
	}
	return int(res.jid), nil
}

// JailGet retrieves jail parameters. The key param identifies the jail (by name or jid).
// names lists the parameter names to fetch. Returns JID and a map of name->value.
// The key param name is automatically excluded from the fetch list to avoid conflicts.
func JailGet(key Param, names []string, flags int) (int, map[string]string, error) {
	// Filter out the key param name from fetch list to avoid duplicate param confusion
	filteredNames := make([]string, 0, len(names))
	for _, n := range names {
		if n != key.Name {
			filteredNames = append(filteredNames, n)
		}
	}

	allParams := make([]Param, 0, 1+len(filteredNames))
	allParams = append(allParams, key)
	for _, n := range filteredNames {
		allParams = append(allParams, Param{Name: n})
	}

	cparams, cleanup := buildCParams(allParams)
	defer cleanup()

	mu.Lock()
	res := C.gj_jail_get(cparams, C.int(len(allParams)), C.int(flags))
	mu.Unlock()

	if res.jid < 0 {
		return -1, nil, gjsyscall.MapErrno(syscall.Errno(res.errnum), C.GoString(&res.errmsg[0]))
	}

	values := make(map[string]string, len(filteredNames))
	slice := unsafe.Slice(cparams, len(allParams))
	for i := 1; i < len(allParams); i++ {
		cval := slice[i].value
		if cval != nil {
			values[filteredNames[i-1]] = C.GoString(cval)
			C.gj_free(unsafe.Pointer(cval))
		}
	}

	// If the key param was also requested as a fetch name, include it from the key's exported value
	cKeyVal := slice[0].value
	if cKeyVal != nil {
		exported := C.GoString(cKeyVal)
		C.gj_free(unsafe.Pointer(cKeyVal))
		for _, n := range names {
			if n == key.Name {
				values[n] = exported
				break
			}
		}
	}

	return int(res.jid), values, nil
}

// JailRemove removes a jail by JID.
func JailRemove(jid int) error {
	mu.Lock()
	res := C.gj_jail_remove(C.int(jid))
	mu.Unlock()

	if res.jid < 0 {
		return gjsyscall.MapErrno(syscall.Errno(res.errnum), C.GoString(&res.errmsg[0]))
	}
	return nil
}

// JailAttach attaches the calling process to a jail.
// WARNING: This changes the jail context of the entire process.
// Only use in forked child processes.
func JailAttach(jid int) error {
	mu.Lock()
	res := C.gj_jail_attach(C.int(jid))
	mu.Unlock()

	if res.jid < 0 {
		return gjsyscall.MapErrno(syscall.Errno(res.errnum), C.GoString(&res.errmsg[0]))
	}
	return nil
}

func buildCParams(params []Param) (*C.struct_gj_param, func()) {
	cparams := (*C.struct_gj_param)(C.calloc(C.size_t(len(params)), C.size_t(unsafe.Sizeof(C.struct_gj_param{}))))
	cstrings := make([]*C.char, 0, len(params)*2)

	slice := unsafe.Slice(cparams, len(params))
	for i, p := range params {
		cname := C.CString(p.Name)
		cstrings = append(cstrings, cname)
		slice[i].name = cname

		if p.Value != "" {
			cval := C.CString(p.Value)
			cstrings = append(cstrings, cval)
			slice[i].value = cval
		}

		if p.IsBool {
			slice[i].is_bool = 1
		}
	}

	cleanup := func() {
		for _, cs := range cstrings {
			C.free(unsafe.Pointer(cs))
		}
		C.free(unsafe.Pointer(cparams))
	}

	return cparams, cleanup
}
