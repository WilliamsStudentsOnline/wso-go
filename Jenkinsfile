pipeline {
  agent {
    dockerfile {
      filename 'Dockerfile'
      args '-u root:sudo'
    }

  }
  stages {
    stage('Test') {
      steps {
        sh '''go get -u github.com/jstemmer/go-junit-report'''
        sh '''go test -v -race ./... 2>&1 | go-junit-report > report.xml'''
      }
      post {
        always {
          junit(testResults: 'report.xml', allowEmptyResults: true, healthScaleFactor: 1)
        }
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
