pipeline {
  agent {
    dockerfile {
      filename 'Dockerfile'
      args '-u root:sudo --rm'
    }

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
