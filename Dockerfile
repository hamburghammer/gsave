FROM golang AS build
COPY ./go.* /src/
WORKDIR /src

RUN go mod download

COPY . /src
RUN go build -o /gsave


FROM debian:buster
COPY --from=build /gsave /

ENTRYPOINT [ "/gsave" ]
CMD [ "--quiet" ]
EXPOSE 8080