FROM openjdk:8-jre-alpine
RUN apk add --no-cache --update nodejs
COPY drone-sonar /bin/
COPY lib/sonar-scanner-cli-3.3.0.1492.zip /bin

WORKDIR /bin
RUN unzip sonar-scanner-cli-3.3.0.1492.zip \
  && mv sonar-scanner-3.3.0.1492 sonar-scanner \
  && rm -rf sonar-scanner-cli-3.3.0.1492.zip

ENV PATH $PATH:/bin/sonar-scanner/bin

ENTRYPOINT /bin/drone-sonar
