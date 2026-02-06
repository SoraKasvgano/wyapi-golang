FROM debian:bookworm-slim

ARG BIN_FILE=wyapi-golang_linux_amd64

RUN if [ -f /etc/apt/sources.list ]; then \
        sed -i 's|https://deb.debian.org/debian|http://mirrors.tuna.tsinghua.edu.cn/debian|g' /etc/apt/sources.list; \
        sed -i 's|http://deb.debian.org/debian|http://mirrors.tuna.tsinghua.edu.cn/debian|g' /etc/apt/sources.list; \
        sed -i 's|https://security.debian.org/debian-security|http://mirrors.tuna.tsinghua.edu.cn/debian-security|g' /etc/apt/sources.list; \
        sed -i 's|http://security.debian.org/debian-security|http://mirrors.tuna.tsinghua.edu.cn/debian-security|g' /etc/apt/sources.list; \
    elif [ -f /etc/apt/sources.list.d/debian.sources ]; then \
        sed -i 's|https://deb.debian.org/debian|http://mirrors.tuna.tsinghua.edu.cn/debian|g' /etc/apt/sources.list.d/debian.sources; \
        sed -i 's|http://deb.debian.org/debian|http://mirrors.tuna.tsinghua.edu.cn/debian|g' /etc/apt/sources.list.d/debian.sources; \
        sed -i 's|https://security.debian.org/debian-security|http://mirrors.tuna.tsinghua.edu.cn/debian-security|g' /etc/apt/sources.list.d/debian.sources; \
        sed -i 's|http://security.debian.org/debian-security|http://mirrors.tuna.tsinghua.edu.cn/debian-security|g' /etc/apt/sources.list.d/debian.sources; \
    fi \
  && apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates tzdata \
  && rm -rf /var/lib/apt/lists/* \
  && mkdir -p /app /data

WORKDIR /data

COPY dist/${BIN_FILE} /app/wyapi-golang
RUN chmod +x /app/wyapi-golang

ENV TZ=Asia/Shanghai

EXPOSE 8000
VOLUME ["/data"]

ENTRYPOINT ["/app/wyapi-golang", "-config", "/data/config.json"]
