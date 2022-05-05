FROM ishankhare07/kubectl:arm64-1.23

ARG VERSION=0.11.2
ARG OS=Linux
ARG ARCH=x86_64

RUN curl -L https://github.com/google/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz | tar xzf - ko
RUN chmod +x ./ko
RUN cp ./ko /usr/bin

