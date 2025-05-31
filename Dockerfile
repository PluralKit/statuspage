FROM alpine:latest AS build

RUN apk add go git make

COPY backend/ /build/backend/
COPY Makefile /build/
COPY .git/ /build/.git

WORKDIR /build

RUN make backend

FROM alpine:latest

COPY --from=build /build/build /app/
ENTRYPOINT [ "/app/status" ]