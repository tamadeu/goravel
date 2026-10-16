package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type Connector struct {
	orm.Model
	UserID               uint       `json:"-"`
	Purpose              string     `json:"-"`
	Name                 string     `json:"name"`
	Kind                 string     `json:"kind"`
	Status               string     `json:"status"`
	Host                 string     `json:"-"`
	Port                 *int       `json:"-"`
	DatabaseName         string     `json:"-"`
	Username             string     `json:"-"`
	PasswordEncrypted    string     `json:"-"`
	SchemaName           string     `json:"-"`
	SSLMode              string     `json:"-"`
	SSLRootCertEncrypted string     `json:"-"`
	SSLCertEncrypted     string     `json:"-"`
	SSLKeyEncrypted      string     `json:"-"`
	ConnectTimeout       int        `json:"-"`
	LastSyncAt           *time.Time `json:"-"`
}
