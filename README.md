### Danbooru RSS for discord web hook
a simple discord webhook to send new lewd pic from Danbooru

| environment  | description  |
|--------------|--------------|
|PROXY         | set proxy server(optional) |
|DISCORD       | Discord webhook URL |
|TAGS          | tags you want to add it (separated by comma) |
|MALE          | `Enable` or `disable` Male lewd pic(yaoi) by default is disabled |
|DUPLICATE     | `Enable` or `disable` duplicate image by default is disabled |
|DISABLETAGS   | blacklist for Danbooru tags (separated by comma) |  

#### Setup

###### Go
```
export DISCORD=export DISCORD=https://discordapp.com/api/webhooks/blablabla/blablabla
export TAGS=arknights
export MALE=disable
go build -o lewd 
export DISCORD=<Discord webhook> ./lewd
```

###### Docker
```
docker build . -t lewd
docker run -it -e DISCORD=<Discord webhook> -e TAGS=arknights -e MALE=disable lewd
```