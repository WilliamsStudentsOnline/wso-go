pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
       sh '''#!/bin/bash -l
GOOS=linux go build -a -tags=jsoniter -o wso-go main.go'''
      }
    }
    stage('Test') {
      steps {
        sh '''#!/bin/bash -l
GOCACHE=cache GOOS=linux go test -race ./...'''
      }
    }
  }
  options { buildDiscarder(logRotator(numToKeepStr: '2')) }
  post {
    cleanup {
      cleanWs()
    }
  }
}
