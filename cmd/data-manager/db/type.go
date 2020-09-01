package db

type database interface {
	Insert() error
	Update() error
	Query()
	Delete() error
}
