FROM node:18-alpine

ENV NODE_ENV=production
WORKDIR /app

COPY ["package.json", "package-lock.json*", "./"]

RUN npm install --production

COPY ["index.js", "./"]

EXPOSE 4000

CMD [ "node", "index.js" ]
