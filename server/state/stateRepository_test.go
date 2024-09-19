package state

import (
	"fmt"
	"os"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	l := []Log{
		{
			Term:    1,
			Index:   1,
			Command: "command1",
		},
		{
			Term:    1,
			Index:   2,
			Command: "command1",
		},
	}

	f, err := os.OpenFile("logs.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		t.Errorf("Error opening file: %v", err)
	}

	defer f.Close()

	for _, v := range l {
		s := fmt.Sprintf("%v|%v|%v\n", v.Term, v.Index, v.Command)

		_, err = f.WriteString(s)
		if err != nil {
			t.Errorf("Error writing logs to file: %v", err)
		}
	}
}
