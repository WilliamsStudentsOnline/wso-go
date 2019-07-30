pipeline {
  agent {
    dockerfile true
    args '-u root:sudo'
  }
  stages {
    stage('Test') {
      steps {
        sh '''go test -race ./...'''
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
