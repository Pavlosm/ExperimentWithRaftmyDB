package cfg

import "testing"

func TestIsEqualIrEmpty(t *testing.T) {
	type test struct {
		id            NodeId
		testStr       string
		expectedEqual bool
		desc          string
	}

	tests := []test{
		{NodeId("AaB"), "AaB", true, "Same string"},
		{NodeId("Aa"), "aa", false, "Same characters, different capitalization"},
		{NodeId(""), "aa", true, "Empty node"},
		{NodeId("aa"), "", false, "Empty string"},
	}

	for _, v := range tests {
		e := v.id.IsEqualOrEmpty(v.testStr)
		if e != v.expectedEqual {
			t.Error(v.desc, "Expected:", v.expectedEqual, "Got:", e)
		}
	}
}
