// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "v8go.h"
import "C"
import (
	"runtime"
	"runtime/cgo"
	"unsafe"
)

// Context is a global root execution environment that allows separate,
// unrelated, JavaScript applications to run in a single instance of V8.
type Context struct {
	ptr        C.ContextPtr // Pointer to C++ V8GoContext object
	iso        *Isolate     // The Isolate this Context belongs to
	selfHandle cgo.Handle   // Opaque handle pointing to the Context itself
}

type contextOptions struct {
	iso   *Isolate
	gTmpl *ObjectTemplate
}

// ContextOption sets options such as Isolate and Global Template to the NewContext
type ContextOption interface {
	apply(*contextOptions)
}

// NewContext creates a new JavaScript context; if no Isolate is passed as a
// ContextOption than a new Isolate will be created.
func NewContext(opt ...ContextOption) *Context {
	opts := contextOptions{}
	for _, o := range opt {
		if o != nil {
			o.apply(&opts)
		}
	}

	if opts.iso == nil {
		opts.iso = NewIsolate()
	}

	if opts.gTmpl == nil {
		opts.gTmpl = &ObjectTemplate{&template{}}
	}

	ctx := &Context{
		iso: opts.iso,
	}
	ctx.selfHandle = cgo.NewHandle(ctx)
	ctx.ptr = C.NewContext(opts.iso.ptr, opts.gTmpl.ptr, C.uintptr_t(ctx.selfHandle))
	runtime.KeepAlive(opts.gTmpl)
	return ctx
}

func contextFromHandle(handle C.uintptr_t) *Context {
	return cgo.Handle(handle).Value().(*Context)
}

// Isolate gets the current context's parent isolate.
func (c *Context) Isolate() *Isolate {
	return c.iso
}

// RunScript executes the source JavaScript; origin (a.k.a. filename) provides a
// reference for the script and used in the stack trace if there is an error.
// error will be of type `JSError` if not nil.
func (c *Context) RunScript(source string, origin string) (*Value, error) {
	cSource := C.CString(source)
	cOrigin := C.CString(origin)
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cOrigin))

	rtn := C.RunScript(c.ptr, cSource, C.int(len(source)), cOrigin, C.int(len(origin)))
	return valueResult(c, rtn)
}

// Global returns the global proxy object.
// Global proxy object is a thin wrapper whose prototype points to actual
// context's global object with the properties like Object, etc. This is
// done that way for security reasons.
// Please note that changes to global proxy object prototype most probably
// would break the VM — V8 expects only global object as a prototype of
// global proxy object.
func (c *Context) Global() *Object {
	valPtr := C.ContextGlobal(c.ptr)
	v := &Value{valPtr, c}
	return &Object{v}
}

// PerformMicrotaskCheckpoint runs the default MicrotaskQueue until empty.
// This is used to make progress on Promises.
func (c *Context) PerformMicrotaskCheckpoint() {
	C.IsolatePerformMicrotaskCheckpoint(c.iso.ptr)
}

// Close will dispose the context and free the memory.
// You must call this yourself: the Go garbage collector will not free an unused open Context!
// Access to any values associated with the context after calling Close may panic.
func (c *Context) Close() {
	C.ContextFree(c.ptr)
	c.selfHandle.Delete()
	c.ptr = nil
}

func valueResult(ctx *Context, rtn C.RtnValue) (*Value, error) {
	if rtn.error.msg != nil {
		return nil, newJSError(rtn.error)
	}
	return &Value{rtn.value, ctx}, nil
}

func objectResult(ctx *Context, rtn C.RtnValue) (*Object, error) {
	if rtn.error.msg != nil {
		return nil, newJSError(rtn.error)
	}
	return &Object{&Value{rtn.value, ctx}}, nil
}

func (c *Context) pushValueScope() uint32 {
	return uint32(C.PushValueScope(c.ptr))
}

func (c *Context) popValueScope(scope uint32) {
	if C.PopValueScope(c.ptr, C.uint(scope)) == 0 {
		panic("Improper call to Context.PopValueScope: Scope is not current")
	}
}

// Calls the callback; any Values created in this Context during the callback will be
// invalidated when the callback returns and must not be referenced.
// This helps to reduce memory growth in a long-lived Context, since otherwise the Values
// would hold onto their JavaScript counterparts until the Context is closed.
func (c *Context) WithTemporaryValues(callback func()) {
	scope := c.pushValueScope()
	defer c.popValueScope(scope)
	callback()
}
