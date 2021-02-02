FROM alpine

CMD ["/bin/baikal"]

ADD home-assistant-fix.patch /tmp/

WORKDIR /baikal

# libcap for "$ setcap"
RUN apk add --update libcap patch php7-cgi php7-sqlite3 php7-dom php7-mbstring php7-session \
	php7-openssl php-pdo php7-pdo php7-pdo_sqlite php7-json php7-xmlreader php7-xmlwriter \
	&& cd /tmp \
	&& wget https://github.com/sabre-io/Baikal/releases/download/0.8.0/baikal-0.8.0.zip \
	&& unzip *.zip \
	&& rm *.zip \
	&& mv baikal / \
	&& mv /baikal/Specific /baikal/data-template \
	&& rm -rf /baikal/config \
	&& ln -s /data /baikal/config \
	&& ln -s /data /baikal/Specific

# -DH = don't create password or home directory
RUN adduser -DH nonroot < /dev/null

# NOTE: working patch syntax depend on if we're using Ubuntu/Alpine..
RUN cd /baikal && patch -R vendor/sabre/dav/lib/CalDAV/Plugin.php < /tmp/home-assistant-fix.patch

ADD rel/baikal_linux-amd64 /bin/baikal

# https://medium.com/elbstack/docker-go-and-privileged-ports-d6354db472c3
RUN setcap 'cap_net_bind_service=+ep' /bin/baikal

USER nonroot
