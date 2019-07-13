FROM alicek106/hadoop:2.6.0-k8s-base
RUN apt update && apt install dnsutils -y && \
      rm -rf /var/lib/{apt,dpkg,cache,log}
ADD entrypoint.sh /
RUN chmod +x entrypoint.sh
ENTRYPOINT ./entrypoint.sh
