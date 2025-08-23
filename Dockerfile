FROM alpine:latest AS build

RUN apk add nodejs npm go git make

COPY backend/ /build/backend/
COPY frontend/ /build/frontend/
COPY Makefile /build/
COPY .git/ /build/.git

WORKDIR /build
RUN make

FROM caddy:2.10 AS caddy
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=build /build/build/srv /srv

FROM alpine:latest AS backend
COPY --from=build /build/build /app/
ENTRYPOINT [ "/app/status" ]
