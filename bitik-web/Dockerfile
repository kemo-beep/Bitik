# syntax=docker/dockerfile:1
FROM node:20-bookworm-slim AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:20-bookworm-slim AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
ARG NEXT_PUBLIC_ASSET_BASE_URL=http://localhost:8080
ARG NEXT_PUBLIC_WS_BASE_URL=ws://localhost:8081
ARG NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL=http://localhost:3000
ARG NEXT_PUBLIC_FEATURE_FLAGS={}
ARG NEXT_PUBLIC_ANALYTICS_ENABLED=0
ARG NEXT_PUBLIC_SENTRY_DSN=

ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_ASSET_BASE_URL=$NEXT_PUBLIC_ASSET_BASE_URL
ENV NEXT_PUBLIC_WS_BASE_URL=$NEXT_PUBLIC_WS_BASE_URL
ENV NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL=$NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL
ENV NEXT_PUBLIC_FEATURE_FLAGS=$NEXT_PUBLIC_FEATURE_FLAGS
ENV NEXT_PUBLIC_ANALYTICS_ENABLED=$NEXT_PUBLIC_ANALYTICS_ENABLED
ENV NEXT_PUBLIC_SENTRY_DSN=$NEXT_PUBLIC_SENTRY_DSN

RUN node scripts/verify-web-build-env.mjs
RUN npm run build

FROM node:20-bookworm-slim AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
