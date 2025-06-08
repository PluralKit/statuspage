ARG PUBLIC_API_URL="https://status.pluralkit.me"

FROM alpine:latest AS build

RUN apk add nodejs npm go git make

COPY backend/ /build/backend/
COPY frontend/ /build/frontend/
COPY Makefile /build/
COPY .git/ /build/.git

WORKDIR /build
ARG PUBLIC_API_URL
ENV PUBLIC_API_URL=$PUBLIC_API_URL
RUN make

FROM caddy:2.10 AS caddy
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=build /build/build/srv /srv

FROM alpine:latest AS backend
COPY --from=build /build/build /app/
ENTRYPOINT [ "/app/status" ]
