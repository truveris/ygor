// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// Module is the interface used by all the ygor modules.
type Module interface {
	Init(*Server)
}
