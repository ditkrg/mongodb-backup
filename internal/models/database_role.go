package models

type Role struct {
	Role     string `bson:"role" json:"role"`
	Database string `bson:"db" json:"db"`
}
