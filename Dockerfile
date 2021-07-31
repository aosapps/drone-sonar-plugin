FROM openjdk:8-jre-alpine

LABEL maintainer="Daniel Ramirez <dxas90@gmail.com>"

ARG SONAR_VERSION=4.0.0.1744
ARG SONAR_SCANNER_CLI=sonar-scanner-cli-${SONAR_VERSION}
ARG SONAR_SCANNER=sonar-scanner-${SONAR_VERSION}

# RUN apk add --no-cache --update nodejs curl
COPY sonar-scanner-plugin /bin/
COPY sonar-scanner-cli-4.0.0.1744.zip  /bin/${SONAR_SCANNER_CLI}.zip
WORKDIR /bin

# curl -fsSLO https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/${SONAR_SCANNER_CLI}.zip && \
RUN unzip -q ${SONAR_SCANNER_CLI}.zip && \
    rm -f ${SONAR_SCANNER_CLI}.zip

ENV PATH $PATH:/bin/${SONAR_SCANNER}/bin
ENV SONAR_SCANNER_OPTS -Xmx512m

ENTRYPOINT /bin/sonar-scanner-plugin
