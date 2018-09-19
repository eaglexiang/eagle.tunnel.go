package eagletunnel

import (
	"errors"
	"strconv"
	"strings"
)

const (
	Bigger = iota
	Smaller
	Equal
)

type Version struct {
	nodes []uint
	raw   string
}

func CreateVersion(src string) (Version, error) {
	result := Version{raw: src}
	items := strings.Split(result.raw, ".")
	result.nodes = make([]uint, len(items))
	for index, item := range items {
		item_int, err := strconv.ParseUint(item, 10, 32)
		if err != nil {
			return result, errors.New("invalid version string")
		}
		result.nodes[index] = uint(item_int)
	}
	return result, nil
}

func (src *Version) isBiggerThan(des *Version) bool {
	return src.compareWith(des) == Bigger
}

func (src *Version) isSmallerThan(des *Version) bool {
	return src.compareWith(des) == Smaller
}

func (src *Version) Equals(des *Version) bool {
	return src.compareWith(des) == Equal
}

func (src *Version) isBThanOrE2(des *Version) bool {
	relation := src.compareWith(des)
	return relation == Bigger || relation == Equal
}

func (src *Version) compareWith(des *Version) int {
	ind := 0
	for ; ind < len(src.nodes) && ind < len(des.nodes); ind++ {
		if src.nodes[ind] > des.nodes[ind] {
			return Bigger
		} else if src.nodes[ind] < des.nodes[ind] {
			return Smaller
		}
	}
	// src和des中的某一个更长，谁长谁大
	if len(src.nodes) > len(des.nodes) {
		return Bigger
	} else if len(src.nodes) < len(des.nodes) {
		return Smaller
	} else {
		return Equal
	}
}
