FROM golang:1.22.4 as dev

RUN apt-get update && apt-get install -y \
    iputils-ping \
    curl

RUN mkdir -p /app
WORKDIR /app

EXPOSE 5000

HEALTHCHECK --interval=20s --timeout=1m --start-period=20s \
   CMD curl -f --connect-timeout 5 --max-time 10 --retry 5 --retry-delay 0 --retry-max-time 40 --retry-all-errors 'http://localhost:5000/health' || exit 1

ENTRYPOINT ["go", "run", "."]