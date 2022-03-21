FROM gcr.io/distroless/static

RUN apt-get install -y bluez bluetooth

ENTRYPOINT ["/parasite-scanner"]
COPY parasite-scanner /