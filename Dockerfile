FROM golang:1.16.6-alpine as build
RUN apk add --no-cache --update git
RUN mkdir -p /go/src/github.com/diegopereiraeng/harness-cie-sonarqube-scanner
WORKDIR /go/src/github.com/diegopereiraeng/harness-cie-sonarqube-scanner 
COPY *.go ./
COPY *.mod ./
COPY vendor ./vendor/

RUN go env GOCACHE 

RUN go get github.com/sirupsen/logrus
RUN go get github.com/pelletier/go-toml/cmd/tomll
RUN go get github.com/urfave/cli
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o harness-sonar

FROM openjdk:11.0.8-jre

ARG SONAR_VERSION=4.5.0.2216
ARG SONAR_SCANNER_CLI=sonar-scanner-cli-${SONAR_VERSION}
ARG SONAR_SCANNER=sonar-scanner-${SONAR_VERSION}

RUN apt-get update \
    && apt-get install -y nodejs curl \
    && apt-get clean

COPY --from=build /go/src/github.com/diegopereiraeng/harness-cie-sonarqube-scanner/harness-sonar /bin/
WORKDIR /bin

RUN curl https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/${SONAR_SCANNER_CLI}.zip -so /bin/${SONAR_SCANNER_CLI}.zip
RUN unzip ${SONAR_SCANNER_CLI}.zip \
    && rm ${SONAR_SCANNER_CLI}.zip 

ENV PATH $PATH:/bin/${SONAR_SCANNER}/bin

ENTRYPOINT /bin/harness-sonar