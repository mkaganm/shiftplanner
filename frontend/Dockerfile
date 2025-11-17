# Frontend Dockerfile - Production build with Nginx
FROM node:20-slim AS builder

WORKDIR /app

COPY frontend/package*.json ./
RUN rm -f package-lock.json && \
    npm install && \
    npm cache clean --force

COPY frontend/ ./

ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

RUN npm run build

FROM nginx:alpine

RUN apk --no-cache add wget

COPY --from=builder /app/dist /usr/share/nginx/html

RUN echo 'server { \
    listen 3000; \
    listen [::]:3000; \
    server_name _; \
    root /usr/share/nginx/html; \
    index index.html; \
    location / { \
        try_files $uri $uri/ /index.html; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:3000/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
