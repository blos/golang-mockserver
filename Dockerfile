FROM golang:bookworm

WORKDIR server/

# Should be revisited in the future
COPY . .

RUN ["make", "build"]

ENTRYPOINT ["./mockserver"]

