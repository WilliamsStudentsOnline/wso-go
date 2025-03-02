FROM redis
COPY redis.conf /usr/local/etc/redis/redis.conf
CMD [ "redis-server", "/usr/local/etc/redis/redis.conf", "--save", "300", "1", "--loglevel", "warning" ]

EXPOSE 6379