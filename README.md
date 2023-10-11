# Short URL

## Docker를 사용한 배포방법
```
$ docker compose up -d

$ docker exec -it short-url-redis-1 redis-cli
127.0.0.1:6379> get conf:addr
127.0.0.1:6379> set conf:addr 0.0.0.0:8000

# Exit

$ docker compose restart app
```
