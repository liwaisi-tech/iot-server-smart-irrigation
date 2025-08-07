# Migration Guide: Traditional SQL Migrations to GORM Auto-Migrations

This guide provides step-by-step instructions for transitioning from traditional SQL migrations using `golang-migrate/migrate` to GORM auto-migrations.

## Overview

The migration includes:
- Transitioning from raw SQL queries to GORM ORM
- Preserving existing database schema and data
- Maintaining thread-safety and business logic validation
- Adding soft delete capability
- Implementing proper PostgreSQL triggers and indexes

## Pre-Migration Checklist

### 1. Backup Your Database
```bash
pg_dump -h localhost -U postgres -d iot_smart_irrigation > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 2. Verify Current Schema
```sql
-- Check existing devices table structure
\d devices

-- Count existing records
SELECT COUNT(*) FROM devices;

-- Verify indexes
\di devices*

-- Check constraints
\d+ devices
```

### 3. Test Environment Setup
- Ensure you have a test environment that mirrors production
- Test the migration process thoroughly before applying to production

## Migration Steps

### Step 1: Update Dependencies

The GORM dependencies have already been added to `go.mod`:
```go
gorm.io/gorm v1.30.1
gorm.io/driver/postgres v1.6.0
```

### Step 2: Updated Entity Model

The `Device` entity has been updated with GORM tags while preserving all existing validation logic:

**Key Changes:**
- Added GORM struct tags for database mapping
- Preserved thread-safety with `sync.RWMutex`
- Added audit fields (`CreatedAt`, `UpdatedAt`, `DeletedAt`)
- Implemented GORM hooks for validation (`BeforeCreate`, `BeforeUpdate`)
- Maintained all existing business logic methods

**GORM Features Used:**
- Primary key on `mac_address`
- Database constraints (NOT NULL, CHECK, size limits)
- Indexes on frequently queried fields
- Soft delete capability with `DeletedAt`

### Step 3: Database Connection Setup

Replace the traditional database setup with GORM:

```go
// Before (traditional)
db, err := database.NewPostgresDB(dbConfig)

// After (GORM)
gormDB, err := database.NewGormPostgresDB(dbConfig)
```

### Step 4: Run the Migration

Use the `MigrationManager` to handle the transition:

```go
// Initialize GORM database
gormDB, err := database.NewGormPostgresDB(dbConfig)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}

// Create migration manager
migrationManager := database.NewMigrationManager(gormDB)

// Run the migration
ctx := context.Background()
if err := migrationManager.MigrateFromTraditionalToGORM(ctx); err != nil {
    log.Fatal("Migration failed:", err)
}

// Validate data integrity
if err := migrationManager.ValidateDataIntegrity(ctx); err != nil {
    log.Fatal("Data integrity validation failed:", err)
}
```

### Step 5: Update Repository Implementation

Replace the raw SQL repository with the GORM repository:

```go
// Before (raw SQL)
deviceRepo := persistence.NewPostgresDeviceRepository(db)

// After (GORM)
deviceRepo := persistence.NewGormDeviceRepository(gormDB)
```

### Step 6: Update Main Application

Update your `main.go` or application initialization code:

```go
func initializeDatabase(cfg *config.DatabaseConfig) (*database.GormPostgresDB, error) {
    // Create GORM database connection
    gormDB, err := database.NewGormPostgresDB(cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create GORM database: %w", err)
    }

    // Run migrations
    migrationManager := database.NewMigrationManager(gormDB)
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := migrationManager.MigrateFromTraditionalToGORM(ctx); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }

    return gormDB, nil
}
```

## Migration Process Details

### Automatic Schema Detection

The `MigrationManager` automatically detects whether you're transitioning from traditional migrations:

1. **Checks for existing `devices` table**
2. **Detects traditional schema** (absence of `deleted_at` column)
3. **Creates backup table** with timestamp suffix
4. **Adds missing GORM columns** (`deleted_at` with index)
5. **Migrates data** to GORM-compatible format
6. **Runs GORM auto-migrations** to finalize schema
7. **Validates final schema** and data integrity

### Preserved Features

✅ **Database Constraints:**
- Primary key on `mac_address`
- NOT NULL constraints
- CHECK constraint for status values
- Size limits on varchar fields

✅ **Indexes:**
- `idx_devices_status`
- `idx_devices_last_seen`
- `idx_devices_registered_at`
- `idx_devices_deleted_at` (new)

✅ **PostgreSQL Triggers:**
- `update_updated_at_column()` function
- `update_devices_updated_at` trigger

✅ **Business Logic:**
- All validation methods preserved
- Thread-safety with mutex
- MAC address normalization
- Status update methods

### New Features

🆕 **Soft Delete:**
- Records marked as deleted instead of physically removed
- Use `Delete()` for soft delete, `HardDelete()` for permanent removal
- Soft-deleted records excluded from queries by default

🆕 **Enhanced Repository Methods:**
- `FindByStatus()` - Find devices by status
- `UpdateStatus()` - Update only status and last_seen
- `Count()` and `CountByStatus()` - Counting methods
- `Transaction()` - Transaction support

🆕 **GORM Hooks:**
- Automatic validation before create/update
- Data normalization on save
- Timestamp management

## Data Migration Strategy

### Backup Strategy
1. **Automatic Backup:** Migration creates timestamped backup tables
2. **Manual Backup:** Always create full database dump before migration
3. **Rollback Plan:** Backup tables can be used for rollback if needed

### Data Integrity Validation
The migration includes comprehensive validation:
- Verifies all required columns exist
- Checks all required indexes are present
- Validates data formats (MAC addresses, etc.)
- Ensures no records have missing required fields

### Zero-Downtime Considerations
For production deployments:
1. **Blue-Green Deployment:** Run migration in separate environment
2. **Maintenance Window:** Schedule brief downtime for schema changes
3. **Gradual Rollout:** Test with subset of data first

## Rollback Strategy

### Immediate Rollback (if migration fails)
```go
// The migration is transactional - if it fails, changes are rolled back automatically
// Check backup tables created during migration:
SELECT tablename FROM pg_tables WHERE tablename LIKE 'devices_backup_%';
```

### Manual Rollback (after successful migration)
```sql
-- 1. Find the backup table
SELECT tablename FROM pg_tables WHERE tablename LIKE 'devices_backup_%' ORDER BY tablename DESC LIMIT 1;

-- 2. Restore from backup (replace YYYYMMDD_HHMMSS with actual timestamp)
DROP TABLE devices;
ALTER TABLE devices_backup_YYYYMMDD_HHMMSS RENAME TO devices;

-- 3. Recreate indexes if needed
CREATE INDEX idx_devices_status ON devices(status);
-- ... other indexes
```

## Testing the Migration

### Unit Tests
```go
func TestDeviceGORMValidation(t *testing.T) {
    device := &entities.Device{
        MACAddress:          "AA:BB:CC:DD:EE:FF",
        DeviceName:          "Test Device",
        IPAddress:           "192.168.1.100",
        LocationDescription: "Test Location",
        Status:              "registered",
    }
    
    err := device.Validate()
    assert.NoError(t, err)
}
```

### Integration Tests
```go
func TestGormRepositoryIntegration(t *testing.T) {
    // Test with actual database
    gormDB := setupTestGormDB(t)
    repo := persistence.NewGormDeviceRepository(gormDB)
    
    device, err := entities.NewDevice("AA:BB:CC:DD:EE:FF", "Test", "192.168.1.100", "Location")
    require.NoError(t, err)
    
    // Test save
    err = repo.Save(context.Background(), device)
    assert.NoError(t, err)
    
    // Test find
    found, err := repo.FindByMACAddress(context.Background(), device.MACAddress)
    assert.NoError(t, err)
    assert.Equal(t, device.MACAddress, found.MACAddress)
}
```

## Performance Considerations

### GORM Performance Best Practices
1. **Connection Pooling:** Already configured in `GormPostgresDB`
2. **Prepared Statements:** GORM uses prepared statements automatically
3. **Batch Operations:** Use `CreateInBatches()` for bulk inserts
4. **Eager Loading:** Use `Preload()` for related data (if added later)
5. **Raw Queries:** Use `Raw()` for complex queries when needed

### Monitoring
- Monitor query performance after migration
- Check connection pool utilization
- Validate index usage with `EXPLAIN ANALYZE`

## Troubleshooting

### Common Issues

**Issue:** Migration fails with constraint violation
```
Solution: Check existing data for constraint violations before migration
SQL: SELECT * FROM devices WHERE status NOT IN ('registered', 'online', 'offline');
```

**Issue:** GORM can't find table
```
Solution: Ensure TableName() method is correctly implemented
Code: func (Device) TableName() string { return "devices" }
```

**Issue:** Validation errors after migration
```
Solution: Check GORM hooks are properly implemented and validation logic is preserved
```

### Debug Mode
Enable GORM debug mode for troubleshooting:
```go
gormConfig := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // Set to logger.Silent for production
}
```

## Post-Migration Tasks

### 1. Cleanup Backup Tables
```go
// Clean up backup tables older than 30 days
migrationManager.CleanupBackupTables(ctx, 30*24*time.Hour)
```

### 2. Update Monitoring
- Update database monitoring queries to account for soft deletes
- Adjust alerting thresholds if needed

### 3. Documentation Updates
- Update API documentation to reflect soft delete behavior
- Update operational runbooks

### 4. Performance Validation
- Run performance tests to ensure no regression
- Monitor query patterns and optimize if needed

## Support

If you encounter issues during migration:
1. Check the backup tables are created properly
2. Validate data integrity using the provided validation methods
3. Review GORM logs for detailed error information
4. Test rollback procedure in staging environment first

The migration is designed to be safe and transactional, but always test thoroughly in a non-production environment first.