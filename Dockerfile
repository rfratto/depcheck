FROM golang:1.15-alpine

WORKDIR /code
COPY . .
RUN go install .
ENTRYPOINT ["depcheck"]
