# Database Configuration Guide

App Vault uses PostgreSQL as its database backend. You can use either a local PostgreSQL instance or an external managed database service.

## Connection String Format

App Vault uses the standard PostgreSQL connection string format via the `DATABASE_URL` environment variable:

```
postgres://username:password@host:port/database?options
```

### Connection String Components

- `username` - PostgreSQL user
- `password` - User password
- `host` - Database server hostname or IP
- `port` - Database port (default: 5432)
- `database` - Database name
- `options` - Connection options (e.g., `sslmode=require`)

## Configuration Options

### Option 1: External Database (Recommended for Production)

Use an external PostgreSQL database service. Simply set the `DATABASE_URL` in your `.env` file:

```bash
DATABASE_URL=postgres://user:password@host:5432/appvault?sslmode=require
```

**Supported Providers:**

#### AWS RDS PostgreSQL
```bash
DATABASE_URL=postgres://username:password@mydb.c9akciq32.us-east-1.rds.amazonaws.com:5432/appvault?sslmode=require
```

#### Azure Database for PostgreSQL
```bash
DATABASE_URL=postgres://username@servername:password@servername.postgres.database.azure.com:5432/appvault?sslmode=require
```

#### Google Cloud SQL
```bash
DATABASE_URL=postgres://username:password@<CLOUD_SQL_CONNECTION_NAME>:5432/appvault?sslmode=require
```

#### DigitalOcean Managed Database
```bash
DATABASE_URL=postgres://doadmin:password@db-postgresql-nyc1-12345.db.ondigitalocean.com:25060/appvault?sslmode=require
```

#### Heroku Postgres
```bash
DATABASE_URL=postgres://user:pass@ec2-xxx.compute-1.amazonaws.com:5432/d12345?sslmode=require
```

#### Railway
```bash
DATABASE_URL=postgres://postgres:password@containers-us-west-12.railway.app:5432/railway?sslmode=require
```

#### Supabase
```bash
DATABASE_URL=postgres://postgres:password@db.project.supabase.co:5432/postgres?sslmode=require
```

#### Neon
```bash
DATABASE_URL=postgres://user:password@ep-cool-darkness-123456.us-east-2.aws.neon.tech/neondb?sslmode=require
```

### Option 2: Local PostgreSQL (Development)

For local development, you can run PostgreSQL directly:

#### Install PostgreSQL
**Windows:**
```powershell
choco install postgresql
```

**macOS:**
```bash
brew install postgresql
brew services start postgresql
```

**Linux:**
```bash
sudo apt-get install postgresql postgresql-contrib
sudo systemctl start postgresql
```

#### Connection String
```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/appvault?sslmode=disable
```

### Option 3: Docker Compose Local Database

Uncomment the postgres service in `docker-compose.yml`:

```yaml
services:
  postgres:
    image: postgres:15-alpine
    container_name: appvault-postgres
    environment:
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
      POSTGRES_DB: ${DB_NAME:-appvault}
    ports:
      - "${DB_PORT:-5432}:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

Then use:
```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/appvault?sslmode=disable
```

## SSL/TLS Configuration

### SSL Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `disable` | No SSL | Local development only |
| `require` | SSL required, no verification | Basic SSL |
| `verify-ca` | SSL + verify CA | Production |
| `verify-full` | SSL + verify CA + hostname | Maximum security |

### Production Recommendation
Always use `sslmode=require` or higher in production:

```bash
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
```

### Custom SSL Certificates

For custom CA certificates:

```bash
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=verify-ca&sslrootcert=/path/to/ca.crt
```

## Connection Pool Settings

App Vault uses connection pooling with these defaults:

- Max Open Connections: 25
- Max Idle Connections: 5
- Connection Max Lifetime: 5 minutes

These are configured in the application code and don't need to be set in the connection string.

## Database Requirements

### Minimum PostgreSQL Version
- PostgreSQL 13 or higher

### Required Extensions
App Vault uses standard PostgreSQL features and doesn't require special extensions.

### Required Permissions
The database user needs:
- `CREATE TABLE`
- `INSERT`, `UPDATE`, `DELETE`, `SELECT`
- `CREATE INDEX`
- `CREATE SEQUENCE`

### Database Initialization

App Vault automatically runs migrations on startup. Ensure:

1. Database exists
2. User has proper permissions
3. Network connectivity is available

## Testing Database Connection

### Using psql
```bash
psql "$DATABASE_URL"
```

### Using App Vault Health Check
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "database": "connected"
}
```

### Test Script
```powershell
# Windows PowerShell
$env:DATABASE_URL = "postgres://user:pass@host:5432/db?sslmode=require"
.\bin\appvault.exe

# Test health
curl http://localhost:8080/health
```

## Troubleshooting

### Connection Refused
```
could not connect to server: Connection refused
```

**Solutions:**
- Verify database is running
- Check host and port are correct
- Verify firewall rules allow connection
- For cloud databases, check security groups/firewall rules

### Authentication Failed
```
password authentication failed for user
```

**Solutions:**
- Verify username and password
- Check password special characters are URL-encoded
- For Azure, use `username@servername` format

### SSL Required
```
no pg_hba.conf entry for host, SSL off
```

**Solutions:**
- Change `sslmode=disable` to `sslmode=require`
- Ensure SSL is enabled on the database server

### Database Not Found
```
database "appvault" does not exist
```

**Solutions:**
```sql
-- Create database
CREATE DATABASE appvault;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE appvault TO username;
```

### Connection Timeout
```
could not connect to server: i/o timeout
```

**Solutions:**
- Check network connectivity
- Verify host is reachable
- Check if IP is whitelisted in cloud provider
- Increase connection timeout if needed

## URL Encoding Special Characters

If your password contains special characters, URL-encode them:

| Character | Encoded |
|-----------|---------|
| @ | %40 |
| : | %3A |
| / | %2F |
| ? | %3F |
| # | %23 |
| & | %26 |
| = | %3D |

Example:
```
Password: my@pass:word/123
Encoded:  my%40pass%3Aword%2F123
```

## Environment-Specific Configuration

### Development
```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/appvault?sslmode=disable
```

### Staging
```bash
DATABASE_URL=postgres://appvault_staging:securepass@staging-db.example.com:5432/appvault_staging?sslmode=require
```

### Production
```bash
DATABASE_URL=postgres://appvault_prod:verysecurepass@prod-db.example.com:5432/appvault_prod?sslmode=verify-full&sslrootcert=/etc/ssl/certs/ca.crt
```

## Security Best Practices

1. **Never commit DATABASE_URL** to version control
2. **Use SSL in production** (`sslmode=require` minimum)
3. **Strong passwords** (20+ characters, mixed case, numbers, symbols)
4. **Rotate credentials** regularly
5. **Least privilege** principle for database user
6. **Network isolation** (VPC, private subnets)
7. **Monitoring** for suspicious activity
8. **Backups** automated and tested
9. **Secrets management** (AWS Secrets Manager, Azure Key Vault, etc.)

## Migration from Local to External

1. **Backup local database:**
   ```bash
   pg_dump appvault > backup.sql
   ```

2. **Create external database** (via cloud provider)

3. **Restore backup:**
   ```bash
   psql "$NEW_DATABASE_URL" < backup.sql
   ```

4. **Update .env:**
   ```bash
   DATABASE_URL=<new_external_url>
   ```

5. **Restart application**

6. **Verify:**
   ```bash
   curl http://localhost:8080/health
   ```

## Related Documentation

- [DEPLOYMENT.md](DEPLOYMENT.md) - Production deployment guide
- [docker-compose.yml](docker-compose.yml) - Docker configuration
- [.env.example](.env.example) - Environment variables template
