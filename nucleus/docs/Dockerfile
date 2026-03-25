FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm install --frozen-lockfile || npm install

# Copy source files
COPY . .

# Build the static site
RUN npm run build

# ---

FROM node:20-alpine AS runtime

WORKDIR /app

RUN npm install -g serve

# Copy built site
COPY --from=builder /app/build ./build

EXPOSE 3001

CMD ["serve", "-s", "build", "-l", "3001"]
