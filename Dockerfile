FROM alpine

ENV GATEWAY_LS :80
RUN apk add --update curl && \
    tag=`curl -s -L https://api.github.com/repos/xuybin/go-api-gateway/releases/latest |awk -F "[tag_name]" '/tag_name/{print$0}' | sed  's/.*"\(v[0-9.]*\)".*/\1/'` && \
    curl  -L https://github.com/xuybin/go-api-gateway/releases/download/${tag}/go-api-gateway-linux-amd64 > /go-api-gateway  && \
    chmod +x /go-api-gateway && \
    apk del curl && \
    rm -rf /var/cache/apk/*
COPY docs /docs
EXPOSE 80

CMD ["/go-api-gateway"]