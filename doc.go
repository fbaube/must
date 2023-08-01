// Copyright 2022, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.
// Modifications 2023 Fred Baube.

// Package try (now: must) emulates aspects of the
// ill-fated "try" proposal using generics.
// See https://golang.org/issue/32437 for inspiration.
//
// Example usage:
//
//	func Fizz(...) (..., err error) {
//		defer try.HandleF(&err, func() {
//			if err == io.EOF {
//				err = io.ErrUnexpectedEOF
//			}
//		})
//		... := try.E2(Buzz(...))
//		return ..., nil
//	}
//
// This package is a sharp tool and should be used with care.
// Quick and easy error handling can occlude critical error
// handling logic. Panic handling generally should not cross
// package boundaries or be an explicit part of an API.
//
// Package try is a good fit for short Go programs and unit 
// tests where development speed is a greater priority than 
// reliability. Since the E functions panic if an error is 
// encountered, recovering in such programs is optional.
//
// Code before try:
//
//	func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) error {
//		switch t, err := d.ReadToken(); {
//		case err != nil:
//			return err
//		case t.Kind() != '[':
//			return fmt.Errorf("got %v, expecting array start", t.Kind())
//		}
//		if err := uo.UnmarshalNext(d, &a.Scalar); err != nil {
//			return err
//		}
//		if err := uo.UnmarshalNext(d, &a.Slice); err != nil {
//			return err
//		}
//		if err := uo.UnmarshalNext(d, &a.Map); err != nil {
//			return err
//		}
//		switch t, err := d.ReadToken(); {
//		case err != nil:
//			return err
//		case t.Kind() != ']':
//			return fmt.Errorf("got %v, expecting array end", t.Kind())
//		}
//		return nil
//	}
//
// Code after try:
//
//	func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) (err error) {
//		defer try.Handle(&err)
//		if t := try.E1(d.ReadToken()); t.Kind() != '[' {
//			return fmt.Errorf("found %v, expecting array start", t.Kind())
//		}
//		try.E(uo.UnmarshalNext(d, &a.Scalar))
//		try.E(uo.UnmarshalNext(d, &a.Slice))
//		try.E(uo.UnmarshalNext(d, &a.Map))
//		if t := try.E1(d.ReadToken()); t.Kind() != ']' {
//			return fmt.Errorf("found %v, expecting array end", t.Kind())
//		}
//		return nil
//	}
//
// Quick tour of the API
//
// The E family of functions all remove a final error return, panicking if non-nil.
//
// Handle recovers from that panic and allows assignment of the error
// to a return error value. Other panics are not recovered.
//
//	func f() (err error) {
//		defer try.Handle(&err)
//		...
//	}
//
// HandleF is like Handle, but it calls a function after any such assignment.
//
//	func f() (err error) {
//		defer try.HandleF(&err, func() {
//			if err == io.EOF {
//				err = io.ErrUnexpectedEOF
//			}
//		})
//		...
//	}
//
//	func foo(i int) (err error) {
//		defer try.HandleF(&err, func() {
//			err = fmt.Errorf("unable to foo %d: %w", i, err)
//		})
//		...
//	}
//
// F wraps an error with file and line information and calls a function 
// on error. It inter-operates well with testing.TB and log.Fatal.
//
//	func TestFoo(t *testing.T) {
//		defer try.F(t.Fatal)
//		...
//	}
//
//	func main() {
//		defer try.F(log.Fatal)
//		...
//	}
//
// Recover is like F, but it supports more complicated error handling
// by passing the error and runtime frame directly to a function.
//
//	func f() {
//		defer try.Recover(func(err error, frame runtime.Frame) {
//			// do something useful with err and frame
//		})
//		...
//	}
// .
package must

