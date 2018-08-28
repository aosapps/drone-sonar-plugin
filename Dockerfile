FROM openjdk:8-jre-alpine

COPY drone-sonar /bin/
COPY lib/sonar-scanner-cli-3.2.0.1227.zip /bin

WORKDIR /bin

RUN unzip sonar-scanner-cli-3.2.0.1227.zip \
    && rm sonar-scanner-cli-3.2.0.1227.zip

ENV PATH $PATH:/bin/sonar-scanner-3.2.0.1227/bin

ENTRYPOINT /bin/drone-sonar
