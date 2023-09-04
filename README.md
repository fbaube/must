# Try^H^H^H Must: Simpler Error Handling in Go

This module reduces the syntactic cost of error handling in Go.

[Documentation](https://pkg.go.dev/github.com/dsnet/try#section-documentation)

[API Quick Tour](https://pkg.go.dev/github.com/dsnet/try#hdr-Quick_tour_of_the_API)

The E family of functions all remove a final error return, panicking if non-nil.

Handle recovers from that panic and allows assignment of the error
to a return error value. Other panics are not recovered.

	func f() (err error) {
		defer try.Handle(&err)
		...
	}

HandleF is like Handle, but it calls a function after any such assignment.

	func f() (err error) {
		defer try.HandleF(&err, func() {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
		})
		...
	}

	func foo(i int) (err error) {
		defer try.HandleF(&err, func() {
			err = fmt.Errorf("unable to foo %d: %w", i, err)
		})
		...
	}

F wraps an error with file and line information and calls a function 
on error. It inter-operates well with testing.TB and log.Fatal.

	func TestFoo(t *testing.T) {
		defer try.F(t.Fatal)
		...
	}

	func main() {
		defer try.F(log.Fatal)
		...
	}

Recover is like F, but it supports more complicated error handling
by passing the error and runtime frame directly to a function.

	func f() {
		defer try.Recover(func(err error, frame runtime.Frame) {
			do something useful with err and frame
		})
		...
	}



<tt><b>================================================</b></tt>

Example usage in a main program:

```go
func main() {
    defer try.F(log.Fatal)
    b := try.E1(os.ReadFile(...))
    var v any
    try.E(json.Unmarshal(b, &v))
    ...
}
```

Example usage in a unit test:

```go
func Test(t *testing.T) {
    defer try.F(t.Fatal)
    db := try.E1(setdb.Open(...))
    defer db.Close()
    ...
    try.E(db.Commit())
}
```

Code before `try`:

```go
func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) error {
    switch t, err := d.ReadToken(); {
    case err != nil:
        return err
    case t.Kind() != '[':
        return fmt.Errorf("got %v, expecting array start", t.Kind())
    }
    if err := uo.UnmarshalNext(d, &a.Scalar); err != nil { return err }
    if err := uo.UnmarshalNext(d, &a.Slice);  err != nil { return err }
    if err := uo.UnmarshalNext(d, &a.Map);    err != nil { return err }

    switch t, err := d.ReadToken(); {
    case err != nil:
        return err
    case t.Kind() != ']':
        return fmt.Errorf("got %v, expecting array end", t.Kind())
    }
    return nil
}
```

Code after `try`:

```go
func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) (err error) {
    defer try.Handle(&err)
    if t := try.E1(d.ReadToken()); t.Kind() != '[' {
        return fmt.Errorf("found %v, expecting array start", t.Kind())
    }
    try.E(uo.UnmarshalNext(d, &a.Scalar))
    try.E(uo.UnmarshalNext(d, &a.Slice))
    try.E(uo.UnmarshalNext(d, &a.Map))
    if t := try.E1(d.ReadToken()); t.Kind() != ']' {
        return fmt.Errorf("found %v, expecting array end", t.Kind())
    }
    return nil
}
```

See the [documentation][godev] for more information.

[godev]: https://pkg.go.dev/github.com/dsnet/try
[actions]: https://github.com/dsnet/try/actions

## Install: <tt> go get -u github.com/dsnet/try </tt>

## License: BSD

