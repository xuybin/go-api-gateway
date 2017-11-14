# go-api-gateway

[![Build Status](https://travis-ci.org/xuybin/go-api-gateway.svg?branch=master)](https://travis-ci.org/xuybin/go-api-gateway)
A simple API gateway written by golang.

Support for authenticate and authorization, and web applications will be protected after the gateway.

in development now.

documents will be wrote later.

## ARCH

![arch](https://res.cloudinary.com/digf90pwi/image/upload/v1502367434/go-simple-api-gatway_1_grgl5o.png)

## CONFIGURATION

You could use **cli option** or **environment varibles** to config your api gateway

```bash
./go-api-gateway --help
Options:

  -h, --help                                display help information
  -c, --*conn[=$GATEWAY_CONN_STR]          *mysql connection str
  -l, --*listen[=$GATEWAY_LS]              *gateway listen host and port
  -r, --*resource[=$GATEWAY_RESOURCE_URL]  *gateway resource url

```

* -c --conn **GATEWAY_CONN_STR**, mysql connection string, format is *user:pass@tcp(domain:port)/dbname*

* -l --listen **GATEWAY_LS**, gateway listen addr, format is *host:port*, example: *0.0.0.0:1329*

* -r --resource **GATEWAY_RESOURCE_URL**, gateway protect target, the resource server, could be a api server, format is *http://host:port*

## DOCKER

you could find docker image from [here](https://hub.docker.com/r/xuybin/go-api-gateway/)

docker deployment example:

```bash
docker run -d --restart=always -p 80:80 -e GATEWAY_CONN_STR='user:pass@tcp(mysql:3306)/auth' -e GATEWAY_RESOURCE_URL='http://api:80' --link mysql_1:mysql --link mysql_api:api --name api_gateway xuybin/go-api-gateway
```
docker swarm  deployment example:
```bash
docker stack deploy -c docker-compose.yml mystack
docker stack ps mystack
docker stack rm mystack
```

docker-compose deployment example:
```bash
docker-compose up
docker-compose ps
docker-compose rm
```

## DOWNLOAD

You could download the latest build binaries from [release page](https://github.com/xuybin/go-api-gateway/releases) !

## Swagger UI Support

The go-mysql-api support swagger.json and provide swagger.html page

Open **/gateway/docs/** to see swagger documents, the interactive documention will be helpful.

And **go-api-gateway** provide the *swagger.json* at path **/gateway/swagger/**


