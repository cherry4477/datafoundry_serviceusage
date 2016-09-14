FROM golang:1.6.2

EXPOSE 3000

ENV TIME_ZONE=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone

COPY . /go/src/github.com/asiainfoLDP/datafoundry_serviceusage

WORKDIR /go/src/github.com/asiainfoLDP/datafoundry_serviceusage

RUN go build

CMD ["sh", "-c", "./datafoundry_serviceusage -port=3000"]
