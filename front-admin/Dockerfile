# Stage 1: Build de la aplicación Vue 3 + TypeScript
FROM docker.io/library/node:22-alpine AS builder
WORKDIR /app

# Instalar dependencias aprovechando el cache de capas
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

# Copiar el código fuente y compilar bundle de producción
COPY . .
RUN npm run build

# Stage 2: Servidor Web Nginx liviano para producción (20MB)
FROM docker.io/library/nginx:1.27-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
