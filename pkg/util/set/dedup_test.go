package set

import (
	"fmt"
	"testing"
)

func TestSortDeDup(t *testing.T) {
	{
		l := []string{"c", "a", "b", "c", "e", "d"}
		expected := []string{"a", "b", "c", "d", "e"}
		sl := SortDeDup(l)
		if fmt.Sprint(sl) != fmt.Sprint(expected) {
			t.Errorf("%v should be equal %v", sl, expected)
		}
	}
	{
		l := []string{"mj.mc_trscode", "mj.rowid", "mj.tans_amt", "mj.trans_bran_code", "mj.trans_date", "mj.trans_flag", "trans_date", "mj.trans_flag", "trans_date"}
		expected := []string{"mj.mc_trscode", "mj.rowid", "mj.tans_amt", "mj.trans_bran_code", "mj.trans_date", "mj.trans_flag", "trans_date"}
		sl := SortDeDup(l)
		if fmt.Sprint(sl) != fmt.Sprint(expected) {
			t.Errorf("%v should be equal %v", sl, expected)
		}
	}
	{
		l := []string{"c"}
		expected := []string{"c"}
		sl := SortDeDup(l)
		if fmt.Sprint(sl) != fmt.Sprint(expected) {
			t.Errorf("%v should be equal %v", sl, expected)
		}
	}
	{
		l := make([]string, 0)
		expected := []string{}
		sl := SortDeDup(l)
		if fmt.Sprint(sl) != fmt.Sprint(expected) {
			t.Errorf("%v should be equal %v", sl, expected)
		}
	}
	{
		var l []string
		expected := []string(nil)
		sl := SortDeDup(l)
		if fmt.Sprint(sl) != fmt.Sprint(expected) {
			t.Errorf("%v should be equal %v", sl, expected)
		}
	}
}
