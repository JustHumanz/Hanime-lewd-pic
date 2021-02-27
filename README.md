### Hanime RSS for discord web hook

#### Setup

###### Go
`go build -o lewd`  
`export DISCORD=<Discord webhook> ./lewd`  

###### Docker
`docker build . -t lewd`  
`docker run -it -e DISCORD=<Discord webhook> lewd`  