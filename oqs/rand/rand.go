// Package rand provides support for various RNG-related functions.
package rand // import "github.com/open-quantum-safe/liboqs-go/oqs/rand"

/**************** Callbacks ****************/

/*
#cgo pkg-config: liboqs
#include <oqs/oqs.h>
typedef void(*algorithm_ptr_fn)(uint8_t*, size_t);
void algorithmPtr_cgo(uint8_t*, size_t);
*/
import "C"

import (
	"errors"
	"unsafe"
)

// algorithmPtrCallback is a global RNG algorithm callback set by
// RandomBytesCustomAlgorithm.
var algorithmPtrCallback func([]byte, int)

// algorithmPtr is automatically invoked by RandomBytesCustomAlgorithm. When
// invoked, the memory is provided by the caller,
// i.e. RandomBytes or RandomBytesInPlace.
//export algorithmPtr
func algorithmPtr(randomArray *C.uint8_t, bytesToRead C.size_t) {
	if algorithmPtrCallback == nil {
		panic(errors.New("the RNG algorithm callback is not set"))
	}
	// TODO optimize-me!
	result := make([]byte, int(bytesToRead))
	algorithmPtrCallback(result, int(bytesToRead))
	p := unsafe.Pointer(randomArray)
	for _, v := range result {
		*(*C.uint8_t)(p) = C.uint8_t(v)
		p = unsafe.Pointer(uintptr(p) + 1)
	}
}

/**************** END Callbacks ****************/

/**************** Randomness ****************/

// RandomBytes generates bytesToRead random bytes. This implementation uses
// either the default RNG algorithm ("system"), or whichever algorithm has been
// selected by RandomBytesSwitchAlgorithm.
func RandomBytes(bytesToRead int) []byte {
	result := make([]byte, bytesToRead)
	C.OQS_randombytes((*C.uint8_t)(&result[0]), C.size_t(bytesToRead))
	return result
}

// RandomBytesInPlace generates bytesToRead random bytes. This implementation
// uses either the default RNG algorithm ("system"), or whichever algorithm has
// been selected by RandomBytesSwitchAlgorithm. bytesToRead must not exceed the
// size of randomArray.
func RandomBytesInPlace(randomArray []byte, bytesToRead int) {
	if bytesToRead > len(randomArray) {
		panic(errors.New("bytesToRead exceeds the size of randomArray"))
	}
	C.OQS_randombytes((*C.uint8_t)(&randomArray[0]), C.size_t(bytesToRead))
}

// RandomBytesSwitchAlgorithm switches the core OQS_randombytes to use the
// specified algorithm. Possible values are "system", "NIST-KAT", "OpenSSL".
// See <oqs/rand.h> liboqs header for more details.
func RandomBytesSwitchAlgorithm(algName string) {
	if C.OQS_randombytes_switch_algorithm(C.CString(algName)) != C.OQS_SUCCESS {
		panic(errors.New("can not switch algorithm"))
	}
}

// RandomBytesNistKatInit initializes the NIST DRBG with the entropyInput seed,
// which must be 48 exactly bytes long. The personalizationString is an optional
// personalization string, which, if non-empty, must be at least 48 bytes long.
func RandomBytesNistKatInit(entropyInput [48]byte,
	personalizationString []byte) {

	lenStr := len(personalizationString)
	if lenStr > 0 {
		if lenStr < 48 {
			panic(errors.New("the personalization string must be either empty" +
				" or at least 48 bytes long"))
		}

		C.OQS_randombytes_nist_kat_init((*C.uint8_t)(&entropyInput[0]),
			(*C.uint8_t)(&personalizationString[0]), 256)
		return
	}
	C.OQS_randombytes_nist_kat_init((*C.uint8_t)(&entropyInput[0]),
		(*C.uint8_t)(nil), 256)
}

// RandomBytesCustomAlgorithm switches RandomBytes to use the given function.
// This allows additional custom RNGs besides the provided ones. The provided
// RNG function must have the same signature as RandomBytesInPlace,
// i.e. func([]byte, int).
func RandomBytesCustomAlgorithm(fun func([]byte, int)) {
	algorithmPtrCallback = fun
	C.OQS_randombytes_custom_algorithm((C.algorithm_ptr_fn)(unsafe.Pointer(C.
		algorithmPtr_cgo)))
}

/**************** END Randomness ****************/