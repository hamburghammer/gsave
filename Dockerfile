FROM golang AS build
COPY ./go.* /src/
WORKDIR /src
RUN go mod download

COPY . /src

ENV CGO_ENABLED=0
RUN go build -o /gsave


FROM alpine
COPY --from=build /gsave /usr/bin/

ENTRYPOINT [ "gsave" ]
CMD [ "--quiet" ]
EXPOSE 8080