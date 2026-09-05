#!/bin/sh
set -e

# Target user and group
PUID=${PUID:-1000}
PGID=${PGID:-1000}

# Update lyostar user/group ID if custom PUID/PGID are provided
if [ "$PGID" != "1000" ]; then
    groupmod -o -g "$PGID" lyostar 2>/dev/null || sed -i -e "s/^lyostar:x:[0-9]*:/lyostar:x:$PGID:/" /etc/group
fi

if [ "$PUID" != "1000" ]; then
    usermod -o -u "$PUID" lyostar 2>/dev/null || sed -i -e "s/^lyostar:x:[0-9]*:[0-9]*:/lyostar:x:$PUID:$PGID:/" /etc/passwd
fi

# Ensure /data subdirectories exist and are owned by lyostar
mkdir -p /data/cache/covers /data/uploads
chown -R "$PUID:$PGID" /data

# Note: /books is STRICTLY READ-ONLY so we do not modify its permissions

# Drop root privileges and execute application as non-root user
exec su-exec "$PUID:$PGID" /app/lyostar "$@"
