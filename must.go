// Copyright 2022, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.
// Modifications 2023 Fred Baube.

package must

import (
	"runtime"
	"strconv"
)

// wrapdError wraps an error to ensure that we only
// recover from errors panicked by this package.
type wrapdError struct {
	error
	pc [1]uintptr
}

func (e wrapdError) Error() string {
	// Retrieve the last path segment of the filename.
	// We avoid using strings.LastIndexByte to keep dependencies small.
	frames := runtime.CallersFrames(e.pc[:])
	frame, _ := frames.Next()
	file := frame.File
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			file = file[i+len("/"):]
			break
		}
	}
	return file + ":" + strconv.Itoa(frame.Line) + ": " + e.error.Error()
}

// Unwrap primarily exists for testing purposes.
func (e wrapdError) Unwrap() error {
	return e.error
}

func r(recovered any, fn func(wrapdError)) {
	switch ex := recovered.(type) {
	case nil:
	case wrapdError:
		fn(ex)
	default:
		panic(ex)
	}
}

// Recover recovers an error previously panicked with an E function.
// If it recovers an error, it calls fn with the error and the runtime
// frame in which it occurred.
func Recover(fn func(err error, frame runtime.Frame)) {
	r(recover(), func(w wrapdError) {
		frames := runtime.CallersFrames(w.pc[:])
		frame, _ := frames.Next()
		fn(w.error, frame)
	})
}

// Handle recovers an error previously panicked
// with an E function and stores it into errptr.
func Handle(errptr *error) {
	r(recover(), func(w wrapdError) { *errptr = w.error })
}

// HandleF recovers an error previously panicked
// with an E function and stores it into errptr.
// If it recovers an error, it calls fn.
func HandleF(errptr *error, fn func()) {
	r(recover(), func(w wrapdError) {
		*errptr = w.error
		if w.error != nil {
			fn()
		}
	})
}

// F recovers an error previously panicked with an E function,
// wraps it, and passes it to fn. The wrapping includes the 
// file and line of the runtime frame in which it occurred.
// F pairs well with testing.TB.Fatal and log.Fatal.
func F(fn func(...any)) {
	r(recover(), func(w wrapdError) { f(fn, w) })
}

// e panics. 
func e(err error) {
	we := wrapdError{error: err}
	// 3: runtime.Callers, e, E
	runtime.Callers(3, we.pc[:])
	panic(we)
}

// E panics if err is non-nil.
func E(err error) {
	if err != nil {
		e(err)
	}
}

// E1 returns a as-is.
// It panics if err is non-nil.
func E1[A any](a A, err error) A {
	if err != nil {
		e(err)
	}
	return a
}

// E2 returns a and b as-is.
// It panics if err is non-nil.
func E2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		e(err)
	}
	return a, b
}

// E3 returns a, b, and c as-is.
// It panics if err is non-nil.
func E3[A, B, C any](a A, b B, c C, err error) (A, B, C) {
	if err != nil {
		e(err)
	}
	return a, b, c
}

// E4 returns a, b, c, and d as is.
// It panics if err is non-nil.
func E4[A, B, C, D any](a A, b B, c C, d D, err error) (A, B, C, D) {
	if err != nil {
		e(err)
	}
	return a, b, c, d
}

// f simply calls fn with w.
//
// This uses the special "line" pragma to set the file 
// and line number to be something consistent. It must 
// be declared last in the file to prevent "line" from 
// affecting the line numbers of anything else in this file.
// . 
func f(fn func(...any), w wrapdError) {
//line try.go:1
	fn(w)
}
