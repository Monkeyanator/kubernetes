FROM ubuntu
RUN apt-get update
RUN apt-get install -y ca-certificates
ADD ./main /usr/bin/main
ADD ./creds.json /usr/bin/creds.json
EXPOSE 6060
ENV GOOGLE_APPLICATION_CREDENTIALS /usr/bin/creds.json
ENTRYPOINT ["/usr/bin/main", "-logtostderr"]
