FROM golang:1.14-alpine

WORKDIR /app

# download modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# copy source
COPY *.go ./
ADD src src
ADD static static
ADD templates templates

RUN go build -o /docker-adyen-golang-payments

EXPOSE 8080

CMD [ "/docker-adyen-golang-payments" ]