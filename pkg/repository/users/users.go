package users

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/i7tsov/example-microservice/pkg/model"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// Repository implements component that access users database.
type Repository struct {
	c  Config
	db *bun.DB
}

// Config contains configuration options for Repository.
type Config struct {
	Host     string `yaml:"host" validate:"required"`
	Port     string `yaml:"port" validate:"required"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Name     string `yaml:"name" validate:"required"`
}

// New creates new users repository.
func New(c Config) (*Repository, error) {
	// Validate config & dependencies.
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		// Error contains sufficient information about the cause, no wrapping needed.
		return nil, err
	}

	r := &Repository{
		c: c,
	}
	dsn := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.Name)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	r.db = bun.NewDB(sqldb, pgdialect.New())
	r.db.AddQueryHook(bundebug.NewQueryHook())

	err = r.db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	err = r.setupTables()
	if err != nil {
		return nil, fmt.Errorf("database setup tables failed: %w", err)
	}

	// TODO: add proper migrations.

	return r, nil
}

// Create tables that belong to service if they're not exist.
func (r *Repository) setupTables() error {
	_, err := r.db.NewCreateTable().Model((*model.User)(nil)).IfNotExists().Exec(context.Background())
	return err
}

func (r *Repository) AddUser(ctx context.Context, user model.User) (string, error) {
	_, err := r.db.NewInsert().Model(&user).Returning("id").Exec(ctx)
	return user.ID, err
}

func (r *Repository) RemoveUser(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Where("id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) ListUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.NewSelect().Model(&users).Scan(ctx)
	return users, err
}
