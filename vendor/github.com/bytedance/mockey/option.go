/*
 * Copyright 2022 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mockey

type option struct {
	unsafe  bool
	generic *bool
	method  *bool
}

type optionFn func(*option)

func OptUnsafe(o *option) {
	o.unsafe = true
}

func OptGeneric(o *option) {
	var t = true
	o.generic = &t
}

func OptMethod(o *option) {
	var t = true
	o.method = &t
}

func resolveOpt(target interface{}, fn ...optionFn) *option {
	opt := &option{
		unsafe:  false,
		generic: nil,
		method:  nil,
	}
	for _, f := range fn {
		f(opt)
	}
	return opt
}
