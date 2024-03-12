// This file is part of mirrorlist.
// Copyright (C) 2024 Enindu Alahapperuma
//
// mirrorlist is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// mirrorlist is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with
// mirrorlist. If not, see <https://www.gnu.org/licenses/>.

package main

import "testing"

func BenchmarkEninduMirrorlist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		main()
	}
}