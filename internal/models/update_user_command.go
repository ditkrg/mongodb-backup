package models

type UpdateUserCommand struct {
	UserName   string                 `bson:"updateUser"`
	CustomData map[string]interface{} `bson:"customData"`
	Roles      []Role                 `bson:"roles"`
}
