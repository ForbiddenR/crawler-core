package interfaces

type ModelListBinder interface {
	Bind() (l List, err error)
	Process(d any) (l List, err error)
}
