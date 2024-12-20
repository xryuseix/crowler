FROM chromedp/headless-shell:132.0.6793.2
WORKDIR /app
RUN apt update \
    && apt install -y make curl wget \
    && apt install -y fonts-noto \
    && apt install -y fonts-noto-cjk \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/

# Install go:1.23.2
RUN wget https://dl.google.com/go/go1.23.2.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz && \
    rm go1.23.2.linux-amd64.tar.gz
ENV PATH $PATH:/usr/local/go/bin
ENV PATH $PATH:/root/go/bin

COPY ../app .

RUN go mod download
# RUN go build -o /bin/crowler

# ENTRYPOINT [ "/bin/crowler" ]
ENTRYPOINT [ "make", "run" ]