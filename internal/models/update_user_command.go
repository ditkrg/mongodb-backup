package models

type UpdateUserCommand struct {
	UserName   string                   `bson:"updateUser"`
	CustomData map[string]interface{}   `bson:"customData"`
	Roles      []map[string]interface{} `bson:"roles"`
}
