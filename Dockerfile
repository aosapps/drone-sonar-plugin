FROM golang:1.13.4-alpine as build
RUN mkdir -p /go/src/github.com/aosapps/drone-sonar-plugin
WORKDIR /go/src/github.com/aosapps/drone-sonar-plugin 
COPY *.go ./
COPY vendor ./vendor/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o drone-sonar

FROM openjdk:8-jre-alpine

ARG SONAR_VERSION=3.3.0.1492
ARG SONAR_SCANNER_CLI=sonar-scanner-cli-${SONAR_VERSION}
ARG SONAR_SCANNER=sonar-scanner-${SONAR_VERSION}

RUN apk add --no-cache --update nodejs curl
COPY --from=build /go/src/github.com/aosapps/drone-sonar-plugin/drone-sonar /bin/
WORKDIR /bin

RUN curl https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/${SONAR_SCANNER_CLI}.zip -so /bin/${SONAR_SCANNER_CLI}.zip
RUN unzip ${SONAR_SCANNER_CLI}.zip \
    && rm ${SONAR_SCANNER_CLI}.zip \
    && apk del curl

ENV PATH $PATH:/bin/${SONAR_SCANNER}/bin

ENTRYPOINT /bin/drone-sonar
