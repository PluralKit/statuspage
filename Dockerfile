FROM alpine:latest AS build

RUN apk add go make

COPY backend/ /build/backend/
COPY Makefile /build/

WORKDIR /build
RUN make backend


FROM node:20-alpine AS build-sveltekit
WORKDIR /app

COPY frontend/package*.json ./
RUN npm install

COPY frontend .
RUN npm run build


FROM node:20-alpine AS frontend
WORKDIR /app

COPY --from=build-sveltekit /app/package*.json ./
RUN npm install --omit=dev

COPY --from=build-sveltekit /app/build ./build
COPY --from=build-sveltekit /app/static ./static
EXPOSE 3000
ENV NODE_ENV=production
CMD ["node", "./build/index.js"]


FROM alpine:latest AS backend
COPY --from=build /build/build /app/
ENTRYPOINT [ "/app/status" ]