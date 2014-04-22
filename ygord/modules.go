// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

var (
	// All the modules currently registered.
	modules = make([]Module, 0)
)

type Module interface {
	Init()
}

func RegisterModule(module Module) {
	module.Init()
	modules = append(modules, module)
}
