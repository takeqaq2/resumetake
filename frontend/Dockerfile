FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --unsafe-perm
COPY . .
RUN npx vite build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup -S app && adduser -S -G app -u 10001 app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./
RUN npm ci --omit=dev --unsafe-perm && npm cache clean --force
RUN chown -R app:app /app
USER app
ENV PORT=8080
ENV HOST=0.0.0.0
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD wget -qO- http://localhost:8080/ || exit 1
CMD ["node", "build"]
