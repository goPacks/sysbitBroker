FROM golang:1.22-alpine AS build

WORKDIR /src/

# COPY main.go go.* ./auth/auth.go ./data/data.go /src/
# COPY ./auth/auth.go  /src/auth/auth.go
# COPY ./data/data.go  /src/data/data.go

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/demo

FROM scratch
COPY --from=build /bin/demo /bin/demo
ENTRYPOINT ["/bin/demo"]
