package interfaces

type FilterCondition interface {
	GetKey() (key string)
	SetKey(key string)
	GetOp() (op string)
	SetOp(op string)
	GetValue() (value any)
	SetValue(value any)
}
