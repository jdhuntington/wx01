package db

import "github.com/jackc/pgx/v5/pgxpool"

// Pool is a type alias so other packages don't need to import pgx directly.
type Pool = *pgxpool.Pool
