# Project Overview: nht-bsihuyen.com-api

This project is a Go-based API for nht-bsihuyen.com.

## Key Technologies

*   **Language:** Go
*   **Database:** SQL (with migrations)
*   **Database Code Generation:** `sqlc` is used to generate Go code from SQL queries.
*   **API Framework:** The `api` directory suggests a custom-built API framework.
*   **Authentication:** JWT and Paseto tokens are used for authentication.

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
*   `make build`: To build the project

Based on the `scripts` directory, you can use the following commands:

*   `./scripts/deploy.sh`: To deploy the project
*   `./scripts/reload.sh`: To reload the project
*   `./scripts/uninstall.sh`: To uninstall the project
