package proxyman_test

import (
	"reflect"
	"testing"

	. "github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/common"
)

func TestTagParser(t *testing.T) {

	datas := []data{
		{"", common.ErrNoClue},
		{"#", common.ErrNoClue},
		{"#0", 0},
		{"#000000", 0},
		{"##", "#"},
		{"###1#a##", "##1#a##"},
		{"#00001", 1},
		{"#-00001", "-00001"},
		{"#-3", "-3"},
		{"#+345", 345},
		{"#3.45", "3.45"},
		{"#----3", "----3"},
		{"#12345", 12345},
		{"#中文", "中文"},
		{"中文", "中文"},
		{"😀😁", "😀😁"},
		{"#00123abc", "00123abc"},
		{"#123😀😁中文abc", "123😀😁中文abc"},
	}

	for _, d := range datas {
		nos := ParseTag(d.tag)
		switch v := d.exp.(type) {
		case int, string, error:
			if v != nos {
				t.Fatalf("%s(%v) != %s(%v)", reflect.TypeOf(nos).String(), nos, reflect.TypeOf(v).String(), v)
			}
		default:
			t.Fatalf("unsupported test data: %v", d)
		}
	}
}

type data struct {
	tag string
	exp interface{}
}
