FROM golang:1.18-alpine

WORKDIR /code
COPY . .
RUN go install .
ENTRYPOINT ["depcheck"]
