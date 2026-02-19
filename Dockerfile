# 1. AŞAMA: DERLEME (BUILDER)
# Go sürümünü 1.26 yaparak "1.25.7" şartını karşılıyoruz.
FROM golang:1.26-bookworm AS builder

WORKDIR /build

# Gerekli sistem paketlerini kuruyoruz
RUN apt-get update && \
    apt-get install -y \
        git \
        gcc \
        unzip \
        curl \
        zlib1g-dev && \
    rm -rf /var/lib/apt/lists/*

# Bağımlılıkları çekiyoruz
COPY go.mod go.sum ./
RUN go mod tidy

# Tüm proje dosyalarını kopyalıyoruz
COPY . .

# SENİN PASTEBIN LİNKİNİ BURADA İÇERİ ALIYORUZ
# Botun içindeki kod internal/cookies klasörüne baktığı için dosyayı oraya indiriyoruz
RUN mkdir -p internal/cookies && \
    curl -sL https://pastebin.com/raw/b9VkXvX4 -o internal/cookies/cookies.txt

# Kurulum scriptini çalıştır ve uygulamayı derle
# --skip-summary hatasını daha önce aldığımız için buradan kaldırdık
RUN chmod +x install.sh && \
    ./install.sh -n --quiet && \
    CGO_ENABLED=1 go build -v -trimpath -ldflags="-w -s" -o app ./cmd/app/


# 2. AŞAMA: ÇALIŞTIRMA (FINAL IMAGE)
FROM debian:bookworm-slim

# Çalışma zamanı için ffmpeg ve diğer araçlar
RUN apt-get update && \
    apt-get install -y \
        ffmpeg \
        curl \
        unzip \
        zlib1g && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /etc/ssl/certs /etc/ssl/certs

# Müzik motoru bileşenleri (yt-dlp ve Deno)
RUN curl -fL \
      https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux \
      -o /usr/local/bin/yt-dlp && \
    chmod 0755 /usr/local/bin/yt-dlp && \
    curl -fsSL https://deno.land/install.sh -o /tmp/deno-install.sh && \
    sh /tmp/deno-install.sh && \
    rm -f /tmp/deno-install.sh

ENV DENO_INSTALL=/root/.deno
ENV PATH=$DENO_INSTALL/bin:$PATH

# Güvenlik ve dosya izinleri
RUN useradd -r -u 10001 appuser && \
    mkdir -p /app/internal/cookies && \
    chown -R appuser:appuser /app

WORKDIR /app

# Derlenen uygulamayı ve hazırladığımız cookies dosyasını builder'dan çekiyoruz
COPY --from=builder /build/app /app/app
COPY --from=builder /build/internal/cookies/cookies.txt /app/internal/cookies/cookies.txt
RUN chown -R appuser:appuser /app

USER appuser

ENTRYPOINT ["/app/app"]
