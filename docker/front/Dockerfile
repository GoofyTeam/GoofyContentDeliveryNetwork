FROM node:20-alpine AS builder

WORKDIR /app

COPY app/front/package*.json ./

RUN npm ci

COPY app/front/ ./

RUN npm run build

FROM nginx:alpine

COPY docker/front/nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]