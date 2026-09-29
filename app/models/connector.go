package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type Connector struct {
	orm.Model
	UserID       uint       `json:"-"`
	Name         string     `json:"name"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	Host         string     `json:"-"`
	Port         *int       `json:"-"`
	DatabaseName string     `json:"-"`
	LastSyncAt   *time.Time `json:"-"`
}
