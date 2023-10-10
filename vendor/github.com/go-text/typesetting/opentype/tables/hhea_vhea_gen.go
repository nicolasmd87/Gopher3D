// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"encoding/binary"
	"fmt"
)

// Code generated by binarygen from hhea_vhea_src.go. DO NOT EDIT

func (item *Hhea) mustParse(src []byte) {
	_ = src[35] // early bound checking
	item.majorVersion = binary.BigEndian.Uint16(src[0:])
	item.minorVersion = binary.BigEndian.Uint16(src[2:])
	item.Ascender = int16(binary.BigEndian.Uint16(src[4:]))
	item.Descender = int16(binary.BigEndian.Uint16(src[6:]))
	item.LineGap = int16(binary.BigEndian.Uint16(src[8:]))
	item.AdvanceMax = binary.BigEndian.Uint16(src[10:])
	item.MinFirstSideBearing = int16(binary.BigEndian.Uint16(src[12:]))
	item.MinSecondSideBearing = int16(binary.BigEndian.Uint16(src[14:]))
	item.MaxExtent = int16(binary.BigEndian.Uint16(src[16:]))
	item.CaretSlopeRise = int16(binary.BigEndian.Uint16(src[18:]))
	item.CaretSlopeRun = int16(binary.BigEndian.Uint16(src[20:]))
	item.CaretOffset = int16(binary.BigEndian.Uint16(src[22:]))
	item.reserved[0] = binary.BigEndian.Uint16(src[24:])
	item.reserved[1] = binary.BigEndian.Uint16(src[26:])
	item.reserved[2] = binary.BigEndian.Uint16(src[28:])
	item.reserved[3] = binary.BigEndian.Uint16(src[30:])
	item.metricDataformat = int16(binary.BigEndian.Uint16(src[32:]))
	item.NumOfLongMetrics = binary.BigEndian.Uint16(src[34:])
}

func ParseHhea(src []byte) (Hhea, int, error) {
	var item Hhea
	n := 0
	if L := len(src); L < 36 {
		return item, 0, fmt.Errorf("reading Hhea: "+"EOF: expected length: 36, got %d", L)
	}
	item.mustParse(src)
	n += 36
	return item, n, nil
}