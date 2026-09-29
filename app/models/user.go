package models

import (
	"github.com/goravel/framework/database/orm"
)

type User struct {
	orm.Model
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}
