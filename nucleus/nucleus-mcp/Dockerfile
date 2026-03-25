FROM node:20-alpine

WORKDIR /app

COPY package.json package-lock.json* ./

RUN npm install

COPY tsconfig.json ./
COPY src/ ./src/

RUN npm run build

ENV NUCLEUS_API_URL=http://localhost:8080
ENV NUCLEUS_API_KEY=dev-nucleus-key-local
ENV NUCLEUS_MCP_LOG_LEVEL=info

CMD ["node", "dist/index.js"]
