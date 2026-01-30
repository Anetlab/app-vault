# External PostgreSQL Configuration Guide

## Using External PostgreSQL Database

By default, the docker-compose setup now supports **both internal and external PostgreSQL**:

### Option 1: External PostgreSQL (Default)

Perfect for production or when you already have PostgreSQL running.

**1. Configure `.env` file:**
```bash
# Copy the example
cp .env.docker .env

# Edit .env with your database settings
DB_HOST=localhost                    # or your PostgreSQL host
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=appvault
DB_SSLMODE=disable                   # use 'require' for production
```

**2. Create the database:**
```sql
CREATE DATABASE appvault;
```

**3. Run migrations:**
```bash
# Apply database schema
psql -h localhost -U postgres -d appvault -f migrations/001_initial_schema.sql
```

**4. Start services (WITHOUT internal PostgreSQL):**
```bash
docker-compose up -d
```

This will start **only** the `appvault` backend and `web` frontend, connecting to your external database.

---

### Option 2: Internal PostgreSQL (Docker Compose)

For development or testing environments.

**1. Configure `.env` file:**
```bash
cp .env.docker .env

# Use these settings in .env
DB_HOST=postgres                     # Docker service name
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=appvault
DB_SSLMODE=disable
```

**2. Start services WITH internal PostgreSQL:**
```bash
docker-compose --profile with-postgres up -d
```

This will start **all services** including the PostgreSQL container.

---

## Quick Commands

### External DB (Default)
```bash
# Start without internal postgres
docker-compose up -d

# Check logs
docker-compose logs -f appvault

# Stop services
docker-compose down
```

### Internal DB (Docker)
```bash
# Start with internal postgres
docker-compose --profile with-postgres up -d

# Check postgres logs
docker-compose --profile with-postgres logs postgres

# Stop all including postgres
docker-compose --profile with-postgres down

# Remove postgres data
docker-compose --profile with-postgres down -v
```

---

## Development Mode (No Docker)

**Using external PostgreSQL:**

```bash
# Terminal 1 - Backend
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=appvault
export DB_SSLMODE=disable

go run cmd/server/main.go

# Terminal 2 - Frontend
cd web
npm run dev
```

---

## Connection String Format

The application uses this connection string format:
```
postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=SSLMODE
```

Example:
```
postgres://postgres:mypassword@localhost:5432/appvault?sslmode=disable
```

---

## Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DB_HOST` | PostgreSQL host | `postgres` | `localhost`, `db.example.com` |
| `DB_PORT` | PostgreSQL port | `5432` | `5432` |
| `DB_USER` | Database user | `postgres` | `appvault_user` |
| `DB_PASSWORD` | Database password | `postgres` | `secure_password_123` |
| `DB_NAME` | Database name | `appvault` | `appvault_prod` |
| `DB_SSLMODE` | SSL mode | `disable` | `require`, `verify-full` |

---

## Troubleshooting

### Backend can't connect to database

**1. Check if PostgreSQL is running:**
```bash
# For external DB
pg_isready -h localhost -p 5432

# For docker compose postgres
docker ps | grep postgres
```

**2. Check environment variables:**
```bash
docker exec appvault-server env | grep DB_
```

**3. Test connection manually:**
```bash
psql -h localhost -U postgres -d appvault
```

**4. Check backend logs:**
```bash
docker-compose logs appvault
```

### Common Errors

**Error: "connection refused"**
- Check if PostgreSQL is running
- Verify `DB_HOST` and `DB_PORT` are correct
- For external DB: use `localhost` or actual IP
- For docker: use service name `postgres`

**Error: "role does not exist"**
- Create the database user:
```sql
CREATE USER postgres WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE appvault TO postgres;
```

**Error: "database does not exist"**
- Create the database:
```sql
CREATE DATABASE appvault;
```

---

## Production Recommendations

For production deployments with external PostgreSQL:

1. **Use SSL/TLS**: Set `DB_SSLMODE=require` or `verify-full`
2. **Strong Password**: Use a secure, randomly generated password
3. **Dedicated User**: Create a dedicated database user (not `postgres`)
4. **Connection Pooling**: Configure max connections in PostgreSQL
5. **Backups**: Set up regular automated backups
6. **Monitoring**: Monitor connection pool usage and query performance

```bash
# Production .env example
DB_HOST=postgres.internal.example.com
DB_PORT=5432
DB_USER=appvault_prod
DB_PASSWORD=very_secure_random_password_here
DB_NAME=appvault_production
DB_SSLMODE=require
```

---

## Summary

**Quick Start:**
- External DB: `docker-compose up -d`
- Internal DB: `docker-compose --profile with-postgres up -d`

The new setup gives you **flexibility** - use what works best for your environment! 🚀
