# SQUABBLE BACK END

Personal project created to learn Golang. Squabble is a wordle party game. ( immitating https://squabble.me/ ) 

### To run:

- Create `.env` file containing:

```
# HTTP Section
PORT=8000

# Database Section
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_DATABASE_NAME=squabble
DB_SSLMODE=disable

# Redis Section
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=password
```

- `go run main.go`
