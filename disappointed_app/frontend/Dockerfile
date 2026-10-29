FROM node:22-alpine AS build
ARG VITE_WEB_PUSH_PUBLIC_KEY
ENV VITE_WEB_PUSH_PUBLIC_KEY=$VITE_WEB_PUSH_PUBLIC_KEY
WORKDIR /app/frontend
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/frontend/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
