# Short URL

## Binding ip 변경 방법
```
$ redis-cli
127.0.0.1:6379> get conf:addr
127.0.0.1:6379> set conf:addr 0.0.0.0:8000

# Restarting server
```
