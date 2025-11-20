#!/bin/bash
# deploy_with_owner.sh
# Usage: sudo ./deploy_with_owner.sh example.com client_username
# Follows the exact structure you provided, with robust fixes.

set -euo pipefail
IFS=$'\n\t'

DOMAIN="$1"
SYS_USER="${2:-}"   # client username (required)
BASE_PATH="/home/$SYS_USER/projuktisheba/bin"
PROJECT_PATH="$BASE_PATH/$DOMAIN"
NGINX_AVAILABLE="/etc/nginx/sites-available"
NGINX_ENABLED="/etc/nginx/sites-enabled"

if [ -z "$DOMAIN" ] || [ -z "$SYS_USER" ]; then
    echo "Usage: sudo ./deploy_with_owner.sh domain.com username"
    exit 1
fi

if [ ! -d "$PROJECT_PATH" ]; then
    echo "Error: Directory $PROJECT_PATH does not exist."
    exit 1
fi

echo "▶ Deploying $DOMAIN for linux user '$SYS_USER'"
echo "  Project path: $PROJECT_PATH"

# 1. Permission Fixes (Allow Nginx traversal)
# Add www-data to the client's group so nginx/php can traverse user home
if ! id -nG www-data | grep -qw "$SYS_USER"; then
    usermod -aG "$SYS_USER" www-data
    echo "  → Added www-data to group $SYS_USER"
fi

# Ensure the home and bin folders have group execute (traverse) bit
if [ -d "/home/$SYS_USER" ]; then
    chmod g+x "/home/$SYS_USER" || true
fi
if [ -d "/home/$SYS_USER/projuktisheba/bin" ]; then
    chmod g+x "/home/$SYS_USER/projuktisheba/bin" || true
fi

# Ensure project files are owned by the client (do not force root ownership)
chown -R "$SYS_USER:$SYS_USER" "$PROJECT_PATH"
echo "  → Project files owned by $SYS_USER:$SYS_USER"

# 2. Install Dependencies (Standard)
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
# install basic tools and allow adding PPA
apt-get install -y -qq software-properties-common apt-transport-https ca-certificates gnupg lsb-release

# Add Ondřej PHP PPA if not already present
if ! grep -q "^deb .\+ondrej/php" /etc/apt/sources.list /etc/apt/sources.list.d/* 2>/dev/null; then
    add-apt-repository -y ppa:ondrej/php
fi

apt-get update -qq
apt-get install -y -qq acl zip unzip

# 3. Detect PHP Version (robust)
# Prefer "major.minor" (e.g. 7.4, 8.1).
TARGET_PHP=""
if [ -f "$PROJECT_PATH/composer.json" ]; then
    RAW_VERSION=$(grep -Po '"php"\s*:\s*"[~^>=\s]*\K[0-9]+\.[0-9]+' "$PROJECT_PATH/composer.json" | head -n1 || true)
    if [ -n "$RAW_VERSION" ]; then
        MAJOR=$(echo "$RAW_VERSION" | cut -d. -f1)
        if [ "$MAJOR" -lt 7 ]; then
            TARGET_PHP="7.4"
        else
            TARGET_PHP="$RAW_VERSION"
        fi
    fi
fi

# Fallback
if [ -z "$TARGET_PHP" ]; then
    # if a php X.Y is installed, choose the latest one; else default to 7.4
    if [ -d /etc/php ]; then
        INSTALLED=$(ls /etc/php | sort -r | head -n1 || true)
        if [ -n "$INSTALLED" ]; then
            TARGET_PHP="$INSTALLED"
        else
            TARGET_PHP="7.4"
        fi
    else
        TARGET_PHP="7.4"
    fi
fi

# ensure format like "7.4"
TARGET_PHP=$(echo "$TARGET_PHP" | awk -F. '{print $1"."$2}')
echo "  → Using PHP version target: $TARGET_PHP"

# 3b. Install PHP and common extensions if missing
if [ ! -d "/etc/php/$TARGET_PHP/fpm" ]; then
    echo "  → Installing php$TARGET_PHP and common extensions..."
    apt-get update -qq
    apt-get install -y -qq "php${TARGET_PHP}-fpm" "php${TARGET_PHP}-cli" \
        "php${TARGET_PHP}-mysql" "php${TARGET_PHP}-xml" "php${TARGET_PHP}-mbstring" \
        "php${TARGET_PHP}-curl" "php${TARGET_PHP}-gd" "php${TARGET_PHP}-zip" "php${TARGET_PHP}-json" || {
        echo "Error: failed to install PHP $TARGET_PHP packages. Aborting."
        exit 1
    }
fi

# 4. CREATE CUSTOM FPM POOL (run PHP as the client user)
POOL_CONF="/etc/php/$TARGET_PHP/fpm/pool.d/$DOMAIN.conf"
SOCKET_PATH="/run/php/php${TARGET_PHP}-${DOMAIN}-fpm.sock"

cat > "$POOL_CONF" <<EOF
[$DOMAIN]
user = $SYS_USER
group = $SYS_USER
listen = $SOCKET_PATH
listen.owner = www-data
listen.group = www-data
pm = dynamic
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 5
chdir = $PROJECT_PATH
EOF

echo "  → PHP-FPM pool created: $POOL_CONF"

# Restart php-fpm to load the new pool safely
systemctl restart "php${TARGET_PHP}-fpm"
sleep 1

# 5. Ensure framework-specific writable folders exist and proper perms
cd "$PROJECT_PATH" || exit 1

# detect framework by simple checks
FRAMEWORK="php"
if [ -f "$PROJECT_PATH/artisan" ]; then
    FRAMEWORK="laravel"
elif [ -f "$PROJECT_PATH/wp-config.php" ]; then
    FRAMEWORK="wordpress"
elif [ -f "$PROJECT_PATH/index.php" ] && grep -q "CodeIgniter" "$PROJECT_PATH/index.php" 2>/dev/null; then
    FRAMEWORK="codeigniter"
fi

echo "  → Framework detected: $FRAMEWORK"

case "$FRAMEWORK" in
    laravel)
        mkdir -p "$PROJECT_PATH/storage" "$PROJECT_PATH/bootstrap/cache"
        chown -R "$SYS_USER:$SYS_USER" "$PROJECT_PATH/storage" "$PROJECT_PATH/bootstrap/cache"
        chmod -R 775 "$PROJECT_PATH/storage" "$PROJECT_PATH/bootstrap/cache"
        ;;
    codeigniter)
        mkdir -p "$PROJECT_PATH/application/logs" "$PROJECT_PATH/application/cache"
        chown -R "$SYS_USER:$SYS_USER" "$PROJECT_PATH/application/logs" "$PROJECT_PATH/application/cache"
        chmod -R 775 "$PROJECT_PATH/application/logs" "$PROJECT_PATH/application/cache"
        ;;
    wordpress)
        # wp-content must be writable by site owner
        mkdir -p "$PROJECT_PATH/wp-content"
        chown -R "$SYS_USER:$SYS_USER" "$PROJECT_PATH/wp-content"
        chmod -R 775 "$PROJECT_PATH/wp-content"
        ;;
    *)
        # generic php: nothing special
        ;;
esac

# Ensure the project is owned by client user and not root (so PHP-FPM can write)
chown -R "$SYS_USER:$SYS_USER" "$PROJECT_PATH"

# Lock down general permissions: files 644, dirs 755 (but preserve writable dirs)
find "$PROJECT_PATH" -type f -exec chmod 644 {} \; || true
find "$PROJECT_PATH" -type d -exec chmod 755 {} \; || true

echo "  → Ownership and base permissions applied"

# 6. Run Composer as the client user (if composer.json present)
COMPOSER_BIN=""
if command -v composer >/dev/null 2>&1; then
    COMPOSER_BIN="$(command -v composer)"
elif [ -x /usr/local/bin/composer ]; then
    COMPOSER_BIN="/usr/local/bin/composer"
fi

if [ -f "$PROJECT_PATH/composer.json" ]; then
    if [ -n "$COMPOSER_BIN" ]; then
        echo "  → Running composer install as $SYS_USER"
        # ensure composer uses the intended php binary
        PHP_BIN="/usr/bin/php${TARGET_PHP}"
        if [ ! -x "$PHP_BIN" ]; then PHP_BIN="php"; fi

        sudo -H -u "$SYS_USER" bash -c "export COMPOSER_HOME=\$HOME/.composer; PATH=\$PATH; $PHP_BIN $COMPOSER_BIN install --no-dev --optimize-autoloader --no-interaction" || {
            echo "Warning: composer install failed (check permissions or composer errors)."
        }
    else
        echo "  → Composer not found; skipping composer install."
    fi
fi

# 7. Nginx Config (point to custom socket)
WEB_ROOT="$PROJECT_PATH"
if [ -d "$PROJECT_PATH/public" ] && [ ! -f "$PROJECT_PATH/index.php" ]; then
    WEB_ROOT="$PROJECT_PATH/public"
fi

NGINX_CONF="$NGINX_AVAILABLE/$DOMAIN"
cat > "$NGINX_CONF" <<EOF
server {
    listen 80;
    server_name $DOMAIN www.$DOMAIN;

    root $WEB_ROOT;
    index index.php index.html;

    location / {
        try_files \$uri \$uri/ /index.php?\$query_string;
    }

    location ~ \.php\$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:$SOCKET_PATH;
        fastcgi_buffers 16 16k;
        fastcgi_buffer_size 32k;
    }

    location ~* \.(jpg|jpeg|png|gif|ico|css|js)\$ {
        expires 6M;
        add_header Cache-Control "public";
    }

    location ~ /\.ht {
        deny all;
    }
}
EOF

ln -sf "$NGINX_CONF" "$NGINX_ENABLED/$DOMAIN"
echo "  → Nginx config written: $NGINX_CONF"

# Test & reload nginx
if nginx -t; then
    systemctl reload nginx
    echo "  → Nginx reloaded"
else
    echo "Error: nginx config test failed. Check $NGINX_CONF"
    exit 1
fi

echo ""
echo "Done. PHP-FPM is running as '$SYS_USER' on socket '$SOCKET_PATH'"
echo "Visit: http://$DOMAIN"
