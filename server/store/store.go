package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-suggestions/server/bot"
	"github.com/mattermost/mattermost-server/v5/model"
)

const period = 1000 * 60 * 60 * 24 * 7 // week

// Store .
type Store struct {
	db  *sql.DB
	sq  sq.StatementBuilderType
	log bot.Logger
}

// NewStore .
func NewStore(driverName string, pluginAPI *pluginapi.Client, log bot.Logger) *Store {
	db, err := pluginAPI.Store.GetReplicaDB()
	if err != nil {
		log.Errorf("error while getting DB replica", err)
		return nil
	}
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if driverName == model.DATABASE_DRIVER_POSTGRES {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}
	builder = builder.RunWith(db)
	return &Store{
		db:  db,
		sq:  builder,
		log: log,
	}
}

// Close store
func (s *Store) Close() {
	if err := s.db.Close(); err != nil {
		s.log.Errorf("error while closing DB", err)
	}
}
