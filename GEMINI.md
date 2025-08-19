# Project Overview: nht-bsihuyen.com-api

This project is a Go-based API for nht-bsihuyen.com.

## Key Technologies

*   **Language:** Go
*   **Database:** SQL (with migrations)
*   **Database Code Generation:** `sqlc` is used to generate Go code from SQL queries.
*   **API Framework:** The `api` directory suggests a custom-built API framework.
*   **Authentication:** JWT and Paseto tokens are used for authentication.

### Recent API Enhancements

*   **Authorization Middleware:**
    *   A new `requireAuthorization` middleware has been implemented to enforce role-based access control (RBAC). It checks if the authenticated user has a specific required role (e.g., "admin") before allowing access to certain API endpoints.
*   **Role Management APIs:**
    *   **Authorization:** Admin-only authorization is now enforced for `POST /roles`, `PUT /roles/:id`, and `DELETE /roles/:id` endpoints, leveraging the new `requireAuthorization` middleware and the `user_roles` join table.
    *   **New Endpoints:**
        *   `GET /roles`: List all roles with pagination.
        *   `GET /roles/:id`: Retrieve a specific role by ID.
        *   `PUT /roles/:id`: Update an existing role (name and description).
        *   `DELETE /roles/:id`: Delete a role.
    *   **API Response Improvement:** The `description` field in role-related API responses now returns a simple string (or `null` if not set) instead of the internal `sql.NullString` object.


## Project Structure

*   `.`: The root directory contains the main application (`main.go`), configuration (`app.env`), and dependency management files (`go.mod`, `go.sum`).
*   `api/`: This directory contains the API endpoints for managing accounts, transfers, and users. It also includes middleware for authentication and validation.
*   `db/`: This directory contains the database-related files, including:
    *   `migration/`: Database migration files.
    *   `query/`: SQL queries for `sqlc`.
    *   `sqlc/`: Generated Go code from `sqlc`.
*   `scripts/`: This directory contains shell scripts for deployment and other tasks.
*   `token/`: This directory contains the logic for creating and managing JWT and Paseto tokens.
*   `utils/`: This directory contains utility functions for configuration, password management, and other tasks.

## How to run the project

Based on the `Makefile`, you can use the following commands:

*   `make server`: To start the server
*   `make test`: To run the tests
*   `make sqlc`: To generate Go code from SQL queries
*   `make mock`: To generate mock interfaces for testing

Based on the `scripts` directory, you can use the following commands:

*   `./scripts/deploy.sh`: To deploy the project
*   `./scripts/reload.sh`: To reload the project
*   `./scripts/uninstall.sh`: To uninstall the project
