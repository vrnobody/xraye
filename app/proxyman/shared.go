package proxyman

import (
	"strconv"

	"github.com/xtls/xray-core/common"
)

func ParseTag(tag string) interface{} {
	if tag == "" || (len(tag) == 1 && tag[0] == '#') {
		return common.ErrNoClue
	}
	if tag[0] != '#' {
		return tag
	}
	t := tag[1:]
	if i, err := strconv.Atoi(t); err == nil && i >= 0 {
		return i
	}
	return t
}
