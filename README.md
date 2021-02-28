### Danbooru RSS for discord web hook
a simple disocrd webhook to send new lewd pic

#### Setup

###### Go
```
go build -o lewd 
export DISCORD=<Discord webhook> ./lewd
```

###### Docker
```
docker build . -t lewd
docker run -it -e DISCORD=<Discord webhook> lewd
```