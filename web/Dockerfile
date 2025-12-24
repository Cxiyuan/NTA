FROM node:18-alpine as builder

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production --legacy-peer-deps || npm install --legacy-peer-deps

COPY . .
RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]