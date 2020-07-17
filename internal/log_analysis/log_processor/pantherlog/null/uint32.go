// nolint: dupl
package null

/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"strconv"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/jsonutil"
)

type Uint32 struct {
	Value  uint32
	Exists bool
}

// FromUint32 creates a non-null Uint32.
// It is inlined by the compiler.
func FromUint32(n uint32) Uint32 {
	return Uint32{
		Value:  n,
		Exists: true,
	}
}

func (u *Uint32) UnmarshalJSON(data []byte) error {
	if string(data) == `null` {
		*u = Uint32{}
		return nil
	}
	data = jsonutil.UnquoteJSON(data)
	if len(data) == 0 {
		*u = Uint32{}
		return nil
	}
	n, err := strconv.ParseUint(string(data), 10, 32)
	if err != nil {
		return err
	}
	*u = Uint32{
		Value:  uint32(n),
		Exists: true,
	}
	return nil
}

// int64Codec is a jsoniter encoder/decoder for integer values
type uint32Codec struct{}

// Decode implements jsoniter.ValDecoder interface
func (*uint32Codec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	const opName = "ReadNullUint32"
	switch iter.WhatIsNext() {
	case jsoniter.NilValue:
		iter.ReadNil()
		*((*Uint32)(ptr)) = Uint32{}
	case jsoniter.StringValue:
		s := iter.ReadStringAsSlice()
		if len(s) == 0 {
			*((*Uint32)(ptr)) = Uint32{}
			return
		}
		n, err := strconv.ParseUint(string(s), 10, 32)
		if err != nil {
			iter.ReportError(opName, err.Error())
			return
		}
		*((*Uint32)(ptr)) = Uint32{
			Value:  uint32(n),
			Exists: true,
		}
	default:
		iter.Skip()
		iter.ReportError(opName, "invalid null uint32 value")
	}
}

// Encode implements jsoniter.ValEncoder interface
func (*uint32Codec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if u := (*Uint32)(ptr); u.Exists {
		stream.WriteUint32(u.Value)
	} else {
		stream.WriteNil()
	}
}

// IsEmpty implements jsoniter.ValEncoder interface
// WARNING: This considers `null` values as empty and omits them
func (*uint32Codec) IsEmpty(ptr unsafe.Pointer) bool {
	return !((*Uint32)(ptr)).Exists
}
