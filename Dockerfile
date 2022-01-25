FROM golang:1.17-alpine

WORKDIR /code
COPY . .
RUN go install .
ENTRYPOINT ["depcheck"]
