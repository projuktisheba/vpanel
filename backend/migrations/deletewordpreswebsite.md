# Remove Wordpress Website From The VPS Server

## Delete Database and User 
- Delete database and corresponding users
## Delete nginx configuration file
- ``sudo rm -f /etc/nginx/sites-available/example.com.conf``
## Delete nginx symlink file
- ``sudo rm -f /etc/nginx/sites-enabled/example.com.conf``

## Remove SSL certificates(optional)
```
sudo rm -rf /etc/letsencrypt/live/example.com
sudo rm -rf /etc/letsencrypt/archive/example.com
sudo rm -f /etc/letsencrypt/renewal/example.com.conf
```

## Reload Nginx
```
sudo nginx -t
sudo systemctl reload nginx
```
