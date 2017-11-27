pipeline {
    agent none
    stages {
        stage('cluster-manager-Build') {
            agent {
                docker {
                    image 'seldonio/core-builder'
                    args '-v /root/.m2:/root/.m2'
                }
            }
            steps {
                sh 'cd cluster-manager && make -f Makefile.ci clean build_jar'
            }
        }
    }
}

