# Frontend Dockerfile - Works with root context (both local and AWS)
FROM node:20-slim AS builder

WORKDIR /app

# Copy package files (build context is root, so use frontend/package.json)
COPY frontend/package*.json ./

# Install dependencies
RUN rm -f package-lock.json && npm install

# Copy frontend source code
COPY frontend/ ./

# Build arguments for environment variables
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

# Build the application
RUN npm run build

# Production stage with Nginx
FROM nginx:alpine

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
