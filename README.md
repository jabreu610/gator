# Gator

Gator is a command-line RSS feed aggregator that allows you to manage and browse RSS feeds from your terminal.

## Prerequisites

Before running Gator, you'll need to have the following installed on your system:

- **Go** (version 1.25 or higher)
- **PostgreSQL**

## Installation

Install the Gator CLI using Go:

```bash
go install github.com/jabreu610/gator@latest
```

## Configuration

Before using Gator, you need to create a configuration file in your home directory:

1. Create a file named `.gatorconfig.json` in your home directory
2. Add your database connection string and initial user configuration:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace `username`, `password`, and the database name as needed for your PostgreSQL setup.

## Database Setup

You'll need to run the database migrations to set up the required tables. You can use a migration tool like `goose` to apply the migrations found in the `sql/schema` directory.

## Usage

Once installed and configured, you can use the following commands:

### User Management

- `gator register <username>` - Register a new user and log in
- `gator login <username>` - Log in as an existing user
- `gator users` - List all users (the current user is marked)
- `gator reset` - Clear all users from the database

### Feed Management

- `gator addfeed <name> <url>` - Add a new RSS feed (automatically follows it)
- `gator feeds` - List all feeds in the system
- `gator follow <url>` - Follow an existing feed
- `gator following` - List all feeds you're following
- `gator unfollow <url>` - Unfollow a feed

### Browsing Posts

- `gator browse [limit]` - View recent posts from feeds you follow (default limit: 2)
- `gator agg <duration>` - Start the aggregator to fetch new posts at the specified interval (e.g., `1m`, `30s`, `1h`)

## Example Workflow

```bash
# Register a new user
gator register alice

# Add and follow a feed
gator addfeed "TechCrunch" https://techcrunch.com/feed/

# Start aggregating feeds every 5 minutes
gator agg 5m

# In another terminal, browse posts
gator browse 10
```
