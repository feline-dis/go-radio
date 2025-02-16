FROM golang:1.24 AS build-backend

WORKDIR /app

COPY ./build/go.radio /usr/local/bin/go.radio
COPY ./ingest /app/ingest
COPY ./public_react/dist /app/public_react/dist

# Install yt-dlp
RUN apt-get update && apt-get install -y curl
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp

# Install ffmpeg
RUN apt-get install -y ffmpeg

EXPOSE 8080

CMD ["go.radio"]


