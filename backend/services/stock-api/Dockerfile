FROM node:22-alpine AS builder
WORKDIR /app

# Primero solo las dependencias: esta capa se cachea mientras no cambie package.json
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

# Después el código
COPY tsconfig.json ./
COPY src/ src/
RUN npm run build

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY package.json package-lock.json ./
RUN npm ci --omit=dev --no-audit --no-fund
COPY --from=builder /app/dist dist/
CMD ["node", "dist/index.js"]
