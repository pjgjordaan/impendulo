package util

type Stringer interface {
	TypeName() string
	String() string
}
