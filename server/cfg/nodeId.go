package cfg

type NodeId string

func (c NodeId) IsEqualOrEmpty(i string) bool {
	return c == "" || string(c) == i
}

func (c NodeId) IsEqual(i NodeId) bool {
	return string(c) == string(i)
}
