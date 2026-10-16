package models

import (
	"github.com/goravel/framework/database/orm"
)

type Pipeline struct {
	orm.Model
	UserID           uint   `json:"-"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	SourceID         uint   `json:"-"`
	DestinationID    uint   `json:"-"`
	SourceSchema     string `json:"-"`
	SourceTable      string `json:"-"`
	DestinationTable string `json:"-"`
	ColumnsJSON      string `json:"-"`
	Frequency        string `json:"-"`
}
