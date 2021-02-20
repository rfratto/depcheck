FROM golang:1.16-alpine

WORKDIR /code
COPY . .
RUN go install .
ENTRYPOINT ["depcheck"]
