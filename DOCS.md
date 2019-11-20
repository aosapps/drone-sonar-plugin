---
date: 2019-02-12T10:50:00+00:00
title: SonarQube
author: aosapps
tags: [ Sonar, SonarQube, Analysis, report ]
logo: sonarqube.svg
repo: aosapps/drone-sonar-plugin
image: aosapps/drone-sonar-plugin
---

This plugin can scan your code quality and post the analysis report to your SonarQube server. SonarQube (previously called Sonar), is an open source code quality management platform.

The below pipeline configuration demonstrates simple usage:

```yaml
steps
- name: code-analysis
  image: aosapps/drone-sonar-plugin
  settings:
      sonar_host:
        from_secret: sonar_host
      sonar_token:
        from_secret: sonar_token
```

Customized parameters could be specified:

```diff
  steps
  - name: code-analysis
    image: aosapps/drone-sonar-plugin
    settings:
        sonar_host:
          from_secret: sonar_host
        sonar_token:
          from_secret: sonar_token
+       ver: 1.0
+       timeout: 20
+       sources: .
+       level: DEBUG
+       showProfiling: true
+       exclusions: **/static/**/*,**/dist/**/*.js
```

# Secret Reference

Safety first, the host and token are stored in Drone Secrets.
* `sonar_host`: Host of SonarQube with schema(http/https).
* `sonar_token`: User token used to post the analysis report to SonarQube Server. Click User -- My Account -- Security -- Generate Tokens.


# Parameter Reference

* `ver`: Code version, Default value `DRONE_BUILD_NUMBER`.
* `timeout`: Default seconds `60`.
* `sources`: Comma-separated paths to directories containing source files. 
* `inclusions`: Comma-delimited list of file path patterns to be included in analysis. When set, only files matching the paths set here will be included in analysis.
* `exclusions`: Comma-delimited list of file path patterns to be excluded from analysis. Example: `**/static/**/*,**/dist/**/*.js`.
* `level`: Control the quantity / level of logs produced during an analysis. Default value `INFO`. 
    * DEBUG: Display INFO logs + more details at DEBUG level.
    * TRACE: Display DEBUG logs + the timings of all ElasticSearch queries and Web API calls executed by the SonarQube Scanner.
* `showProfiling`: Display logs to see where the analyzer spends time. Default value `false`
* `branchAnalysis`: Pass currently analysed branch to SonarQube. (Must not be active for initial scan!) Default value `false`

# Notes

* projectKey: `DRONE_REPO`
* projectName: `DRONE_REPO`
* You could also add a file named `sonar-project.properties` at the root of your project to specify parameters.

Code repository: [aosapps/drone-sonar-plugin](https://github.com/aosapps/drone-sonar-plugin).  
SonarQube Parameters: [Analysis Parameters](https://docs.sonarqube.org/display/SONAR/Analysis+Parameters)

# Test your SonarQube Server:

Replace the parameter values with your own：

```commandline
sonar-scanner \
  -Dsonar.projectKey=Neptune:news \
  -Dsonar.sources=. \
  -Dsonar.projectName=Neptune/news \
  -Dsonar.projectVersion=1.0 \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.login=60878847cea1a31d817f0deee3daa7868c431433
```
