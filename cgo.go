// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// NOTE: These flags are now for building with Couchbase's build of V8.

//go:generate clang-format -i --verbose -style=Chromium v8go.h v8go.cc

// #cgo CXXFLAGS: -fno-rtti -fPIC -std=c++17 -DV8_32BIT_SMIS_ON_64BIT_ARCH -I${SRCDIR}/deps/include -Wall
// #cgo LDFLAGS: -pthread -lv8 -lv8_libplatform
// #cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/deps/darwin_x86_64
// #cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/deps/darwin_arm64
// #cgo linux,amd64 LDFLAGS: -L${SRCDIR}/deps/linux_x86_64 -ldl
// #cgo linux,arm64 LDFLAGS: -L${SRCDIR}/deps/linux_arm64 -ldl
import "C"
