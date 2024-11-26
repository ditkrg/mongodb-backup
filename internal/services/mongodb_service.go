package services

import (
	"context"
	"slices"

	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongodbService struct {
	client *mongo.Client
}

func NewMongodbService(mongodbConnectionString string, ctx context.Context) (*MongodbService, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbConnectionString))

	if err != nil {
		log.Error().Err(err).Msg("Error connecting to MongoDB")
		return nil, err
	}

	return &MongodbService{client: client}, nil
}

func (m *MongodbService) RemoveUserRoles(ctx context.Context, usersToSkip []string) error {
	log.Info().Msg("Removing roles from users")
	// ############################
	// Get all users
	// ############################
	collection := m.client.Database("admin").Collection("system.users")
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Error().Err(err).Msg("error listing users")
		return err
	}

	defer cursor.Close(ctx)

	// ############################
	// Iterate through all users
	// ############################
	for cursor.Next(ctx) {

		// ############################
		// Decode user
		// ############################
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			log.Error().Err(err).Msgf("Error decoding user %s", user.Id)
			return err
		}

		// ############################
		// Skip users that are in the usersToSkip list
		// ############################
		if slices.Contains(usersToSkip, user.Id) {
			log.Info().Msgf("skipping user %s because it was found in the skip list %v", user.Id, usersToSkip)
			continue
		}

		// ############################
		// Skip users that have no roles
		// ############################
		if len(user.Roles) == 0 {
			log.Info().Msgf("skipping user %s because it has no roles", user.Id)
			continue
		}

		// ############################
		// Prepare the updated user command
		// ############################
		command := user.PrepareRemoveUserRolesCommand()

		// ############################
		// Update the user
		// ############################
		if result := m.client.Database(user.Database).RunCommand(ctx, command); result.Err() != nil {
			log.Error().Err(result.Err()).Msgf("error updating user %s", user.Id)
			return result.Err()
		}

	}

	// ############################
	// Check for errors
	// ############################
	if err := cursor.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating through users")
		return err
	}

	log.Info().Msg("Roles removed from users")

	return nil
}

func (m *MongodbService) SetOriginalUserRoles(ctx context.Context, usersToSkip []string) error {
	log.Info().Msg("Setting backup original user roles")
	// ############################
	// Get all users
	// ############################
	collection := m.client.Database("admin").Collection("system.users")
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Error().Err(err).Msg("error listing users")
		return err
	}

	defer cursor.Close(ctx)

	// ############################
	// Iterate through all users
	// ############################
	for cursor.Next(ctx) {

		// ############################
		// Decode user
		// ############################
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			log.Error().Err(err).Msgf("Error decoding user %s", user.Id)
			return err
		}

		// ############################
		// Skip users that are in the usersToSkip list
		// ############################
		if slices.Contains(usersToSkip, user.Id) {
			log.Info().Msgf("skipping user %s because it was found in the skip list %v", user.Id, usersToSkip)
			continue
		}

		_, exists := user.CustomData[models.UserOriginalRoles]
		if !exists {
			log.Info().Msgf("skipping user %s because it has no original roles", user.Id)
			continue
		}

		// ############################
		// Prepare the updated user command
		// ############################
		command, err := user.PrepareSetOriginalUserRolesCommand()
		if err != nil {
			log.Error().Err(err).Msgf("error preparing update command for user %s", user.Id)
		}

		// ############################
		// Update the user
		// ############################
		if result := m.client.Database(user.Database).RunCommand(ctx, command); result.Err() != nil {
			log.Error().Err(result.Err()).Msgf("error updating user %s", user.Id)
			return result.Err()
		}

		log.Info().Msgf("Original roles set for user %s", user.Id)

	}

	// ############################
	// Check for errors
	// ############################
	if err := cursor.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating through users")
		return err
	}

	log.Info().Msg("Original user roles set successfully")

	return nil
}
