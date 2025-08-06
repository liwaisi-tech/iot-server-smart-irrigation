-- Drop the trigger first
DROP TRIGGER IF EXISTS update_devices_updated_at ON devices;

-- Drop the function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop the devices table (this will also drop all indexes)
DROP TABLE IF EXISTS devices;