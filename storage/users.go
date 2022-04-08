package storage

type User struct {
	Id       int64
	Name     string
	Email    string
	Password string
}

func (c *Storage) UserAdd() {
}

func (c *Storage) UserConfirm() {
}

func (c *Storage) UserRemove() {
}

func (c *Storage) Get() (*User, error) {
	var result User
	return &result, nil
}
