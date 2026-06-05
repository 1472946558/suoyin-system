FROM node:20-alpine AS builder

WORKDIR /src
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM alpine:3.20

RUN adduser -D -H appuser && \
    mkdir -p /srv/admin && \
    chown -R appuser:appuser /srv/admin
COPY --from=builder /src/dist /srv/admin
USER appuser
EXPOSE 80
CMD ["busybox", "httpd", "-f", "-p", "80", "-h", "/srv/admin"]
