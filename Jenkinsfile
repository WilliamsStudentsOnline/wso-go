import groovy.json.JsonSlurper

pipeline {
  agent none
  environment {
    CGO_ENABLED = 1
    WSO_GO_DISCORD_WEBHOOK_URL = credentials('WSO_GO_DISCORD_WEBHOOK_URL')
  }
  stages {
    stage('Deploy for development') {
      when {
        beforeAgent true
        branch 'master'
      }
      agent {
        dockerfile {
          filename 'Dockerfile.builder'
          args '-u root:sudo'
        }
      }
      steps {
        // to fix the error fatal: unsafe repository ('<...> is owned by someone else)
        sh '''git config --global --add safe.directory "*"'''
        sh '''make build-prod-linux'''
        sh '''make build-jobs-prod-linux'''
        script {
          def remote_dev = [:]
          remote_dev.name = "wsodev"
          remote_dev.host = "wso-dev.williams.edu"
          remote_dev.port = 22
          remote_dev.allowAnyHosts = true

          withCredentials([usernamePassword(credentialsId: 'wsodev_ssh_server', passwordVariable: 'SSH_PASS', usernameVariable: 'SSH_USER')]) {
            remote_dev.user = SSH_USER
            remote_dev.password = SSH_PASS

            // Main executable:
            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/wso-backend'
            sshPut remote: remote_dev, from: 'wso-backend_linux', into: '/home/wsodev/wso-go/wso-backend'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/wso-backend'

            // Jobs:
            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-catalog-update'
            sshPut remote: remote_dev, from: 'job-catalog-update_linux', into: '/home/wsodev/wso-go/job-catalog-update'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-catalog-update'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-update-all-factrak-survey-deficits'
            sshPut remote: remote_dev, from: 'job-update-all-factrak-survey-deficits_linux', into: '/home/wsodev/wso-go/job-update-all-factrak-survey-deficits'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-update-all-factrak-survey-deficits'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-update-all-users-from-ldap'
            sshPut remote: remote_dev, from: 'job-update-all-users-from-ldap_linux', into: '/home/wsodev/wso-go/job-update-all-users-from-ldap'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-update-all-users-from-ldap'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-dorms-update'
            sshPut remote: remote_dev, from: 'job-dorms-update_linux', into: '/home/wsodev/wso-go/job-dorms-update'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-dorms-update'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-frosh-photos'
            sshPut remote: remote_dev, from: 'job-frosh-photos_linux', into: '/home/wsodev/wso-go/job-frosh-photos'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-frosh-photos'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/user-csv-data'
            sshPut remote: remote_dev, from: 'job-user-csv-data_linux', into: '/home/wsodev/wso-go/job-user-csv-data'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-user-csv-data'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/dining-update'
            sshPut remote: remote_dev, from: 'job-dining-update_linux', into: '/home/wsodev/wso-go/job-dining-update'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-dining-update'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/schedule-notifs'
            sshPut remote: remote_dev, from: 'job-schedule-notifs_linux', into: '/home/wsodev/wso-go/job-schedule-notifs'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-schedule-notifs'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-update-on-campus-semesters'
            sshPut remote: remote_dev, from: 'job-update-on-campus-semesters_linux', into: '/home/wsodev/wso-go/job-update-on-campus-semesters'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-update-on-campus-semesters'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-initialize-on-campus-semesters'
            sshPut remote: remote_dev, from: 'job-initialize-on-campus-semesters_linux', into: '/home/wsodev/wso-go/job-initialize-on-campus-semesters'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-initialize-on-campus-semesters'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-update_profs_areas_of_study'
            sshPut remote: remote_dev, from: 'job-update_profs_areas_of_study_linux', into: '/home/wsodev/wso-go/job-update_profs_areas_of_study'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-update_profs_areas_of_study'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-ephmatch-reset'
            sshPut remote: remote_dev, from: 'job-ephmatch-reset_linux', into: '/home/wsodev/wso-go/job-ephmatch-reset'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-ephmatch-reset'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/jobs/library-hours-update'
            sshPut remote: remote_dev, from: 'job-library-hours-update_linux', into: '/home/wsodev/wso-go/jobs/library-hours-update'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/jobs/library-hours-update'

            // Restart WSO-Go
            sshCommand remote: remote_dev, command: '/bin/systemctl restart WSO-Go', sudo: true
          }
        }
        script {
          try {
            URL apiUrl = new URL("https://wso-dev.williams.edu/api/v2/health-check")
            def resp = new JsonSlurper().parseText(apiUrl.getText())
            return resp.ok
          } catch (Exception e) {
            return false
          }
        }
      }
      post {
        success {
          slackSend (color: '#00FF00', message: "WSO-Go Deployed on Development\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
          discordSend (title: "WSO-Go Deployed on Development", description: "Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})", link: env.BUILD_URL, result: currentBuild.currentResult, webhookURL: "${env.WSO_GO_DISCORD_WEBHOOK_URL}")
        }
        failure {
          slackSend (color: 'danger', message: "WSO-Go Failed to Deploy on Development\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
          discordSend (title: "WSO-Go Fail to Deploy on Development", description: "Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})", link: env.BUILD_URL, result: currentBuild.currentResult, webhookURL: "${env.WSO_GO_DISCORD_WEBHOOK_URL}")
        }
      }
    }
    stage('Deploy for production') {
          when {
            beforeAgent true
            branch 'production'
          }
          agent {
            dockerfile {
              filename 'Dockerfile.builder'
              args '-u root:sudo'
            }
          }
          steps {
            sh '''git config --global --add safe.directory "*"'''
            sh '''make build-prod-linux'''
            sh '''make build-jobs-prod-linux'''
            script {
              def remote_dev = [:]
              remote_dev.name = "wso"
              remote_dev.host = "wso.williams.edu"
              remote_dev.port = 22
              remote_dev.allowAnyHosts = true

              withCredentials([usernamePassword(credentialsId: 'wso_ssh_server', passwordVariable: 'SSH_PASS', usernameVariable: 'SSH_USER')]) {
                remote_dev.user = SSH_USER
                remote_dev.password = SSH_PASS

                // Main executable:
                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/wso-backend'
                sshPut remote: remote_dev, from: 'wso-backend_linux', into: '/home/wso/wso/wso-backend/wso-backend'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/wso-backend'

                // Jobs:
                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/catalog-update'
                sshPut remote: remote_dev, from: 'job-catalog-update_linux', into: '/home/wso/wso/wso-backend/jobs/catalog-update'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/catalog-update'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/update-all-factrak-survey-deficits'
                sshPut remote: remote_dev, from: 'job-update-all-factrak-survey-deficits_linux', into: '/home/wso/wso/wso-backend/jobs/update-all-factrak-survey-deficits'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/update-all-factrak-survey-deficits'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/update-all-users-from-ldap'
                sshPut remote: remote_dev, from: 'job-update-all-users-from-ldap_linux', into: '/home/wso/wso/wso-backend/jobs/update-all-users-from-ldap'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/update-all-users-from-ldap'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/dorms-update'
                sshPut remote: remote_dev, from: 'job-dorms-update_linux', into: '/home/wso/wso/wso-backend/jobs/dorms-update'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/dorms-update'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/frosh-photos'
                sshPut remote: remote_dev, from: 'job-frosh-photos_linux', into: '/home/wso/wso/wso-backend/jobs/frosh-photos'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/frosh-photos'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/user-csv-data'
                sshPut remote: remote_dev, from: 'job-user-csv-data_linux', into: '/home/wso/wso/wso-backend/jobs/user-csv-data'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/user-csv-data'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/dining-update'
                sshPut remote: remote_dev, from: 'job-dining-update_linux', into: '/home/wso/wso/wso-backend/jobs/dining-update'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/dining-update'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/schedule-notifs'
                sshPut remote: remote_dev, from: 'job-schedule-notifs_linux', into: '/home/wso/wso/wso-backend/jobs/schedule-notifs'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/schedule-notifs'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/update-on-campus-semesters'
                sshPut remote: remote_dev, from: 'job-update-on-campus-semesters_linux', into: '/home/wso/wso/wso-backend/jobs/update-on-campus-semesters'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/update-on-campus-semesters'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/initialize-on-campus-semesters'
                sshPut remote: remote_dev, from: 'job-initialize-on-campus-semesters_linux', into: '/home/wso/wso/wso-backend/jobs/initialize-on-campus-semesters'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/initialize-on-campus-semesters'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/job-update_profs_areas_of_study'
                sshPut remote: remote_dev, from: 'job-update_profs_areas_of_study_linux', into: '/home/wso/wso/wso-backend/jobs/job-update_profs_areas_of_study'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/job-update_profs_areas_of_study'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/ephmatch-reset'
                sshPut remote: remote_dev, from: 'job-ephmatch-reset_linux', into: '/home/wso/wso/wso-backend/jobs/ephmatch-reset'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/ephmatch-reset'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/update_ephmatch_dates'
                sshPut remote: remote_dev, from: 'job-ephmatch_update_dates_linux', into: '/home/wso/wso/wso-backend/jobs/ephmatch_update_dates'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/ephmatch_update_dates'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/library-hours-update'
                sshPut remote: remote_dev, from: 'job-library-hours-update_linux', into: '/home/wso/wso/wso-backend/jobs/library-hours-update'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/library-hours-update'

                // Restart WSO-Go
                sshCommand remote: remote_dev, command: '/bin/systemctl restart WSO-Go', sudo: true
              }
            }
            script {
              try {
                URL apiUrl = new URL("https://wso.williams.edu/api/v2/health-check")
                def resp = new JsonSlurper().parseText(apiUrl.getText())
                return resp.ok
              } catch (Exception e) {
                return false
              }
            }
          }
          post {
            success {
              slackSend (color: '#00FF00', message: "WSO-Go Deployed on Production\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
              discordSend (title: "WSO-Go Deployed on Production", description: "Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})", link: env.BUILD_URL, result: currentBuild.currentResult, webhookURL: "${env.WSO_GO_DISCORD_WEBHOOK_URL}")
            }
            failure {
              slackSend (color: 'danger', message: "WSO-Go Failed to Deploy on Production\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
              discordSend (title: "WSO-Go Failed to Deploy on Production", description: "Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})", link: env.BUILD_URL, result: currentBuild.currentResult, webhookURL: "${env.WSO_GO_DISCORD_WEBHOOK_URL}")
            }
            // cleanup {
            //   // Run the cleaning-up in built-in node, only during production builds (to clean docker cache)
            //   node('master || built-in') {
            //     sh 'docker system prune -f'
            //   }
            // }
          }
        }
  }
  options { buildDiscarder(logRotator(numToKeepStr: '2')) }
  post {
    failure {
      slackSend (color: 'warning', message: "WSO-Go Failed tests\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
    }
  }
}
