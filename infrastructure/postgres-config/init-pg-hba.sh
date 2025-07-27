#!/bin/bash
set -e

# Update pg_hba.conf to allow connections from Docker bridge network
echo "host all all 172.20.0.0/16 scram-sha-256" >> /var/lib/postgresql/data/pgdata/pg_hba.conf
echo "host all all 0.0.0.0/0 scram-sha-256" >> /var/lib/postgresql/data/pgdata/pg_hba.conf

# Reload PostgreSQL configuration
pg_ctl reload -D /var/lib/postgresql/data/pgdata