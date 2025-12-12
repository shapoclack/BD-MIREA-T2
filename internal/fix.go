package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ============ VIEW Functions ============

// CreateView creates a regular PostgreSQL VIEW
func CreateView(ctx context.Context, pool *pgxpool.Pool, viewName string, selectQuery string) error {
	if err := validateSQLIdent(viewName); err != nil {
		return err
	}
	if strings.TrimSpace(selectQuery) == "" {
		return fmt.Errorf("SELECT query cannot be empty")
	}

	query := fmt.Sprintf("CREATE VIEW %s AS %s", viewName, selectQuery)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Error creating view: %v", err)
		return fmt.Errorf("failed to create view: %w", err)
	}
	fmt.Printf("VIEW '%s' created successfully!\n", viewName)
	return nil
}

// CreateOrReplaceView creates or replaces a view
func CreateOrReplaceView(ctx context.Context, pool *pgxpool.Pool, viewName string, selectQuery string) error {
	if err := validateSQLIdent(viewName); err != nil {
		return err
	}
	if strings.TrimSpace(selectQuery) == "" {
		return fmt.Errorf("SELECT query cannot be empty")
	}

	query := fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", viewName, selectQuery)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Error creating or replacing view: %v", err)
		return fmt.Errorf("failed to create or replace view: %w", err)
	}
	fmt.Printf("VIEW '%s' created or updated successfully!\n", viewName)
	return nil
}

// DropView drops an existing view
func DropView(ctx context.Context, pool *pgxpool.Pool, viewName string) error {
	if err := validateSQLIdent(viewName); err != nil {
		return err
	}

	query := fmt.Sprintf("DROP VIEW IF EXISTS %s CASCADE", viewName)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Error dropping view: %v", err)
		return fmt.Errorf("failed to drop view: %w", err)
	}
	fmt.Printf("VIEW '%s' dropped successfully!\n", viewName)
	return nil
}

// GetViewDefinition retrieves the definition of a view
func GetViewDefinition(ctx context.Context, pool *pgxpool.Pool, viewName string) (string, error) {
	if err := validateSQLIdent(viewName); err != nil {
		return "", err
	}

	query := `
		SELECT definition 
		FROM pg_views 
		WHERE viewname = $1 AND schemaname = 'public'
	`
	var definition string
	err := pool.QueryRow(ctx, query, viewName).Scan(&definition)
	if err != nil {
		log.Printf("Error getting view definition: %v", err)
		return "", fmt.Errorf("failed to get view definition: %w", err)
	}
	return definition, nil
}

// ListAllViews returns all views in the public schema
func ListAllViews(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	query := `
		SELECT viewname 
		FROM pg_views 
		WHERE schemaname = 'public' 
		ORDER BY viewname
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list views: %w", err)
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var viewName string
		if err := rows.Scan(&viewName); err != nil {
			continue
		}
		views = append(views, viewName)
	}
	return views, nil
}

// ============ MATERIALIZED VIEW Functions ============

// CreateMaterializedView creates a materialized view (cached results)
func CreateMaterializedView(ctx context.Context, pool *pgxpool.Pool, mvName string, selectQuery string) error {
	if err := validateSQLIdent(mvName); err != nil {
		return err
	}
	if strings.TrimSpace(selectQuery) == "" {
		return fmt.Errorf("SELECT query cannot be empty")
	}

	query := fmt.Sprintf("CREATE MATERIALIZED VIEW %s AS %s", mvName, selectQuery)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Error creating materialized view: %v", err)
		return fmt.Errorf("failed to create materialized view: %w", err)
	}
	fmt.Printf("MATERIALIZED VIEW '%s' created successfully!\n", mvName)
	return nil
}

// RefreshMaterializedView refreshes the data in a materialized view
func RefreshMaterializedView(ctx context.Context, pool *pgxpool.Pool, mvName string, concurrently bool) error {
	if err := validateSQLIdent(mvName); err != nil {
		return err
	}

	concurrentlyStr := ""
	if concurrently {
		concurrentlyStr = "CONCURRENTLY"
	}

	query := fmt.Sprintf("REFRESH MATERIALIZED VIEW %s %s", concurrentlyStr, mvName)
	query = strings.TrimSpace(query)

	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Error refreshing materialized view: %v", err)
		return fmt.Errorf("failed to refresh materialized view: %w", err)
	}
	fmt.Printf("MATERIALIZED VIEW '%s' refreshed successfully!\n", mvName)
	return nil
}

// DropMaterializedView drops a materialized view
func DropMaterializedView(ctx context.Context, pool *pgxpool.Pool, mvName string) error {
	if err := validateSQLIdent(mvName); err != nil {
		return err
	}

	query := fmt.Sprintf("DROP MATERIALIZED VIEW IF EXISTS %s CASCADE", mvName)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		log.Printf("Error dropping materialized view: %v", err)
		return fmt.Errorf("failed to drop materialized view: %w", err)
	}
	fmt.Printf("MATERIALIZED VIEW '%s' dropped successfully!\n", mvName)
	return nil
}

// GetMaterializedViewDefinition retrieves the definition of a materialized view
func GetMaterializedViewDefinition(ctx context.Context, pool *pgxpool.Pool, mvName string) (string, error) {
	if err := validateSQLIdent(mvName); err != nil {
		return "", err
	}

	query := `
		SELECT definition 
		FROM pg_matviews 
		WHERE matviewname = $1 AND schemaname = 'public'
	`
	var definition string
	err := pool.QueryRow(ctx, query, mvName).Scan(&definition)
	if err != nil {
		log.Printf("Error getting materialized view definition: %v", err)
		return "", fmt.Errorf("failed to get materialized view definition: %w", err)
	}
	return definition, nil
}

// ListAllMaterializedViews returns all materialized views in the public schema
func ListAllMaterializedViews(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	query := `
		SELECT matviewname 
		FROM pg_matviews 
		WHERE schemaname = 'public' 
		ORDER BY matviewname
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list materialized views: %w", err)
	}
	defer rows.Close()

	var mvs []string
	for rows.Next() {
		var mvName string
		if err := rows.Scan(&mvName); err != nil {
			continue
		}
		mvs = append(mvs, mvName)
	}
	return mvs, nil
}
