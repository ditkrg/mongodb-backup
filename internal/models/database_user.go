package models

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

const UserOriginalRoles = "userOriginalRoles"

type User struct {
	Id         string                 `bson:"_id"`
	Database   string                 `bson:"db"`
	UserName   string                 `bson:"user"`
	Roles      []Role                 `bson:"roles"`
	CustomData map[string]interface{} `bson:"customData,omitempty"`
}

func (user *User) PrepareRemoveUserRolesCommand() *UpdateUserCommand {
	// ############################
	// Store the original roles
	// ############################
	if user.CustomData == nil {
		user.CustomData = make(map[string]interface{})
	}

	user.CustomData[UserOriginalRoles] = user.Roles

	// ############################
	// Prepare the update command
	// ############################
	return &UpdateUserCommand{
		UserName:   user.UserName,
		Roles:      make([]Role, 0),
		CustomData: user.CustomData,
	}
}

func (user *User) PrepareSetOriginalUserRolesCommand() (*UpdateUserCommand, error) {
	// ############################
	// Get the original roles
	// ############################
	rolesBson := user.CustomData[UserOriginalRoles]

	// ############################
	// From the original roles from the custom data
	// ############################
	delete(user.CustomData, UserOriginalRoles)

	jsonData, err := json.Marshal(rolesBson)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling bson array")
		return nil, err
	}

	var roles []Role
	err = json.Unmarshal(jsonData, &roles)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshalling json")
		return nil, err
	}
	// ############################
	// Prepare the update command
	// ############################
	return &UpdateUserCommand{
		UserName:   user.UserName,
		Roles:      roles,
		CustomData: user.CustomData,
	}, nil
}
