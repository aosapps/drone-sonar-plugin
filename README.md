# Harness CIE SonarCube Plugin with Quality Gateway

The plugin of Harness CIE to integrate with SonarQube (previously called Sonar), which is an open source code quality management platform and check the report results for status OK.

<img src="https://github.com/diegopereiraeng/harness-cie-sonarqube-scanner/blob/master/Sonar-CIE.png" alt="Plugin Configuration" width="400"/>

<img src="https://github.com/diegopereiraeng/harness-cie-sonarqube-scanner/blob/master/SonarResult.png" alt="Results" width="800"/>

<img src="https://github.com/diegopereiraeng/harness-cie-sonarqube-scanner/blob/master/SonarResultConsole.png" alt="Console Results" width="800"/>

Detail tutorials: [DOCS.md](DOCS.md).

### Build process
build go binary file: 
`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o harness-sonar`

build docker image
`docker build -t diegokoala/harness-cie-sonarqube-scanner .`


### Testing the docker image:
```commandline
docker run --rm \
  -e DRONE_REPO=test \
  -e PLUGIN_SOURCES=. \
  -e SONAR_HOST=http://localhost:9000 \
  -e SONAR_TOKEN=60878847cea1a31d817f0deee3daa7868c431433 \
  -e PLUGIN_SONAR_KEY=project-sonar \
  -e PLUGIN_SONAR_NAME=project-sonar \
  diegokoala/harness-cie-sonarqube-scanner
```

### Pipeline example
```yaml
- step:
    type: Plugin
    name: "Check Sonar "
    identifier: Check_Sonar
    spec:
        connectorRef: account.DockerHubDiego
        image: diegokoala/harness-cie-sonarqube-scanner:master
        privileged: false
        settings:
            sonar_host: http://34.100.11.50
            sonar_token: 60878847cea1a31d817f0deee3daa7868c431433
            sources: "."
            binaries: "."
            sonar_name: harness-cie-sonarqube-scanner
            sonar_key: harness-cie-sonarqube-scanner
- step:
    type: Run
    name: Sonar Show Results
    identifier: Sonar_Results
    spec:
        connectorRef: account.DockerHubDiego
        image: maven:3.6.3-jdk-8
        command: ls sonarResults.xml
        privileged: false
        reports:
            type: JUnit
            spec:
                paths:
                    - sonarResults.xml
```
