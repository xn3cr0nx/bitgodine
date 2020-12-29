package db

// Repository interface represents relational db interface
// the interface should expose default methods inherited from gorm DB in the current phase
type Repository interface {
	Connect() error
	Model(value interface{}) (tx Repository)
	Find(dest interface{}, conds ...interface{}) (tx Repository)
	Where(query interface{}, args ...interface{}) (tx Repository)
	Create(value interface{}) (tx Repository)

	Error() error
}
