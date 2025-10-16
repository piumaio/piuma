# Nginx Reverse Proxy Example

```nginx
location /img/ {
  proxy_pass http://piuma:8080/image/;
  proxy_set_header Accept $http_accept;
  # Example rewrite: /img/<directive>/<path>
}
```
