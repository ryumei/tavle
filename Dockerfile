FROM busybox
ADD ./dist/linux-amd64/ /app
ADD ./contrib/docker/log/ /app/log/
ADD ./contrib/docker/chatlog/ /app/chatlog/
ADD ./contrib/docker/conf/ /app/conf/
WORKDIR /app
EXPOSE 8000
ENTRYPOINT ["/app/tavle", "-c", "/app/conf/tavle.tml"]
