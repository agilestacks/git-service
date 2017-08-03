FROM alpine

ENV GIT_API_SECRET ""
ENV BUCKET         "terraform.agilestacks.com"

COPY bin/linux/gits /gits
ADD  https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant /gits-key

VOLUME /git

RUN chmod +x /gits && \
    chmod 600 /gits-key && \
    apk add --update git openssh && \
    rm /var/cache/apk/*

EXPOSE 2022
EXPOSE 8005

ENTRYPOINT "/gits"
CMD ["-blobsFrom", "s3://$BUCKET/"]